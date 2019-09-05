package logs

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jenkins-x/jx/pkg/cmd/opts"
	step2 "github.com/jenkins-x/jx/pkg/cmd/opts/step"
	"github.com/jenkins-x/jx/pkg/cmd/step"
	"github.com/jenkins-x/jx/pkg/util"

	"github.com/jenkins-x/jx/pkg/cloud/gke"
	"github.com/jenkins-x/jx/pkg/tekton"

	"github.com/fatih/color"
	v1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/pkg/builds"
	"github.com/jenkins-x/jx/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx/pkg/kube"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/pkg/errors"
	v1alpha12 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TektonLogger contains the necessary clients and the namespace to get data from the cluster, an implementation of
// LogWriter to write logs to and a logs retriever function to override the default way to obtain logs
type TektonLogger struct {
	JXClient          versioned.Interface
	TektonClient      tektonclient.Interface
	KubeClient        kubernetes.Interface
	LogWriter         LogWriter
	Namespace         string
	LogsRetrieverFunc retrieverFunc
	logsChannel       chan LogLine
	errorsChannel     chan error
	wg                *sync.WaitGroup
}

// LogWriter is an interface that can be implemented to define different ways to stream / write logs
// it's the implementer's responsibility to route those logs through the corresponding medium
type LogWriter interface {
	WriteLog(line LogLine, lch chan<- LogLine) error
	StreamLog(lch <-chan LogLine, ech <-chan error) error
	BytesLimit() int
}

// retrieverFunc is a func signature used to define the LogsRetrieverFunc in TektonLogger
type retrieverFunc func(pod *corev1.Pod, container *corev1.Container) (io.Reader, func(), error)

// LogLine is the object sent to and received from the channels in the StreamLog and WriteLog functions
// defined by LogWriter
type LogLine struct {
	Line       string
	ShouldMask bool
}

// GetTektonPipelinesWithActivePipelineActivity returns list of all PipelineActivities with corresponding Tekton PipelineRuns ordered by the PipelineRun creation timestamp and a map to obtain its reference once a name has been selected
func (t TektonLogger) GetTektonPipelinesWithActivePipelineActivity(filters []string, context string) ([]string, map[string]*v1.PipelineActivity, error) {
	labelsFilter := strings.Join(filters, ",")
	paList, err := t.JXClient.JenkinsV1().PipelineActivities(t.Namespace).List(metav1.ListOptions{
		LabelSelector: labelsFilter,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "there was a problem getting the PipelineActivities")
	}

	sort.Slice(paList.Items, func(i, j int) bool {
		return paList.Items[i].CreationTimestamp.After(paList.Items[j].CreationTimestamp.Time)
	})

	paMap := make(map[string]*v1.PipelineActivity)
	for _, pa := range paList.Items {
		p := pa
		paName := createPipelineActivityName(p.Labels, p.Spec.Build)
		log.Logger().Warnf("activity name: %s", paName)
		paMap[paName] = &p
	}

	tektonPRs, _ := t.TektonClient.TektonV1alpha1().PipelineRuns(t.Namespace).List(metav1.ListOptions{
		LabelSelector: labelsFilter,
	})

	// Handle the old "repo" tag as well for legacy purposes.
	if len(tektonPRs.Items) == 0 {
		labelsFilter = strings.Replace(labelsFilter, "repository=", "repo=", 1)
		tektonPRs, _ = t.TektonClient.TektonV1alpha1().PipelineRuns(t.Namespace).List(metav1.ListOptions{
			LabelSelector: labelsFilter,
		})
	}

	prMap := make(map[string][]*v1alpha12.PipelineRun)
	for _, pr := range tektonPRs.Items {
		p := pr
		prBuildNumber := p.Labels[v1.LabelBuild]
		if prBuildNumber == "" {
			prBuildNumber = findLegacyPipelineRunBuildNumber(&p)
		}
		paName := createPipelineActivityName(p.Labels, prBuildNumber)
		if _, exists := prMap[paName]; !exists {
			prMap[paName] = []*v1alpha12.PipelineRun{}
		}
		prMap[paName] = append(prMap[paName], &p)
	}

	var names []string
	for _, pa := range paList.Items {
		paName := createPipelineActivityName(pa.Labels, pa.Spec.Build)
		if _, exists := prMap[paName]; exists {
			hasNonPendingPR := false
			for _, pr := range prMap[paName] {
				if tekton.PipelineRunIsNotPending(pr) {
					hasNonPendingPR = true
				}
			}
			if hasNonPendingPR {
				names = append(names, paName)
			}
		} else if pa.Spec.CompletedTimestamp != nil {
			names = append(names, paName)
		}
	}

	return names, paMap, nil
}

func modifyFilterForPipelineRun(labelsFilter string, context string) string {
	contextFilter := fmt.Sprintf("context=%s", context)
	if labelsFilter == "" {
		return contextFilter
	}
	return fmt.Sprintf("%s,%s", labelsFilter, contextFilter)
}

func createPipelineActivityName(labels map[string]string, buildNumber string) string {
	repository := labels[v1.LabelRepository]
	// The label is called "repo" in the PipelineRun CRD and "repository" in the PipelineActivity CRD
	if repository == "" {
		repository = labels["repo"]
	}
	baseName := strings.ToLower(fmt.Sprintf("%s/%s/%s #%s", labels[v1.LabelOwner], repository, labels[v1.LabelBranch], buildNumber))

	context := labels[v1.LabelContext]
	if context != "" {
		return strings.ToLower(fmt.Sprintf("%s %s", baseName, context))
	}

	return baseName
}

func findLegacyPipelineRunBuildNumber(pipelineRun *v1alpha12.PipelineRun) string {
	var buildNumber string
	for _, p := range pipelineRun.Spec.Params {
		if p.Name == "build_id" {
			buildNumber = p.Value
		}
	}
	return buildNumber
}

func getPipelineRunNamesForActivity(pa *v1.PipelineActivity, tektonClient tektonclient.Interface) ([]string, error) {
	filters := []string{
		fmt.Sprintf("%s=%s", v1.LabelOwner, pa.Spec.GitOwner),
		fmt.Sprintf("%s=%s", v1.LabelRepository, pa.Spec.GitRepository),
		fmt.Sprintf("%s=%s", v1.LabelBranch, pa.Spec.GitBranch),
	}

	tektonPRs, err := tektonClient.TektonV1alpha1().PipelineRuns(pa.Namespace).List(metav1.ListOptions{
		LabelSelector: strings.Join(filters, ","),
	})
	if err != nil {
		return nil, err
	}
	// For legacy purposes, look for the old "repo" label as well.
	if len(tektonPRs.Items) == 0 {
		tektonPRs, err = tektonClient.TektonV1alpha1().PipelineRuns(pa.Namespace).List(metav1.ListOptions{
			LabelSelector: strings.Replace(strings.Join(filters, ","), "repository=", "repo=", 1),
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(tektonPRs.Items, func(i, j int) bool {
		return tektonPRs.Items[i].CreationTimestamp.Before(&tektonPRs.Items[j].CreationTimestamp)
	})

	var names []string
	for _, pr := range tektonPRs.Items {
		buildNumber := pr.Labels[tekton.LabelBuild]
		if buildNumber == "" {
			buildNumber = findLegacyPipelineRunBuildNumber(&pr)
		}
		if buildNumber == pa.Spec.Build {
			names = append(names, pr.Name)
		}
	}

	return names, nil
}

// GetRunningBuildLogs obtains the logs of the provided PipelineActivity and streams the running build pods' logs using the provided LogWriter
func (t TektonLogger) GetRunningBuildLogs(pa *v1.PipelineActivity, buildName string) error {
	pipelineRunNames, err := getPipelineRunNamesForActivity(pa, t.TektonClient)
	if err != nil {
		return errors.Wrapf(err, "failed to get PipelineRun names for activity %s in namespace %s", pa.Name, pa.Namespace)
	}
	pipelineRunsLogged := make(map[string]bool)
	foundLogs := false

	for len(pipelineRunNames) > len(pipelineRunsLogged) {
		for _, prName := range pipelineRunNames {
			_, runSeen := pipelineRunsLogged[prName]
			if !runSeen {
				structure, err := t.JXClient.JenkinsV1().PipelineStructures(pa.Namespace).Get(prName, metav1.GetOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to get pipeline structure for %s in namespace %s", prName, pa.Namespace)
				}

				allStages := structure.GetAllStagesWithSteps()
				stagesSeen := make(map[string]bool)

				// Repeat until we've seen pods for all stages
				for len(allStages) > len(stagesSeen) {
					pods, err := builds.GetPipelineRunPods(t.KubeClient, pa.Namespace, prName)
					if err != nil {
						return errors.Wrapf(err, "failed to get pods for pipeline run %s in namespace %s", prName, pa.Namespace)
					}

					sort.Slice(pods, func(i, j int) bool {
						return pods[i].CreationTimestamp.Before(&pods[j].CreationTimestamp)
					})

					for _, pod := range pods {
						stageName := pod.Labels["jenkins.io/task-stage-name"]
						params := builds.CreateBuildPodInfo(pod)
						if _, seen := stagesSeen[stageName]; !seen && params.Organisation == pa.Spec.GitOwner && params.Repository == pa.Spec.GitRepository &&
							strings.ToLower(params.Branch) == strings.ToLower(pa.Spec.GitBranch) && params.Build == pa.Spec.Build {
							stagesSeen[stageName] = true
							pipelineRunsLogged[prName] = true
							foundLogs = true
							err := t.getContainerLogsFromPod(pod, pa, buildName, stageName)
							if err != nil {
								return errors.Wrapf(err, "failed to obtain the logs for build %s and stage %s", buildName, stageName)
							}
						}
					}
					if !foundLogs {
						break
					}
				}
			}
		}
		if !foundLogs {
			break
		}
		pipelineRunNames, err = getPipelineRunNamesForActivity(pa, t.TektonClient)
		if err != nil {
			return errors.Wrapf(err, "failed to get PipelineRun names for activity %s in namespace %s", pa.Name, pa.Namespace)
		}
	}
	if !foundLogs {
		return errors.New("the build pods for this build have been garbage collected and the log was not found in the long term storage bucket")
	}

	return nil
}

func (t *TektonLogger) getContainerLogsFromPod(pod *corev1.Pod, pa *v1.PipelineActivity, buildName string, stageName string) error {
	infoColor := color.New(color.FgGreen)
	infoColor.EnableColor()
	errorColor := color.New(color.FgRed)
	errorColor.EnableColor()
	containers, _, _ := kube.GetContainersWithStatusAndIsInit(pod)
	t.initializeLoggingRoutine()
	for i, ic := range containers {
		pod, err := t.waitForContainerToStart(pa.Namespace, pod, i, stageName)
		err = t.LogWriter.WriteLog(LogLine{
			Line: fmt.Sprintf("\nShowing logs for build %v stage %s and container %s",
				infoColor.Sprintf(buildName), infoColor.Sprintf(stageName), infoColor.Sprintf(ic.Name)),
		}, t.logsChannel)
		if err != nil {
			return errors.Wrapf(err, "there was a problem writing a single line into the logs writer")
		}
		err = t.fetchLogsToChannel(pa.Namespace, pod, &ic)
		if err != nil {
			return errors.Wrap(err, "couldn't fetch logs into the logs channel")
		}
		if hasStepFailed(pod, i, t.KubeClient, pa.Namespace) {
			err = t.LogWriter.WriteLog(LogLine{
				Line: errorColor.Sprintf("\nPipeline failed on stage '%s' : container '%s'. The execution of the pipeline has stopped.", stageName, ic.Name),
			}, t.logsChannel)
			if err != nil {
				return err
			}
			break
		}
	}
	// We are done using the logs and errors channels, any message in the channels should still be read before calling wg.Done()
	t.closeLoggingChannels()
	// Waiting so we don't finish the main routine before waiting for all traces to be printed
	t.wg.Wait()
	return nil
}

func (t *TektonLogger) syncStreamLog() {
	err := t.LogWriter.StreamLog(t.logsChannel, t.errorsChannel)
	if err != nil {
		log.Logger().Error(err)
	}
	defer t.wg.Done()
}

func (t *TektonLogger) fetchLogsToChannel(ns string, pod *corev1.Pod, container *corev1.Container) error {

	if t.LogsRetrieverFunc == nil {
		t.LogsRetrieverFunc = t.retrieveLogsFromPod
	}

	reader, cleanFN, err := t.LogsRetrieverFunc(pod, container)
	if err != nil {
		t.errorsChannel <- err
		return err
	}
	defer cleanFN()
	err = writeStreamLines(reader, t.logsChannel)
	if err != nil {
		return err
	}
	return nil
}

func writeStreamLines(reader io.Reader, logCh chan<- LogLine) error {
	buffReader := bufio.NewReader(reader)
	if buffReader == nil {
		return errors.New("there was a problem obtaining a buffered reader")
	}
	for {
		line, _, err := buffReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return errors.Wrap(err, "failed to read stream")
		}
		logCh <- LogLine{Line: string(line), ShouldMask: true}
	}
}

func hasStepFailed(pod *corev1.Pod, stepNumber int, kubeClient kubernetes.Interface, ns string) bool {
	pod, err := kubeClient.CoreV1().Pods(ns).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger().Error("couldn't find the updated pod to check the step status")
		return false
	}
	_, containerStatus, _ := kube.GetContainersWithStatusAndIsInit(pod)
	if containerStatus[stepNumber].State.Terminated != nil && containerStatus[stepNumber].State.Terminated.ExitCode != 0 {
		return true
	}
	return false
}

func (t TektonLogger) waitForContainerToStart(ns string, pod *corev1.Pod, idx int, stageName string) (*corev1.Pod, error) {
	if pod.Status.Phase == corev1.PodFailed {
		return pod, nil
	}
	if kube.HasContainerStarted(pod, idx) {
		return pod, nil
	}
	containerName := ""
	containers, _, _ := kube.GetContainersWithStatusAndIsInit(pod)
	if idx < len(containers) {
		containerName = containers[idx].Name
	}
	// This method will be executed by both the CLI and the UI, we don't know if the UI has color enabled, so we are using a local instance instead of the global one
	c := color.New(color.FgGreen)
	c.EnableColor()
	if err := t.LogWriter.WriteLog(LogLine{
		Line: fmt.Sprintf("\nwaiting for stage %s : container %s to start...\n", c.Sprintf(stageName), c.Sprintf(containerName)),
	}, t.logsChannel); err != nil {
		log.Logger().Warn("There was a problem writing a single line into the writeFN")
	}
	for {
		time.Sleep(time.Second)
		p, err := t.KubeClient.CoreV1().Pods(ns).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return p, errors.Wrapf(err, "failed to load pod %s", pod.Name)
		}
		if kube.HasContainerStarted(p, idx) {
			return p, nil
		}
	}
}

// StreamPipelinePersistentLogs reads logs from the provided bucket URL and writes them using the provided LogWriter
func (t *TektonLogger) StreamPipelinePersistentLogs(logsURL string, o *opts.CommonOptions) error {
	//TODO: This should be changed in the future when other bucket providers are supported
	t.initializeLoggingRoutine()
	u, err := url.Parse(logsURL)
	if err != nil {
		return errors.Wrapf(err, "unable to parse logs URL %s to retrieve scheme", logsURL)
	}
	var logBytes []byte
	switch u.Scheme {
	case "gs":
		logBytes, err = gke.DownloadFileFromBucket(logsURL)
		if err != nil {
			return errors.Wrapf(err, "there was a problem obtaining the log file from the configured bucket URL %s", logsURL)
		}
	case "http":
		fallthrough
	case "https":
		logBytes, err = downloadLogFile(logsURL, o)
		if err != nil {
			return errors.Wrapf(err, "there was a problem obtaining the log file from the github pages URL %s", logsURL)
		}
	default:
		return t.writeBlockingLine(LogLine{
			Line: fmt.Sprintf("The provided logsURL scheme is not supported: %s", u.Scheme),
		})
	}

	if len(logBytes) == 0 {
		return t.writeBlockingLine(LogLine{
			Line: "The build pods for this build have been garbage collected and we couldn't find the any stored log file",
		})
	}
	return t.writeBlockingLine(LogLine{
		Line: string(logBytes),
	})
}

// create the logs and errors channels and the waitgroup for this TektonLogger instance
// assign a pointer to the waitgroup to TektonLogger which will be used by all other methods
// then start the log writing goroutine, which calls the implementation of StreamLogs of the given LogWriter
func (t *TektonLogger) initializeLoggingRoutine() {
	t.logsChannel = make(chan LogLine)
	t.errorsChannel = make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	t.wg = &wg
	go t.syncStreamLog()
}

func (t *TektonLogger) closeLoggingChannels() {
	close(t.logsChannel)
	close(t.errorsChannel)
}

// send a line to the main logs channel, close the logging channel and wait
// any message still in the channel should be read even if closed
func (t *TektonLogger) writeBlockingLine(line LogLine) error {
	err := t.LogWriter.WriteLog(line, t.logsChannel)
	if err != nil {
		return err
	}
	t.closeLoggingChannels()
	t.wg.Wait()
	return nil
}

// Uses the same signature as retrieverFunc so it can be used in TektonLogger
func (t TektonLogger) retrieveLogsFromPod(pod *corev1.Pod, container *corev1.Container) (io.Reader, func(), error) {
	options := &corev1.PodLogOptions{
		Container: container.Name,
		Follow:    true,
	}
	bytesLimit := t.LogWriter.BytesLimit()
	if bytesLimit > 0 {
		a := int64(bytesLimit)
		options.LimitBytes = &a
	}
	req := t.KubeClient.CoreV1().Pods(t.Namespace).GetLogs(pod.Name, options)
	stream, err := req.Stream()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "there was an error creating the logs stream for pod %s", pod.Name)
	}
	reader := bufio.NewReader(stream)
	return reader, func() {
		if stream != nil {
			stream.Close()
		}
	}, nil
}

func downloadLogFile(logsURL string, o *opts.CommonOptions) ([]byte, error) {
	f, _ := ioutil.TempFile("", uuid.New().String())
	defer util.DeleteFile(f.Name())
	unstashStep := step.StepUnstashOptions{
		StepOptions: step2.StepOptions{
			CommonOptions: o,
		},
		URL:    logsURL,
		OutDir: f.Name(),
	}
	err := unstashStep.Run()
	if err != nil {
		return nil, err
	}
	logBytes, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}
	return logBytes, nil
}
