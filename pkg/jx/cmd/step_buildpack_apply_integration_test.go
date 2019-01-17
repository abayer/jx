// +build integration

package cmd_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/pkg/builds"
	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/jenkins-x/jx/pkg/helm"
	"github.com/jenkins-x/jx/pkg/jenkins"
	"github.com/jenkins-x/jx/pkg/jx/cmd"
	"github.com/jenkins-x/jx/pkg/testkube"
	"github.com/jenkins-x/jx/pkg/tests"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStepBuildPackApply(t *testing.T) {
	const buildPackURL = builds.KubernetesWorkloadBuildPackURL
	const buildPackRef = "master"

	tests.SkipForWindows(t, "go-expect does not work on windows")
	t.Parallel()

	tempDir, err := ioutil.TempDir("", "test-step-buildpack-apply")
	require.NoError(t, err)

	testData := path.Join("test_data", "import_projects", "maven_camel")
	_, err = os.Stat(testData)
	require.NoError(t, err)

	err = util.CopyDir(testData, tempDir, true)
	require.NoError(t, err)

	o := &cmd.StepBuildPackApplyOptions{
		StepOptions: cmd.StepOptions{
			CommonOptions: cmd.CommonOptions{
				In:  os.Stdin,
				Out: os.Stdout,
				Err: os.Stderr,
			},
		},
		Dir: tempDir,
	}

	cmd.ConfigureTestOptionsWithResources(&o.CommonOptions,
		[]runtime.Object{
			testkube.CreateFakeGitSecret(),
		},
		[]runtime.Object{},
		gits.NewGitCLI(),
		helm.NewHelmCLI("helm", helm.V2, "", true),
	)

	err = o.ModifyDevEnvironment(func(env *v1.Environment) error {
		env.Spec.TeamSettings.BuildPackURL = buildPackURL
		env.Spec.TeamSettings.BuildPackRef = buildPackRef
		return nil
	})
	require.NoError(t, err)

	// lets configure the correct build pack for our test
	settings, err := o.TeamSettings()
	require.NoError(t, err)
	assert.Equal(t, buildPackURL, settings.BuildPackURL, "TeamSettings.BuildPackURL")
	assert.Equal(t, buildPackRef, settings.BuildPackRef, "TeamSettings.BuildPackRef")

	err = o.Run()
	require.NoError(t, err, "failed to run step")

	actualJenkinsfile := filepath.Join(tempDir, jenkins.DefaultJenkinsfile)
	assert.FileExists(t, actualJenkinsfile, "No Jenkinsfile created!")

	t.Logf("Found Jenkinsfile at %s\n", actualJenkinsfile)
}
