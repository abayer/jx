package opts

import (
	"os"

	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/prow"
	"github.com/pkg/errors"
)

// StepOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type StepOptions struct {
	*CommonOptions

	DisableImport bool
	OutDir        string
}

// Run implements this command
func (o *StepOptions) Run() error {
	return o.Cmd.Help()
}

// StepGitMergeOptions contains the command line flags
type StepGitMergeOptions struct {
	StepOptions

	SHAs       []string
	Remote     string
	Dir        string
	BaseBranch string
	BaseSHA    string
}

// Run implements the command
func (o *StepGitMergeOptions) Run() error {
	if o.Remote == "" {
		o.Remote = "origin"
	}

	// set dummy git config details if none set so we can do a local commit when merging
	err := o.setGitConfig()
	if err != nil {
		return errors.Wrapf(err, "failed to set git config")
	}
	if len(o.SHAs) == 0 || o.BaseBranch == "" || o.BaseSHA == "" {
		// Try to look in the env vars
		if pullRefs := os.Getenv("PULL_REFS"); pullRefs != "" {
			log.Infof("Using SHAs from PULL_REFS=%s\n", pullRefs)
			pullRefs, err := prow.ParsePullRefs(pullRefs)
			if err != nil {
				return errors.Wrapf(err, "parsing PULL_REFS=%s", pullRefs)
			}
			if len(o.SHAs) == 0 {
				o.SHAs = make([]string, 0)
				for _, sha := range pullRefs.ToMerge {
					o.SHAs = append(o.SHAs, sha)
				}
			}
			if o.BaseBranch == "" {
				o.BaseBranch = pullRefs.BaseBranch
			}
			if o.BaseSHA == "" {
				o.BaseSHA = pullRefs.BaseSha
			}
		}
	}
	if len(o.SHAs) == 0 {
		log.Warnf("no SHAs to merge, falling back to initial cloned commit")
		return nil
	}
	return gits.FetchAndMergeSHAs(o.SHAs, o.BaseBranch, o.BaseSHA, o.Remote, o.Dir, o.Git(), o.Verbose)
}

func (o *StepGitMergeOptions) setGitConfig() error {
	user, err := o.GetCommandOutput(o.Dir, "git", "config", "user.name")
	if err != nil || user == "" {
		err := o.RunCommandFromDir(o.Dir, "git", "config", "user.name", "jenkins-x")
		if err != nil {
			return err
		}
	}
	email, err := o.GetCommandOutput(o.Dir, "git", "config", "user.email")
	if email == "" || err != nil {
		err := o.RunCommandFromDir(o.Dir, "git", "config", "user.email", "jenkins-x@googlegroups.com")
		if err != nil {
			return err
		}
	}
	return nil
}
