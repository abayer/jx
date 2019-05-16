package cmd

import (
	"github.com/jenkins-x/jx/pkg/jx/cmd/opts"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	"github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/spf13/cobra"
)

var (
	// StepGitMergeLong command long description
	StepGitMergeLong = templates.LongDesc(`
		This pipeline step merges any SHAs specified into the HEAD of master. 

If no SHAs are specified then the PULL_REFS environment variable will be prased for a branch:sha comma separated list of
shas to merge. For example:

master:ef08a6cd194c2687d4bc12df6bb8a86f53c348ba,2739:5b351f4eae3c4afbb90dd7787f8bf2f8c454723f,2822:bac2a1f34fd54811fb767f69543f59eb3949b2a5

`)
	// StepGitMergeExample command example
	StepGitMergeExample = templates.Examples(`
		# Merge the SHAs from the PULL_REFS environment variable
		jx step git merge

		# Merge the SHA into the HEAD of master
		jx step git merge --sha 123456a

		# Merge a number of SHAs into the HEAD of master
		jx step git merge --sha 123456a --sha 789012b
`)
)

// NewCmdStepGitMerge create the 'step git envs' command
func NewCmdStepGitMerge(commonOpts *opts.CommonOptions) *cobra.Command {
	options := opts.StepGitMergeOptions{
		StepOptions: opts.StepOptions{
			CommonOptions: commonOpts,
		},
	}
	cmd := &cobra.Command{
		Use:     "merge",
		Short:   "Merge a number of SHAs into the HEAD of master",
		Long:    StepGitMergeLong,
		Example: StepGitMergeExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			util.CheckErr(err)
		},
	}

	cmd.Flags().StringArrayVarP(&options.SHAs, "sha", "", make([]string, 0), "The SHA(s) to merge, "+
		"if not specified then the value of the env var PULL_REFS is used")
	cmd.Flags().StringVarP(&options.Remote, "remote", "", "origin", "The name of the remote")
	cmd.Flags().StringVarP(&options.Dir, "dir", "", "", "The directory in which the git repo is checked out")
	cmd.Flags().StringVarP(&options.BaseBranch, "baseBranch", "", "", "The branch to merge to, "+
		"if not specified then the  first entry in PULL_REFS is used ")
	cmd.Flags().StringVarP(&options.BaseSHA, "baseSHA", "", "", "The SHA to use on the base branch, "+
		"if not specified then the first entry in PULL_REFS is used")

	return cmd
}

