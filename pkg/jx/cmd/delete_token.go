package cmd

import (
	"github.com/jenkins-x/jx/pkg/jx/cmd/opts"
	"github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/spf13/cobra"
)

// DeleteTokenOptions are the flags for delete commands
type DeleteTokenOptions struct {
	*opts.CommonOptions
}

// NewCmdDeleteToken creates a command object
// retrieves one or more resources from a server.
func NewCmdDeleteToken(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &DeleteTokenOptions{
		commonOpts,
	}

	cmd := &cobra.Command{
		Use:     "token",
		Short:   "Deletes one or more issue token resources",
		Aliases: []string{"api-token"},
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			util.CheckErr(err)
		},
	}

	cmd.AddCommand(NewCmdDeleteTokenAddon(commonOpts))
	return cmd
}

// Run implements this command
func (o *DeleteTokenOptions) Run() error {
	return o.Cmd.Help()
}
