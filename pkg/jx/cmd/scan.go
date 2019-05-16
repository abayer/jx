package cmd

import (
	"github.com/jenkins-x/jx/pkg/jx/cmd/opts"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	"github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/spf13/cobra"
)

// Scan Options contains the command line options for scan commands
type ScanOptions struct {
	*opts.CommonOptions
}

var (
	scan_long = templates.LongDesc(`
		Perform a scan action.
	`)
)

// NewCmdScan creates a command object for the "scan" command
func NewCmdScan(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &ScanOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Perform a scan action",
		Long:  scan_long,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			util.CheckErr(err)
		},
	}

	cmd.AddCommand(NewCmdScanCluster(commonOpts))

	return cmd
}

// Run executes the scan commands
func (o *ScanOptions) Run() error {
	return o.Cmd.Help()
}
