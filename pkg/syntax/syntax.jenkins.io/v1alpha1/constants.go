package v1alpha1

const (
	// TektonAPIVersion the APIVersion for using Tekton
	TektonAPIVersion = "tekton.dev/v1alpha1"

	// LabelStageName - the name for the label that will have the stage name on the Task.
	LabelStageName = "jenkins.io/task-stage-name"

	// DefaultStageNameForBuildPack - the name we use for the single stage created from build packs currently.
	DefaultStageNameForBuildPack = "from-build-pack"
)
