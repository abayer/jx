package buildpipeline

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	tb "github.com/knative/build-pipeline/test/builder"
	"github.com/knative/pkg/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Write a builder for generating the expected objects. Because
// as this is now, there are way too many lines here.
func TestParseJenkinsfileYaml(t *testing.T) {
	// Needed to take address of strings since workspace is *string. Is there a better way to handle optional values?
	defaultWorkspace := "default"
	customWorkspace := "custom"

	tests := []struct {
		name             string
		expected         *Jenkinsfile
		pipeline         *pipelinev1alpha1.Pipeline
		tasks            []*pipelinev1alpha1.Task
		expectedErrorMsg string
	}{
		{
			name: "simple_jenkinsfile",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{{
					Name: "A Working Stage",
					Steps: []Step{{
						Command:   "echo",
						Arguments: []string{"hello", "world"},
					}},
				}},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
			},
		},
		{
			name: "multiple_stages",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
					{
						Name: "A Working Stage",
						Steps: []Step{{
							Command:   "echo",
							Arguments: []string{"hello", "world"},
						}},
					},
					{
						Name: "Another stage",
						Steps: []Step{{
							Command:   "echo",
							Arguments: []string{"again"},
						}},
					},
				},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("another-stage", "somepipeline-build-somebuild-stage-another-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("a-working-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("a-working-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-another-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-another-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("again")),
				)),
			},
		},
		{
			name: "nested_stages",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
					{
						Name: "Parent Stage",
						Stages: []Stage{
							{
								Name: "A Working Stage",
								Steps: []Step{{
									Command:   "echo",
									Arguments: []string{"hello", "world"},
								}},
							},
							{
								Name: "Another stage",
								Steps: []Step{{
									Command:   "echo",
									Arguments: []string{"again"},
								}},
							},
						},
					},
				},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("another-stage", "somepipeline-build-somebuild-stage-another-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("a-working-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("a-working-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-another-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-another-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("again")),
				)),
			},
		},
		{
			name: "parallel_stages",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
					{
						Name: "First Stage",
						Steps: []Step{{
							Command:   "echo",
							Arguments: []string{"first"},
						}},
					},
					{
						Name: "Parent Stage",
						Parallel: []Stage{
							{
								Name: "A Working Stage",
								Steps: []Step{{
									Command:   "echo",
									Arguments: []string{"hello", "world"},
								}},
							},
							{
								Name: "Another stage",
								Steps: []Step{{
									Command:   "echo",
									Arguments: []string{"again"},
								}},
							},
						},
					},
					{
						Name: "Last Stage",
						Steps: []Step{{
							Command:   "echo",
							Arguments: []string{"last"},
						}},
					},
				},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("first-stage", "somepipeline-build-somebuild-stage-first-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace", tb.From("first-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("first-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("another-stage", "somepipeline-build-somebuild-stage-another-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("first-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("first-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("last-stage", "somepipeline-build-somebuild-stage-last-stage-abcd",
					// TODO: Switch from this kind of hackish approach to non-resource-based dependencies once they land.
					tb.PipelineTaskInputResource("workspace", "common-workspace", tb.From("first-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("a-working-stage", "another-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-first-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-first-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("first")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-another-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-another-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("again")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-last-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-last-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("last")),
				)),
			},
		},
		{
			name: "parallel_and_nested_stages",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
					{
						Name: "First Stage",
						Steps: []Step{{
							Command:   "echo",
							Arguments: []string{"first"},
						}},
					},
					{
						Name: "Parent Stage",
						Parallel: []Stage{
							{
								Name: "A Working Stage",
								Steps: []Step{{
									Command:   "echo",
									Arguments: []string{"hello", "world"},
								}},
							},
							{
								Name: "Nested In Parallel",
								Stages: []Stage{
									{
										Name: "Another stage",
										Steps: []Step{{
											Command:   "echo",
											Arguments: []string{"again"},
										}},
									},
									{
										Name: "Some other stage",
										Steps: []Step{{
											Command:   "echo",
											Arguments: []string{"otherwise"},
										}},
									},
								},
							},
						},
					},
					{
						Name: "Last Stage",
						Steps: []Step{{
							Command:   "echo",
							Arguments: []string{"last"},
						}},
					},
				},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("first-stage", "somepipeline-build-somebuild-stage-first-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("first-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("first-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("another-stage", "somepipeline-build-somebuild-stage-another-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("first-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("first-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("some-other-stage", "somepipeline-build-somebuild-stage-some-other-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("another-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("another-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("last-stage", "somepipeline-build-somebuild-stage-last-stage-abcd",
					// TODO: Switch from this kind of hackish approach to non-resource-based dependencies once they land.
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("first-stage")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("a-working-stage", "some-other-stage")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-first-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-first-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("first")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-another-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-another-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("again")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-some-other-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-some-other-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("otherwise")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-last-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-last-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("last")),
				)),
			},
		},
		{
			name: "custom_workspaces",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
					{
						Name: "stage1",
						Steps: []Step{{
							Command: "ls",
						}},
					},
					{
						Name: "stage2",
						Options: StageOptions{
							Workspace: &customWorkspace,
						},
						Steps: []Step{{
							Command: "ls",
						}},
					},
					{
						Name: "stage3",
						Options: StageOptions{
							Workspace: &defaultWorkspace,
						},
						Steps: []Step{{
							Command: "ls",
						}},
					},
					{
						Name: "stage4",
						Options: StageOptions{
							Workspace: &customWorkspace,
						},
						Steps: []Step{{
							Command: "ls",
						}},
					},
				},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("stage1", "somepipeline-build-somebuild-stage-stage1-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("stage2", "somepipeline-build-somebuild-stage-stage2-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("stage1")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("stage3", "somepipeline-build-somebuild-stage-stage3-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace", tb.From("stage1")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("stage2")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("stage4", "somepipeline-build-somebuild-stage-stage4-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace", tb.From("stage2")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("stage3")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-stage1-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage1-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-stage2-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage2-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-stage3-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage3-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-stage4-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage4-step-0-abcd", "some-image", tb.Command("ls")),
				)),
			},
		},
		{
			name: "inherited_custom_workspaces",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
					{
						Name: "stage1",
						Steps: []Step{{
							Command: "ls",
						}},
					},
					{
						Name: "stage2",
						Options: StageOptions{
							Workspace: &customWorkspace,
						},
						Stages: []Stage{
							{
								Name: "stage3",
								Steps: []Step{{
									Command: "ls",
								}},
							},
							{
								Name: "stage4",
								Options: StageOptions{
									Workspace: &defaultWorkspace,
								},
								Steps: []Step{{
									Command: "ls",
								}},
							},
							{
								Name: "stage5",
								Steps: []Step{{
									Command: "ls",
								}},
							},
						},
					},
				},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("stage1", "somepipeline-build-somebuild-stage-stage1-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("stage3", "somepipeline-build-somebuild-stage-stage3-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("stage1")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("stage4", "somepipeline-build-somebuild-stage-stage4-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("stage1")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("stage3")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("stage5", "somepipeline-build-somebuild-stage-stage5-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From("stage3")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From("stage4")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-stage1-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage1-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-stage3-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage3-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-stage4-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage4-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-stage5-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-stage5-step-0-abcd", "some-image", tb.Command("ls")),
				)),
			},
		},
		{
			name: "environment_at_top_and_in_stage",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Environment: []EnvVar{{
					Name:  "SOME_VAR",
					Value: "A value for the env var",
				}},
				Stages: []Stage{{
					Name: "A stage with environment",
					Environment: []EnvVar{{
						Name:  "SOME_OTHER_VAR",
						Value: "A value for the other env var",
					}},
					Steps: []Step{{
						Command:   "echo",
						Arguments: []string{"hello", "${SOME_OTHER_VAR}"},
					}},
				}},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("a-stage-with-environment", "somepipeline-build-somebuild-stage-a-stage-with-environmen-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-stage-with-environmen-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-stage-with-environment-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "${SOME_OTHER_VAR}"),
						// TODO: Ordering doesn't seem to be deterministic.
						tb.EnvVar("SOME_VAR", "A value for the env var"), tb.EnvVar("SOME_OTHER_VAR", "A value for the other env var")),
				)),
			},
		},
		{
			name: "syntactic_sugar_step_and_a_command",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{{
					Name: "A Working Stage",
					Steps: []Step{
						{
							Command:   "echo",
							Arguments: []string{"hello", "world"},
						},
						{
							Step: "some-step",
							Options: map[string]string{
								"firstParam":  "some value",
								"secondParam": "some other value",
							},
						},
					},
				}},
			},
			expectedErrorMsg: "syntactic sugar steps not yet supported",
		},
		{
			name: "post",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{{
					Name: "A Working Stage",
					Steps: []Step{{
						Command:   "echo",
						Arguments: []string{"hello", "world"},
					}},
					Post: []Post{
						{
							Condition: "success",
							Actions: []PostAction{{
								Name: "mail",
								Options: map[string]string{
									"to":      "foo@bar.com",
									"subject": "Yay, it passed",
								},
							}},
						},
						{
							Condition: "failure",
							Actions: []PostAction{{
								Name: "slack",
								Options: map[string]string{
									"whatever": "the",
									"slack":    "config",
									"actually": "is. =)",
								},
							}},
						},
						{
							Condition: "always",
							Actions: []PostAction{{
								Name: "junit",
								Options: map[string]string{
									"pattern": "target/surefire-reports/**/*.xml",
								},
							}},
						},
					},
				}},
			},
			expectedErrorMsg: "post on stages not yet supported",
		},
		{
			name: "top_level_and_stage_options",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Options: RootOptions{
					Timeout: Timeout{
						Time: 50,
						Unit: "minutes",
					},
					Retry: 3,
				},
				Stages: []Stage{{
					Name: "A Working Stage",
					Options: StageOptions{
						RootOptions: RootOptions{
							Timeout: Timeout{
								Time: 5,
								Unit: "seconds",
							},
							Retry: 4,
						},
						Stash: Stash{
							Name:  "Some Files",
							Files: "somedir/**/*",
						},
						Unstash: Unstash{
							Name: "Earlier Files",
							Dir:  "some/sub/dir",
						},
					},
					Steps: []Step{{
						Command:   "echo",
						Arguments: []string{"hello", "world"},
					}},
				}},
			},
			expectedErrorMsg: "Retry at top level not yet supported",
		},
		{
			name: "stage_and_step_agent",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Stages: []Stage{{
					Name: "A Working Stage",
					Agent: Agent{
						Image: "some-image",
					},
					Steps: []Step{
						{
							Command:   "echo",
							Arguments: []string{"hello", "world"},
							Agent: Agent{
								Image: "some-other-image",
							},
						},
						{
							Command:   "echo",
							Arguments: []string{"goodbye"},
						},
					},
				}},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-other-image", tb.Command("echo"), tb.Args("hello", "world")),
					tb.Step("stage-a-working-stage-step-1-abcd", "some-image", tb.Command("echo"), tb.Args("goodbye")),
				)),
			},
		},
		{
			name: "mangled_task_names",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
					{
						Name: ". -a- .",
						Steps: []Step{{
							Command:   "ls",
							Arguments: nil,
						}},
					},
					{
						Name: "Wööh!!!! - This is cool.",
						Steps: []Step{{
							Command:   "ls",
							Arguments: nil,
						}},
					},
				},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask(".--a--.", "somepipeline-build-somebuild-stage-a-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineTask("wööh!!!!---this-is-cool.", "somepipeline-build-somebuild-stage-wh-this-is-cool-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace",
						tb.From(".--a--.")),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource",
						tb.From(".--a--.")),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-wh-this-is-cool-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-wh-this-is-cool-step-0-abcd", "some-image", tb.Command("ls")),
				)),
			},
		},
		{
			name: "stage_timeout",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{{
					Name: "A Working Stage",
					Options: StageOptions{
						RootOptions: RootOptions{
							Timeout: Timeout{
								Time: 50,
								Unit: "minutes",
							},
						},
					},
					Steps: []Step{{
						Command:   "echo",
						Arguments: []string{"hello", "world"},
					}},
				}},
			},
			/* TODO: Stop erroring out once we figure out how to handle task timeouts again
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskTimeout(50*time.Minute),
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
			},*/
			expectedErrorMsg: "Timeout on stage not yet supported",
		},
		{
			name: "top_level_timeout",
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Options: RootOptions{
					Timeout: Timeout{
						Time: 50,
						Unit: "minutes",
					},
				},
				Stages: []Stage{{
					Name: "A Working Stage",
					Steps: []Step{{
						Command:   "echo",
						Arguments: []string{"hello", "world"},
					}},
				}},
			},
			pipeline: tb.Pipeline("somepipeline-build-somebuild-abcd", "somenamespace", tb.PipelineSpec(
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd",
					tb.PipelineTaskInputResource("workspace", "common-workspace"),
					tb.PipelineTaskInputResource("temp-ordering-resource", "temp-ordering-resource"),
					tb.PipelineTaskOutputResource("workspace", "common-workspace"),
					tb.PipelineTaskOutputResource("temp-ordering-resource", "temp-ordering-resource")),
				tb.PipelineDeclaredResource("common-workspace", pipelinev1alpha1.PipelineResourceTypeGit),
				tb.PipelineDeclaredResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(
						tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
							tb.ResourceTargetPath("workspace")),
						tb.InputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit),
						tb.OutputsResource("temp-ordering-resource", pipelinev1alpha1.PipelineResourceTypeImage)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlFileName := filepath.Join("test_data", tt.name+".yaml")
			YamlToRead, err := ioutil.ReadFile(yamlFileName)
			if err != nil {
				t.Fatalf("Could not read yaml file: %s ", yamlFileName)
			}
			tt.name = string(YamlToRead)

			parsed, err := ParseJenkinsfileYaml(tt.name)
			if err != nil {
				t.Fatalf("Failed to parse YAML for %s: %q", tt.name, err)
			}

			if d := cmp.Diff(tt.expected, parsed); d != "" {
				t.Errorf("Parsed Jenkinsfile did not match expected: %s", d)
			}

			validateErr := parsed.Validate()
			if validateErr != nil {
				t.Errorf("Validation failed: %s", validateErr)
			}

			pipeline, tasks, err := parsed.GenerateCRDs("somepipeline", "somebuild", "somenamespace", "abcd")

			if err != nil {
				if tt.expectedErrorMsg != "" {
					if d := cmp.Diff(tt.expectedErrorMsg, err.Error()); d != "" {
						t.Fatalf("CRD generation error did not meet expectation: %s", d)
					}
				} else {
					t.Fatalf("Error generating CRDs: %s", err)
				}
			}

			if tt.expectedErrorMsg == "" {
				pipeline.TypeMeta = metav1.TypeMeta{}
				if d := cmp.Diff(tt.pipeline, pipeline); d != "" {
					t.Errorf("Generated Pipeline did not match expected: %s", d)
				}

				if err := pipeline.Spec.Validate(); err != nil {
					t.Errorf("PipelineSpec.Validate() = %v", err)
				}

				for _, task := range tasks {
					task.TypeMeta = metav1.TypeMeta{}
				}
				if d := cmp.Diff(tt.tasks, tasks); d != "" {
					t.Errorf("Generated Tasks did not match expected: %s", d)
				}

				for _, task := range tasks {
					if err := task.Spec.Validate(); err != nil {
						t.Errorf("TaskSpec.Validate() = %v", err)
					}
				}
			}
		})
	}
}

func TestFailedValidation(t *testing.T) {
	tests := []struct {
		name          string
		expectedError *apis.FieldError
	}{
		{
			name: "bad_apiVersion",
			expectedError: &apis.FieldError{
				Message: "Invalid apiVersion format: must be 'v(digits).(digits)",
				Paths:   []string{"apiVersion"},
			},
		},
		/* TODO: Once we figure out how to differentiate between an empty agent and no agent specified...
		{
			name: "empty_agent",
			expectedError: &apis.FieldError{
				Message: "Invalid apiVersion format: must be 'v(digits).(digits)",
				Paths:   []string{"apiVersion"},
			},
		},
		*/
		{
			name: "agent_with_both_image_and_label",
			expectedError: apis.ErrMultipleOneOf("label", "image").
				ViaField("agent"),
		},
		{
			name:          "no_stages",
			expectedError: apis.ErrMissingField("stages"),
		},
		{
			name:          "no_steps_stages_or_parallel",
			expectedError: apis.ErrMissingOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name:          "steps_and_stages",
			expectedError: apis.ErrMultipleOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name:          "steps_and_parallel",
			expectedError: apis.ErrMultipleOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name:          "stages_and_parallel",
			expectedError: apis.ErrMultipleOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name:          "step_without_command_or_step",
			expectedError: apis.ErrMissingOneOf("command", "step").ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name:          "step_with_both_command_and_step",
			expectedError: apis.ErrMultipleOneOf("command", "step").ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name: "step_with_command_and_options",
			expectedError: (&apis.FieldError{
				Message: "Cannot set options for a command",
				Paths:   []string{"options"},
			}).ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name: "step_with_step_and_arguments",
			expectedError: (&apis.FieldError{
				Message: "Cannot set command-line arguments for a step",
				Paths:   []string{"args"},
			}).ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name: "no_parent_or_stage_agent",
			expectedError: (&apis.FieldError{
				Message: "No agent specified for stage or for its parent(s)",
				Paths:   []string{"agent"},
			}).ViaFieldIndex("stages", 0),
		},
		{
			name: "top_level_timeout_without_time",
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options"),
		},
		{
			name: "stage_timeout_without_time",
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "top_level_timeout_with_invalid_unit",
			expectedError: (&apis.FieldError{
				Message: "years is not a valid time unit. Valid time units are seconds, minutes, hours, days",
				Paths:   []string{"unit"},
			}).ViaField("timeout").ViaField("options"),
		},
		{
			name: "stage_timeout_with_invalid_unit",
			expectedError: (&apis.FieldError{
				Message: "years is not a valid time unit. Valid time units are seconds, minutes, hours, days",
				Paths:   []string{"unit"},
			}).ViaField("timeout").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "top_level_timeout_with_invalid_time",
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options"),
		},
		{
			name: "stage_timeout_with_invalid_time",
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "top_level_retry_with_invalid_count",
			expectedError: (&apis.FieldError{
				Message: "Retry count cannot be negative",
				Paths:   []string{"retry"},
			}).ViaField("options"),
		},
		{
			name: "stage_retry_with_invalid_count",
			expectedError: (&apis.FieldError{
				Message: "Retry count cannot be negative",
				Paths:   []string{"retry"},
			}).ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "stash_without_name",
			expectedError: (&apis.FieldError{
				Message: "The stash name must be provided",
				Paths:   []string{"name"},
			}).ViaField("stash").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "stash_without_files",
			expectedError: (&apis.FieldError{
				Message: "files to stash must be provided",
				Paths:   []string{"files"},
			}).ViaField("stash").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "unstash_without_name",
			expectedError: (&apis.FieldError{
				Message: "The unstash name must be provided",
				Paths:   []string{"name"},
			}).ViaField("unstash").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "blank_stage_name",
			expectedError: (&apis.FieldError{
				Message: "Stage name must contain at least one ASCII letter",
				Paths:   []string{"name"},
			}).ViaFieldIndex("stages", 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlFile := filepath.Join("test_data", "validation_failures", tt.name+".yaml")
			YamlToRead, YamlReadErr := ioutil.ReadFile(yamlFile)
			if YamlReadErr != nil {
				t.Fatalf("Could not read yaml file: %s ", yamlFile)
			}
			tt.name = string(YamlToRead)

			parsed, parseErr := ParseJenkinsfileYaml(tt.name)
			if parseErr != nil {
				t.Fatalf("Failed to parse YAML for %s: %q", tt.name, parseErr)
			}

			err := parsed.Validate()

			if err == nil {
				t.Fatalf("Expected a validation failure but none occurred")
			}

			if d := cmp.Diff(tt.expectedError, err, cmp.AllowUnexported(apis.FieldError{})); d != "" {
				t.Fatalf("Validation error did not meet expectation: %s", d)
			}
		})
	}
}

func TestRfc1035LabelMangling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unmodified",
			input:    "unmodified",
			expected: "unmodified-suffix",
		},
		{
			name:     "spaces",
			input:    "A Simple Test.",
			expected: "a-simple-test-suffix",
		},
		{
			name:     "no leading digits",
			input:    "0123456789no-leading-digits",
			expected: "no-leading-digits-suffix",
		},
		{
			name:     "no leading hyphens",
			input:    "----no-leading-hyphens",
			expected: "no-leading-hyphens-suffix",
		},
		{
			name:     "no consecutive hyphens",
			input:    "no--consecutive- hyphens",
			expected: "no-consecutive-hyphens-suffix",
		},
		{
			name:     "no trailing hyphens",
			input:    "no-trailing-hyphens----",
			expected: "no-trailing-hyphens-suffix",
		},
		{
			name:     "no symbols",
			input:    "&$^#@(*&$^-whoops",
			expected: "whoops-suffix",
		},
		{
			name:     "no unprintable characters",
			input:    "a\n\t\x00b",
			expected: "ab-suffix",
		},
		{
			name:     "no unicode",
			input:    "japan-日本",
			expected: "japan-suffix",
		},
		{
			name:     "no non-bmp characters",
			input:    "happy 😃",
			expected: "happy-suffix",
		},
		{
			name:     "truncated to 63",
			input:    "a0123456789012345678901234567890123456789012345678901234567890123456789",
			expected: "a0123456789012345678901234567890123456789012345678901234-suffix",
		},
		{
			name:     "truncated to 62",
			input:    "a012345678901234567890123456789012345678901234567890123-567890123456789",
			expected: "a012345678901234567890123456789012345678901234567890123-suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mangled := MangleToRfc1035Label(tt.input, "suffix")
			if d := cmp.Diff(tt.expected, mangled); d != "" {
				t.Fatalf("Mangled output did not match expected output: %s", d)
			}
		})
	}
}

func testFindDuplicates(t *testing.T) {

	tests := []struct {
		name     string
		input    []string
		error   int
	}{
		{
			name:     "Two stage name duplicated" ,
			input:    []string{"Stage 1","Stage 1","Stage 2","Stage 2",},
			error:  2,
		},{
			name:     "One stage name duplicated" ,
			input:    []string{"Stage 1","Stage 1",},
			error:  1,
		},{
			name:     "No stage name duplicated" ,
			input:    []string{"Stage 0","Stage 1","Stage 2","Stage 3",},
			error:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := findDuplicates( tt.input )

			if tt.error == 0 && err != nil {
				t.Fatal("Not all duplicates found A")
			}

			if tt.error > 0 && len( err.Paths ) != tt.error {
				t.Fatal("Not all duplicates found")
			}


		})
	}
}

func TestFindDuplicatesWithStages (t *testing.T){
	tests := []struct {
		name     string
		expectedError    []string
	}{
		{
			name:     "stages_names_ok.yaml" ,
			expectedError:    []string{},
		},
		{
			name:     "stages_names_ok_with_sub_stages.yaml" ,
			expectedError:    []string{},
		},
		{
			name:     "stages_names_duplicates_with_sub_stages.yaml" ,
			expectedError:    []string{"Duplicate stage name 'Stage With Stages'",},
		},
		{
			name:     "stages_names_duplicates.yaml" ,
			expectedError:    []string{"Duplicate stage name 'A Working Stage'",},
		},
		{
			name:     "stages_names_with_sub_stages.yaml" ,
			expectedError:    []string{"Duplicate stage name 'A Working title 2'", "Duplicate stage name 'A Working title'",},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName := "test_data/" + tt.name
			file, err := ioutil.ReadFile( fileName)

			if err != nil {
				println("ERROR: Couldn't read file ", fileName, " with error ",err)
			}

			yaml := string(file)
			parsed, _ := ParseJenkinsfileYaml(yaml)

			error := validateStageNames(parsed)

			for _, expected := range tt.expectedError {
				if ! strings.Contains(error.Error(),expected) {
					t.Fatal("missing  expected error '", expected, "'")
				}
			}
		})
	}
}