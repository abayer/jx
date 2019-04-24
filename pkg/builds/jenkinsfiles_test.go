package builds

import (
	"testing"

	"github.com/jenkins-x/jx/pkg/syntax/syntax.jenkins.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestJenkinsfileGenerator(t *testing.T) {
	t.Parallel()
	projectConfig := &v1alpha1.ProjectConfig{
		BuildPack: "maven",
		Env: []corev1.EnvVar{
			{
				Name:  "ORG",
				Value: "myorg",
			},
			{
				Name:  "APP_NAME",
				Value: "thingy",
			},
		},
		PipelineConfig: &v1alpha1.PipelineConfig{
			Pipelines: v1alpha1.Pipelines{
				PullRequest: &v1alpha1.PipelineLifecycles{
					Build: &v1alpha1.PipelineLifecycle{
						Steps: []*v1alpha1.PipelineStep{
							{
								Command: "mvn test",
							},
						},
					},
				},
				Release: &v1alpha1.PipelineLifecycles{
					Build: &v1alpha1.PipelineLifecycle{
						Steps: []*v1alpha1.PipelineStep{
							{
								Command: "mvn test",
							},
							{
								Command: "mvn deploy",
							},
							{
								Command: "jx promote --all-auto",
							},
						},
					},
				},
			},
			Env: []corev1.EnvVar{
				{
					Name:  "PREVIEW_VERSION",
					Value: "0.0.0-SNAPSHOT-$BRANCH_NAME-$BUILD_NUMBER",
				},
			},
		},
	}

	text, err := NewJenkinsConverter(projectConfig).ToJenkinsfile()
	assert.NoError(t, err)

	t.Logf("Generated: %s\n", text)
}
