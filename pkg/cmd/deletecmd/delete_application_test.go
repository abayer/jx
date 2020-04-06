// +build unit

package deletecmd

import (
	"testing"

	v1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"

	"github.com/stretchr/testify/assert"
)

func TestRemoveEnvrionmentFromJobs(t *testing.T) {
	t.Parallel()

	fakeEnv := &v1.Environment{
		Spec: v1.EnvironmentSpec{
			Source: v1.EnvironmentRepository{
				URL: "http://github.com/owner/environment-repo-dev.git",
			},
		},
	}

	tests := []struct {
		name string
		jobs []string
		envs map[string]*v1.Environment
		want []string
	}{
		{
			"when there's no matching environment",
			[]string{"owner/job-repo", "owner/environment-foobar"},
			map[string]*v1.Environment{"dev": fakeEnv},
			[]string{"owner/job-repo", "owner/environment-foobar"},
		},
		{
			"when there's an environment",
			[]string{"owner/job-repo", "owner/environment-repo-dev"},
			map[string]*v1.Environment{"dev": fakeEnv},
			[]string{"owner/job-repo"},
		},
	}

	for _, test := range tests {
		got := removeEnvironments(test.jobs, test.envs)
		assert.Equal(t, test.want, got, test.name)
	}
}
