package cmd

import (
	"testing"

	"github.com/jenkins-x/jx/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/assert"
)

// PollGitStatusAndReactToPipelineChanges performs polling and responds to changes in PipelineActivity
func PollGitStatusAndReactToPipelineChanges(t *testing.T, o *ControllerWorkflowOptions, jxClient versioned.Interface, ns string) error {
	o.ReloadAndPollGitPipelineStatuses(jxClient, ns)
	err := o.Run()
	assert.NoError(t, err, "Failed to react to PipelineActivity changes")
	return err
}

