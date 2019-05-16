package cmd

import (
	"github.com/jenkins-x/jx/pkg/jx/cmd/opts"
	"github.com/jenkins-x/jx/pkg/jx/cmd/util"
	tektoncmd "github.com/jenkins-x/jx/pkg/tekton/cmd"
	"github.com/spf13/cobra"
)

// StepCreateOptions contains the command line flags
type StepCreateOptions struct {
	opts.StepOptions
}

// NewCmdStepCreate Steps a command object for the "step" command
func NewCmdStepCreate(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &StepCreateOptions{
		StepOptions: opts.StepOptions{
			CommonOptions: commonOpts,
		},
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create [command]",
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			util.CheckErr(err)
		},
	}
	cmd.AddCommand(NewCmdStepCreateJenkinsConfig(commonOpts))
	cmd.AddCommand(tektoncmd.NewCmdStepCreateTask(commonOpts))
	cmd.AddCommand(NewCmdStepCreateVersionPullRequest(commonOpts))
	return cmd
}

// Run implements this command
func (o *StepCreateOptions) Run() error {
	return o.Cmd.Help()
}
