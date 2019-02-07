package buildpipeline

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	pipelinev1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/knative/pkg/apis"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
)

const (
	PipelineFileName = "Jenkinsfile.yaml"
)

type Jenkinsfile struct {
	APIVersion  string      `yaml:"apiVersion"`
	Agent       Agent       `yaml:"agent,omitempty"`
	Environment []EnvVar    `yaml:"environment,omitempty"`
	Options     RootOptions `yaml:"options,omitempty"`
	Stages      []Stage     `yaml:"stages"`
	Post        []Post      `yaml:"post,omitempty"`
}

type Agent struct {
	// One of label or image is required.
	Label string `yaml:"label,omitempty"`
	Image string `yaml:"image,omitempty"`
	// Perhaps we'll eventually want to add something here for specifying a volume to create? Would play into stash.
}

type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type TimeoutUnit string

const (
	TimeoutUnitSeconds TimeoutUnit = "seconds"
	TimeoutUnitMinutes TimeoutUnit = "minutes"
	TimeoutUnitHours   TimeoutUnit = "hours"
	TimeoutUnitDays    TimeoutUnit = "days"
)

var AllTimeoutUnits = []TimeoutUnit{TimeoutUnitSeconds, TimeoutUnitMinutes, TimeoutUnitHours, TimeoutUnitDays}

func allTimeoutUnitsAsStrings() []string {
	tu := make([]string, len(AllTimeoutUnits))

	for i, u := range AllTimeoutUnits {
		tu[i] = string(u)
	}

	return tu
}

type Timeout struct {
	Time int64 `yaml:"time"`
	// Has some sane default - probably seconds
	Unit TimeoutUnit `yaml:"unit,omitempty"`
}

// TODO: Not yet implemented in build-pipeline
type RootOptions struct {
	Timeout Timeout `yaml:"timeout,omitempty"`
	Retry   int8    `yaml:"retry,omitempty"`
}

type Stash struct {
	Name string `yaml:"name"`
	// Eventually make this optional so that you can do volumes instead
	Files string `yaml:"files"`
}

type Unstash struct {
	Name string `yaml:"name"`
	Dir  string `yaml:"dir,omitempty"`
}

// TODO: Not yet implemented in build-pipeline
type StageOptions struct {
	RootOptions `yaml:",inline"`

	Stash     Stash   `yaml:"stash,omitempty"`
	Unstash   Unstash `yaml:"unstash,omitempty"`
	Workspace *string `yaml:"workspace,omitempty"`
}

type Step struct {
	// One of command or step is required.
	Command string `yaml:"command,omitempty"`
	// args is optional, but only allowed with command
	Arguments []string `yaml:"args,omitempty"`
	// dir is optional, but only allowed with command. Refers to subdirectory of workspace
	Dir string `yaml:"dir,omitempty"`

	Step string `yaml:"step,omitempty"`
	// options is optional, but only allowed with step
	// Also, we'll need to do some magic to do type verification during translation - i.e., this step wants a number
	// for this option, so translate the string value for that option to a number.
	Options map[string]string `yaml:"options,omitempty"`

	// agent can be overridden on a step
	Agent Agent `yaml:"agent,omitempty"`
}

type Stage struct {
	Name        string       `yaml:"name"`
	Agent       Agent        `yaml:"agent,omitempty"`
	Options     StageOptions `yaml:"options,omitempty"`
	Environment []EnvVar     `yaml:"environment,omitempty"`
	Steps       []Step       `yaml:"steps,omitempty"`
	Stages      []Stage      `yaml:"stages,omitempty"`
	Parallel    []Stage      `yaml:"parallel,omitempty"`
	Post        []Post       `yaml:"post,omitempty"`
}

type PostCondition string

// Probably extensible down the road
const (
	PostConditionSuccess PostCondition = "success"
	PostConditionFailure PostCondition = "failure"
	PostConditionAlways  PostCondition = "always"
)

var AllPostConditions = []PostCondition{PostConditionAlways, PostConditionSuccess, PostConditionFailure}

// TODO: Conditional execution of something after a Task or Pipeline completes is not yet implemented
type Post struct {
	Condition PostCondition `yaml:"condition"`
	Actions   []PostAction  `yaml:"actions"`
}

// TODO: Notifications are not yet supported in Build Pipeline per se.
type PostAction struct {
	Name string `yaml:"name"`
	// Also, we'll need to do some magic to do type verification during translation - i.e., this action wants a number
	// for this option, so translate the string value for that option to a number.
	Options map[string]string `yaml:"options,omitempty"`
}

var _ apis.Validatable = (*Jenkinsfile)(nil)
var _ apis.Defaultable = (*Jenkinsfile)(nil)

func (s *Stage) TaskName() string {
	return strings.ToLower(strings.NewReplacer(" ", "-").Replace(s.Name))
}

// Task/Step names need to be RFC 1035/1123 compliant DNS labels, so we mangle
// them to make them compliant. Results should match the following regex and be
// no more than 63 characters long:
//     [a-z]([-a-z0-9]*[a-z0-9])?
// cf. https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
// body is assumed to have at least one ASCII letter.
// suffix is assumed to be alphanumeric and non-empty.
func MangleToRfc1035Label(body string, suffix string) string {
	const MAX_LABEL_LENGTH = 63
	MAX_BODY_LENGTH := MAX_LABEL_LENGTH - len(suffix) - 1 // Add an extra hyphen before the suffix

	var sb strings.Builder
	bufferedHyphen := false // Used to make sure we don't output consecutive hyphens.
	for _, codepoint := range body {
		toWrite := 0
		if sb.Len() != 0 { // Digits and hyphens aren't allowed to be the first character
			if codepoint == ' ' || codepoint == '-' || codepoint == '.' {
				bufferedHyphen = true
			} else if codepoint >= '0' && codepoint <= '9' {
				toWrite = 1
			}
		}

		if codepoint >= 'A' && codepoint <= 'Z' {
			codepoint += ('a' - 'A') // Offset to make character lowercase
			toWrite = 1
		} else if codepoint >= 'a' && codepoint <= 'z' {
			toWrite = 1
		}

		if toWrite > 0 {
			if bufferedHyphen {
				toWrite += 1
			}
			if sb.Len()+toWrite > MAX_BODY_LENGTH {
				break
			}
			if bufferedHyphen {
				sb.WriteRune('-')
				bufferedHyphen = false
			}
			sb.WriteRune(codepoint)
		}
	}

	sb.WriteRune('-')
	sb.WriteString(suffix)

	return sb.String()
}

func ParseJenkinsfileYaml(jenkinsfileYaml string) (*Jenkinsfile, error) {
	jf := Jenkinsfile{}

	err := yaml.Unmarshal([]byte(jenkinsfileYaml), &jf)
	if err != nil {
		return &jf, errors.Wrapf(err, "Failed to unmarshal string %s", jenkinsfileYaml)
	}

	return &jf, nil
}

func (j *Jenkinsfile) SetDefaults() {

}

// TODO: Improve validation to actually return all the errors via the nested errors?
// TODO: Add validation for the not-yet-supported-for-CRD-generation sections
func (j *Jenkinsfile) Validate() *apis.FieldError {
	if err := validateApiVersion(j.APIVersion); err != nil {
		return err
	}

	if err := validateAgent(j.Agent).ViaField("agent"); err != nil {
		return err
	}

	if err := validateStages(j.Stages, j.Agent); err != nil {
		return err
	}

	if err := validateRootOptions(j.Options).ViaField("options"); err != nil {
		return err
	}

	return nil
}

func validateApiVersion(apiVersion string) *apis.FieldError {
	valid, err := regexp.MatchString("^v\\d+\\.\\d+", apiVersion)

	if err != nil {
		return &apis.FieldError{
			Message: fmt.Sprintf("Error validating apiVersion: %s", err),
			Paths:   []string{"apiVersion"},
		}
	}

	if !valid {
		return &apis.FieldError{
			Message: "Invalid apiVersion format: must be 'v(digits).(digits)",
			Paths:   []string{"apiVersion"},
		}
	}

	return nil
}

func validateAgent(a Agent) *apis.FieldError {
	// TODO: This is the same whether you specify an agent without label or image, or if you don't specify an agent
	// at all, which is nonoptimal.
	if !equality.Semantic.DeepEqual(a, Agent{}) {
		if a.Image != "" && a.Label != "" {
			return apis.ErrMultipleOneOf("label", "image")
		}

		if a.Image == "" && a.Label == "" {
			return apis.ErrMissingOneOf("label", "image")
		}
	}

	return nil
}

var containsAsciiLetter = regexp.MustCompile(`[a-zA-Z]`).MatchString

func validateStage(s Stage, parentAgent Agent) *apis.FieldError {
	if len(s.Steps) == 0 && len(s.Stages) == 0 && len(s.Parallel) == 0 {
		return apis.ErrMissingOneOf("steps", "stages", "parallel")
	}

	if !containsAsciiLetter(s.Name) {
		return &apis.FieldError{
			Message: "Stage name must contain at least one ASCII letter",
			Paths:   []string{"name"},
		}
	}

	stageAgent := s.Agent
	if equality.Semantic.DeepEqual(stageAgent, Agent{}) {
		stageAgent = parentAgent
	}

	if equality.Semantic.DeepEqual(stageAgent, Agent{}) {
		return &apis.FieldError{
			Message: "No agent specified for stage or for its parent(s)",
			Paths:   []string{"agent"},
		}
	}

	if len(s.Steps) > 0 {
		if len(s.Stages) > 0 || len(s.Parallel) > 0 {
			return apis.ErrMultipleOneOf("steps", "stages", "parallel")
		}
		for i, step := range s.Steps {
			if err := validateStep(step).ViaFieldIndex("steps", i); err != nil {
				return err
			}
		}
	}

	if len(s.Stages) > 0 {
		if len(s.Parallel) > 0 {
			return apis.ErrMultipleOneOf("steps", "stages", "parallel")
		}
		for i, stage := range s.Stages {
			if err := validateStage(stage, parentAgent).ViaFieldIndex("stages", i); err != nil {
				return err
			}
		}
	}

	if len(s.Parallel) > 0 {
		for i, stage := range s.Parallel {
			return validateStage(stage, parentAgent).ViaFieldIndex("parallel", i)
		}
	}

	return validateStageOptions(s.Options).ViaField("options")
}

func validateStep(s Step) *apis.FieldError {
	if s.Command == "" && s.Step == "" {
		return apis.ErrMissingOneOf("command", "step")
	}

	if s.Command != "" {
		if s.Step != "" {
			return apis.ErrMultipleOneOf("command", "step")
		} else if len(s.Options) > 0 {
			return &apis.FieldError{
				Message: "Cannot set options for a command",
				Paths:   []string{"options"},
			}
		}
	}

	if s.Step != "" && len(s.Arguments) != 0 {
		return &apis.FieldError{
			Message: "Cannot set command-line arguments for a step",
			Paths:   []string{"args"},
		}
	}

	return validateAgent(s.Agent).ViaField("agent")
}

func validateStages(stages []Stage, parentAgent Agent) *apis.FieldError {
	if len(stages) == 0 {
		return apis.ErrMissingField("stages")
	}

	for i, s := range stages {
		if err := validateStage(s, parentAgent).ViaFieldIndex("stages", i); err != nil {
			return err
		}
	}

	return nil
}

func validateRootOptions(o RootOptions) *apis.FieldError {
	if !equality.Semantic.DeepEqual(o, RootOptions{}) {
		if !equality.Semantic.DeepEqual(o.Timeout, Timeout{}) {
			if err := validateTimeout(o.Timeout); err != nil {
				return err.ViaField("timeout")
			}
		}

		// TODO: retry will default to 0, so we're kinda stuck checking if it's less than zero here.
		if o.Retry < 0 {
			return &apis.FieldError{
				Message: "Retry count cannot be negative",
				Paths:   []string{"retry"},
			}
		}
	}

	return nil
}

func validateStageOptions(o StageOptions) *apis.FieldError {
	if !equality.Semantic.DeepEqual(o.Stash, Stash{}) {
		if err := validateStash(o.Stash); err != nil {
			return err.ViaField("stash")
		}
	}

	if !equality.Semantic.DeepEqual(o.Unstash, Unstash{}) {
		if err := validateUnstash(o.Unstash); err != nil {
			return err.ViaField("unstash")
		}
	}

	if o.Workspace != nil {
		if err := validateWorkspace(*o.Workspace); err != nil {
			return err
		}
	}

	return validateRootOptions(o.RootOptions)
}

func validateTimeout(t Timeout) *apis.FieldError {
	if !equality.Semantic.DeepEqual(t, Timeout{}) {
		isAllowed := false
		for _, allowed := range AllTimeoutUnits {
			if t.Unit == allowed {
				isAllowed = true
			}
		}

		if !isAllowed {
			return &apis.FieldError{
				Message: fmt.Sprintf("%s is not a valid time unit. Valid time units are %s", string(t.Unit),
					strings.Join(allTimeoutUnitsAsStrings(), ", ")),
				Paths: []string{"unit"},
			}
		}

		if t.Time < 1 {
			return &apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}
		}
	}

	return nil
}

func validateUnstash(u Unstash) *apis.FieldError {
	if !equality.Semantic.DeepEqual(u, Unstash{}) {
		// TODO: Check to make sure the corresponding stash is defined somewhere
		if u.Name == "" {
			return &apis.FieldError{
				Message: "The unstash name must be provided",
				Paths:   []string{"name"},
			}
		}
	}

	return nil
}

func validateStash(s Stash) *apis.FieldError {
	if !equality.Semantic.DeepEqual(s, Stash{}) {
		if s.Name == "" {
			return &apis.FieldError{
				Message: "The stash name must be provided",
				Paths:   []string{"name"},
			}
		}
		if s.Files == "" {
			return &apis.FieldError{
				Message: "files to stash must be provided",
				Paths:   []string{"files"},
			}
		}
	}

	return nil
}

func validateWorkspace(w string) *apis.FieldError {
	if w == "" {
		return &apis.FieldError{
			Message: "The workspace name must be unspecified or non-empty",
			Paths:   []string{"workspace"},
		}
	}

	return nil
}

var randReader = rand.Reader

func scopedEnv(s Stage, parentEnv []corev1.EnvVar) []corev1.EnvVar {
	if len(parentEnv) == 0 && len(s.Environment) == 0 {
		return nil
	}
	envMap := make(map[string]corev1.EnvVar)

	for _, e := range parentEnv {
		envMap[e.Name] = e
	}

	for _, e := range s.Environment {
		envMap[e.Name] = corev1.EnvVar{
			Name:  e.Name,
			Value: e.Value,
		}
	}

	env := make([]corev1.EnvVar, 0, len(envMap))

	for _, value := range envMap {
		env = append(env, value)
	}

	return env
}

func (j *Jenkinsfile) toStepEnvVars() []corev1.EnvVar {
	env := make([]corev1.EnvVar, 0, len(j.Environment))

	for _, e := range j.Environment {
		env = append(env, corev1.EnvVar{Name: e.Name, Value: e.Value})
	}

	return env
}

type TransformedStage struct {
	Stage Stage
	// Only one of Sequential, Parallel, and Task is non-empty
	Sequential []*TransformedStage
	Parallel   []*TransformedStage
	Task       *pipelinev1alpha1.Task
	// PipelineTask is non-empty only if Task is non-empty, but it is populated
	// after Task is populated so the reverse is not true.
	PipelineTask *pipelinev1alpha1.PipelineTask
	// The depth of this stage in the full tree of stages
	Depth int8
	// The parallel or sequntial stage enclosing this stage, or nil if this stage is at top level
	EnclosingStage *TransformedStage
	// The stage immediately before this stage at the same depth, or nil if there is no such stage
	PreviousSiblingStage *TransformedStage
	// TODO: Add the equivalent reverse relationship
}

func (ts TransformedStage) isSequential() bool {
	return len(ts.Sequential) > 0
}

func (ts TransformedStage) isParallel() bool {
	return len(ts.Parallel) > 0
}

func (ts TransformedStage) getLinearTasks() []*pipelinev1alpha1.Task {
	if ts.isSequential() {
		var tasks []*pipelinev1alpha1.Task
		for _, seqTs := range ts.Sequential {
			tasks = append(tasks, seqTs.getLinearTasks()...)
		}
		return tasks
	} else if ts.isParallel() {
		var tasks []*pipelinev1alpha1.Task
		for _, parTs := range ts.Parallel {
			tasks = append(tasks, parTs.getLinearTasks()...)
		}
		return tasks
	} else {
		return []*pipelinev1alpha1.Task{ts.Task}
	}
}

// If the workspace is nil, sets it to the parent's workspace
func (ts *TransformedStage) computeWorkspace(parentWorkspace string) {
	if ts.Stage.Options.Workspace == nil {
		ts.Stage.Options.Workspace = &parentWorkspace
	}
}

func stageToTask(s Stage, pipelineIdentifier string, buildIdentifier string, namespace string, wsPath string, parentEnv []corev1.EnvVar, parentAgent Agent, parentWorkspace string, suffix string, depth int8, enclosingStage *TransformedStage, previousSiblingStage *TransformedStage) (*TransformedStage, error) {
	if len(s.Post) != 0 {
		return nil, errors.New("post on stages not yet supported")
	}

	if !equality.Semantic.DeepEqual(s.Options, StageOptions{}) {
		o := s.Options
		if !equality.Semantic.DeepEqual(o.Timeout, Timeout{}) {
			return nil, errors.New("Timeout on stage not yet supported")
		}
		if o.Retry != 0 {
			return nil, errors.New("Retry on stage not yet supported")
		}
		if !equality.Semantic.DeepEqual(o.Stash, Stash{}) {
			return nil, errors.New("Stash on stage not yet supported")
		}
		if !equality.Semantic.DeepEqual(o.Unstash, Unstash{}) {
			return nil, errors.New("Unstash on stage not yet supported")
		}
	}

	env := scopedEnv(s, parentEnv)

	agent := s.Agent

	if equality.Semantic.DeepEqual(agent, Agent{}) {
		agent = parentAgent
	}

	if len(s.Steps) > 0 {
		if suffix == "" {
			// Generate a short random hex string.
			b, err := ioutil.ReadAll(io.LimitReader(randReader, 3))
			if err != nil {
				return nil, err
			}
			suffix = hex.EncodeToString(b)
		}

		t := &pipelinev1alpha1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      MangleToRfc1035Label(fmt.Sprintf("%s-build-%s-stage-%s", pipelineIdentifier, buildIdentifier, s.Name), suffix),
			},
		}
		t.SetDefaults()

		ws := &pipelinev1alpha1.TaskResource{
			Name: "workspace",
			Type: pipelinev1alpha1.PipelineResourceTypeGit,
		}

		if wsPath != "" {
			ws.TargetPath = wsPath
		}

		t.Spec.Inputs = &pipelinev1alpha1.Inputs{
			Resources: []pipelinev1alpha1.TaskResource{*ws,
				{
					Name: "temp-ordering-resource",
					Type: pipelinev1alpha1.PipelineResourceTypeImage,
				},
			},
		}

		t.Spec.Outputs = &pipelinev1alpha1.Outputs{
			Resources: []pipelinev1alpha1.TaskResource{
				{
					Name: "workspace",
					Type: pipelinev1alpha1.PipelineResourceTypeGit,
				},
				{
					Name: "temp-ordering-resource",
					Type: pipelinev1alpha1.PipelineResourceTypeImage,
				},
			},
		}

		for i, step := range s.Steps {
			// TODO: Ignoring everything but commands right now, but will eventually need to handle syntactic sugar steps too
			if step.Command != "" {
				stepImage := agent.Image
				if !equality.Semantic.DeepEqual(step.Agent, Agent{}) {
					stepImage = step.Agent.Image
				}
				t.Spec.Steps = append(t.Spec.Steps, corev1.Container{
					Name:    MangleToRfc1035Label(fmt.Sprintf("stage-%s-step-%d", s.Name, i), suffix),
					Env:     env,
					Image:   stepImage,
					Command: []string{step.Command},
					Args:    step.Arguments,
				})
			} else {
				return nil, errors.New("syntactic sugar steps not yet supported")
			}
		}
		ts := TransformedStage{Stage: s, Task: t, Depth: depth, EnclosingStage: enclosingStage, PreviousSiblingStage: previousSiblingStage}
		ts.computeWorkspace(parentWorkspace)
		return &ts, nil
	}

	if len(s.Stages) > 0 {
		var tasks []*TransformedStage
		ts := TransformedStage{Stage: s, Depth: depth, EnclosingStage: enclosingStage, PreviousSiblingStage: previousSiblingStage}
		ts.computeWorkspace(parentWorkspace)

		for i, nested := range s.Stages {
			nestedWsPath := ""
			if wsPath != "" && i == 0 {
				nestedWsPath = wsPath
			}
			var nestedPreviousSibling *TransformedStage
			if i > 0 {
				nestedPreviousSibling = tasks[i-1]
			}
			nestedTask, err := stageToTask(nested, pipelineIdentifier, buildIdentifier, namespace, nestedWsPath, env, agent, *ts.Stage.Options.Workspace, suffix, depth+1, &ts, nestedPreviousSibling)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, nestedTask)
		}
		ts.Sequential = tasks

		return &ts, nil
	}

	if len(s.Parallel) > 0 {
		var tasks []*TransformedStage
		ts := TransformedStage{Stage: s, Depth: depth, EnclosingStage: enclosingStage, PreviousSiblingStage: previousSiblingStage}
		ts.computeWorkspace(parentWorkspace)

		for i, nested := range s.Parallel {
			nestedWsPath := ""
			if wsPath != "" && i == 0 {
				nestedWsPath = wsPath
			}
			nestedTask, err := stageToTask(nested, pipelineIdentifier, buildIdentifier, namespace, nestedWsPath, env, agent, *ts.Stage.Options.Workspace, suffix, depth+1, &ts, nil)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, nestedTask)
		}
		ts.Parallel = tasks

		return &ts, nil
	}

	return nil, errors.New("no steps, sequential stages, or parallel stages")
}

func (j *Jenkinsfile) GenerateCRDs(pipelineIdentifier string, buildIdentifier string, namespace string, suffix string) (*pipelinev1alpha1.Pipeline, []*pipelinev1alpha1.Task, error) {
	if len(j.Post) != 0 {
		return nil, nil, errors.New("post at top level not yet supported")
	}
	if !equality.Semantic.DeepEqual(j.Options, RootOptions{}) {
		return nil, nil, errors.New("options at top level not yet supported")
	}

	if suffix == "" {
		// Generate a short random hex string.
		b, err := ioutil.ReadAll(io.LimitReader(randReader, 3))
		if err != nil {
			return nil, nil, err
		}
		suffix = hex.EncodeToString(b)
	}

	p := &pipelinev1alpha1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("%s-build-%s-%s", pipelineIdentifier, buildIdentifier, suffix),
		},
		Spec: pipelinev1alpha1.PipelineSpec{
			Resources: []pipelinev1alpha1.PipelineDeclaredResource{
				{
					Name: "common-workspace",
					Type: pipelinev1alpha1.PipelineResourceTypeGit,
				},
				{
					// TODO: Switch from this kind of hackish approach to non-resource-based dependencies once they land.
					Name: "temp-ordering-resource",
					Type: pipelinev1alpha1.PipelineResourceTypeImage,
				},
			},
		},
	}

	p.SetDefaults()

	var previousStage *TransformedStage
	var tasks []*pipelinev1alpha1.Task

	baseEnv := j.toStepEnvVars()

	for _, s := range j.Stages {
		wsPath := ""
		if len(tasks) == 0 {
			wsPath = "workspace"
		}
		stage, err := stageToTask(s, pipelineIdentifier, buildIdentifier, namespace, wsPath, baseEnv, j.Agent, "default", suffix, 0, nil, previousStage)
		if err != nil {
			return nil, nil, err
		}
		previousStage = stage

		tasks = append(tasks, stage.getLinearTasks()...)
		p.Spec.Tasks = append(p.Spec.Tasks, createPipelineTasks(stage)...)
	}

	return p, tasks, nil
}

func createPipelineTasks(stage *TransformedStage) []pipelinev1alpha1.PipelineTask {
	if stage.isSequential() {
		var pTasks []pipelinev1alpha1.PipelineTask
		for _, nestedStage := range stage.Sequential {
			pTasks = append(pTasks, createPipelineTasks(nestedStage)...)
		}
		return pTasks
	} else if stage.isParallel() {
		var pTasks []pipelinev1alpha1.PipelineTask
		for _, nestedStage := range stage.Parallel {
			pTasks = append(pTasks, createPipelineTasks(nestedStage)...)
		}
		return pTasks
	} else {
		pTask := pipelinev1alpha1.PipelineTask{
			Name: stage.Stage.TaskName(), // TODO: What should this actually be named?
			TaskRef: pipelinev1alpha1.TaskRef{
				Name: stage.Task.Name,
			},
		}

		_, provider := findWorkspaceProvider(stage, stage.getEnclosing(0))
		var previousStageNames []string
		for _, previousStage := range findPreviousNonBlockStages(*stage) {
			previousStageNames = append(previousStageNames, previousStage.Task.Name)
		}
		pTask.Resources = &pipelinev1alpha1.PipelineTaskResources{
			Inputs: []pipelinev1alpha1.PipelineTaskInputResource{
				{
					Name:     "workspace",
					Resource: "common-workspace",
					From:     provider,
				},
				{
					// TODO: Switch from this kind of hackish approach to non-resource-based dependencies once they land.
					Name:     "temp-ordering-resource",
					Resource: "temp-ordering-resource",
					From:     previousStageNames,
				},
			},
		}
		stage.PipelineTask = &pTask

		return []pipelinev1alpha1.PipelineTask{pTask}
	}
}

// Looks for the most recent Task using the desired workspace that was not in the
// same parallel stage and returns the name of the corresponding Task.
func findWorkspaceProvider(stage, sibling *TransformedStage) (bool, []string) {
	if *stage.Stage.Options.Workspace == "empty" {
		return true, nil
	}

	for sibling != nil {
		if sibling.isSequential() {
			found, provider := findWorkspaceProvider(stage, sibling.Sequential[len(sibling.Sequential)-1])
			if found {
				return true, provider
			}
		} else if sibling.isParallel() {
			// We don't want to use a workspace from a parallel stage outside of that stage,
			// but we do need to descend inwards in case stage is in that same stage.
			if stage.getEnclosing(sibling.Depth) == sibling {
				for _, nested := range sibling.Parallel {
					// Pick the parallel branch that has stage
					if stage.getEnclosing(nested.Depth) == nested {
						found, provider := findWorkspaceProvider(stage, nested)
						if found {
							return true, provider
						}
					}
				}
			}
			// TODO: What to do about custom workspaces? Check for erroneous uses specially?
			// Allow them if only one of the parallel tasks uses the same resource?
		} else if sibling.PipelineTask != nil {
			if *sibling.Stage.Options.Workspace == *stage.Stage.Options.Workspace {
				return true, []string{sibling.Task.Name}
			}
		} else {
			// We are in a sequential stage and sibling has not had its PipelineTask created.
			// Check the task before it so we don't use a workspace of a later task.
		}
		sibling = sibling.PreviousSiblingStage
	}

	return false, nil
}

// Find the end tasks for this stage, traversing down to the end stages of any
// nested sequential or parallel stages as well.
func findEndStages(stage TransformedStage) []*TransformedStage {
	if stage.isSequential() {
		return findEndStages(*stage.Sequential[len(stage.Sequential)-1])
	} else if stage.isParallel() {
		var endTasks []*TransformedStage
		for _, pStage := range stage.Parallel {
			endTasks = append(endTasks, findEndStages(*pStage)...)
		}
		return endTasks
	} else {
		return []*TransformedStage{&stage}
	}
}

// Find the tasks that run immediately before this stage, not including
// sequential or parallel wrapper stages.
func findPreviousNonBlockStages(stage TransformedStage) []*TransformedStage {
	if stage.PreviousSiblingStage != nil {
		return findEndStages(*stage.PreviousSiblingStage)
	} else if stage.EnclosingStage != nil {
		return findPreviousNonBlockStages(*stage.EnclosingStage)
	} else {
		return []*TransformedStage{}
	}
}

// Return the stage that encloses this stage at the given depth, or nil if there is no such stage.
// Depth must be >= 0. Returns the stage itself if depth == stage.Depth
func (stage *TransformedStage) getEnclosing(depth int8) *TransformedStage {
	if stage.Depth == depth {
		return stage
	} else if stage.EnclosingStage == nil {
		return nil
	} else {
		return stage.EnclosingStage.getEnclosing(depth)
	}
}

// Return the first stage that will execute before this stage
// Depth must be >= 0
func (stage TransformedStage) getClosestAncestor() *TransformedStage {
	if stage.PreviousSiblingStage != nil {
		return stage.PreviousSiblingStage
	} else if stage.EnclosingStage == nil {
		return nil
	} else {
		return stage.EnclosingStage.getClosestAncestor()
	}
}
