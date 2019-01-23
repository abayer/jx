package buildpipeline

import (
	"github.com/google/go-cmp/cmp"
	pipelinev1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	tb "github.com/knative/build-pipeline/test/builder"
	"github.com/knative/pkg/apis"
	"testing"
)

// TODO: Probably move the YAML to external files, like in Declarative's tests, and write a builder for generating the
// expected objects. Because as this is now, there are way too many lines here.
func TestParseJenkinsfileYaml(t *testing.T) {
	tests := []struct {
		name             string
		yaml             string
		expected         *Jenkinsfile
		pipeline         *pipelinev1alpha1.Pipeline
		tasks            []*pipelinev1alpha1.Task
		expectedErrorMsg string
	}{
		{
			name: "simple jenkinsfile",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
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
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd"))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
						tb.ResourceTargetPath("workspace"))),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
			},
		},
		{
			name: "multiple stages",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
  - name: Another stage
    steps:
      - command: echo
        args: ['again']
`,
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
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd"),
				tb.PipelineTask("another-stage", "somepipeline-build-somebuild-stage-another-stage-abcd",
					tb.PipelineTaskResourceDependency("workspace", tb.ProvidedBy("somepipeline-build-somebuild-stage-a-working-stage-abcd"))))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
						tb.ResourceTargetPath("workspace"))),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-another-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-another-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("again")),
				)),
			},
		},
		{
			name: "nested stages",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: Parent Stage
    stages:
      - name: A Working Stage
        steps:
          - command: echo
            args:
              - hello
              - world
      - name: Another stage
        steps:
          - command: echo
            args: ['again']
`,
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
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd"),
				tb.PipelineTask("another-stage", "somepipeline-build-somebuild-stage-another-stage-abcd",
					tb.PipelineTaskResourceDependency("workspace", tb.ProvidedBy("somepipeline-build-somebuild-stage-a-working-stage-abcd"))))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
						tb.ResourceTargetPath("workspace"))),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "world")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-another-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-another-stage-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("again")),
				)),
			},
		},
		{
			name: "parallel stages",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: Parent Stage
    parallel:
      - name: A Working Stage
        steps:
          - command: echo
            args:
              - hello
              - world
      - name: Another stage
        steps:
          - command: echo
            args: ['again']
`,
			expected: &Jenkinsfile{
				APIVersion: "v0.1",
				Agent: Agent{
					Image: "some-image",
				},
				Stages: []Stage{
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
				},
			},
			expectedErrorMsg: "parallel stages not yet implemented for CRD translation",
		},
		{
			name: "environment at top and in stage",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
environment:
  - name: SOME_VAR
    value: A value for the env var
stages:
  - name: A stage with environment
    environment:
        - name: SOME_OTHER_VAR
          value: A value for the other env var
    steps:
      - command: echo
        args: ['hello', '${SOME_OTHER_VAR}']
`,
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
				tb.PipelineTask("a-stage-with-environment", "somepipeline-build-somebuild-stage-a-stage-with-environmen-abcd"))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-stage-with-environmen-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
						tb.ResourceTargetPath("workspace"))),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-a-stage-with-environment-step-0-abcd", "some-image", tb.Command("echo"), tb.Args("hello", "${SOME_OTHER_VAR}"),
						// TODO: Ordering doesn't seem to be deterministic.
						tb.EnvVar("SOME_VAR", "A value for the env var"), tb.EnvVar("SOME_OTHER_VAR", "A value for the other env var")),
				)),
			},
		},
		{
			name: "syntactic sugar step and a command",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
      - step: "some-step"
        options:
          firstParam: "some value"
          secondParam: "some other value"
`,
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
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
    post:
      - condition: success
        actions:
          - name: mail
            options:
              to: foo@bar.com
              subject: "Yay, it passed"
      - condition: failure
        actions:
          - name: slack
            options:
              whatever: the
              slack: config
              actually: "is. =)"
      - condition: always
        actions:
          - name: junit
            options:
              pattern: "target/surefire-reports/**/*.xml"
`,
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
			name: "top-level and stage options",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
options:
  timeout:
    time: 50
    unit: minutes
  retry: 3
stages:
  - name: A Working Stage
    options:
      timeout:
        time: 5
        unit: seconds
      retry: 4
      stash:
        name: Some Files
        files: "somedir/**/*"
      unstash:
        name: Earlier Files
        dir: some/sub/dir
    steps:
      - command: echo
        args:
          - hello
          - world
`,
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
			expectedErrorMsg: "options at top level not yet supported",
		},
		{
			name: "stage and step agent",
			yaml: `apiVersion: v0.1
stages:
  - name: A Working Stage
    agent:
      image: some-image
    steps:
      - command: echo
        args:
          - hello
          - world
        agent:
          image: some-other-image
      - command: echo
        args: ['goodbye']
`,
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
				tb.PipelineTask("a-working-stage", "somepipeline-build-somebuild-stage-a-working-stage-abcd"))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-working-stage-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
						tb.ResourceTargetPath("workspace"))),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-a-working-stage-step-0-abcd", "some-other-image", tb.Command("echo"), tb.Args("hello", "world")),
					tb.Step("stage-a-working-stage-step-1-abcd", "some-image", tb.Command("echo"), tb.Args("goodbye")),
				)),
			},
		},
		{
			name: "mangled task names",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: . -a- .
    steps:
      - command: ls
  - name: Wööh!!!! - This is cool.
    steps:
      - command: ls
`,
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
				tb.PipelineTask(".--a--.", "somepipeline-build-somebuild-stage-a-abcd"),
				tb.PipelineTask("wööh!!!!---this-is-cool.", "somepipeline-build-somebuild-stage-wh-this-is-cool-abcd",
					tb.PipelineTaskResourceDependency("workspace", tb.ProvidedBy("somepipeline-build-somebuild-stage-a-abcd"))))),
			tasks: []*pipelinev1alpha1.Task{
				tb.Task("somepipeline-build-somebuild-stage-a-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit,
						tb.ResourceTargetPath("workspace"))),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-a-step-0-abcd", "some-image", tb.Command("ls")),
				)),
				tb.Task("somepipeline-build-somebuild-stage-wh-this-is-cool-abcd", "somenamespace", tb.TaskSpec(
					tb.TaskInputs(tb.InputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.TaskOutputs(tb.OutputsResource("workspace", pipelinev1alpha1.PipelineResourceTypeGit)),
					tb.Step("stage-wh-this-is-cool-step-0-abcd", "some-image", tb.Command("ls")),
				)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseJenkinsfileYaml(tt.yaml)
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

			if d := cmp.Diff(tt.pipeline, pipeline); d != "" {
				t.Errorf("Generated Pipeline did not match expected: %s", d)
			}

			if d := cmp.Diff(tt.tasks, tasks); d != "" {
				t.Errorf("Generated Tasks did not match expected: %s", d)
			}
		})
	}
}

func TestFailedValidation(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedError *apis.FieldError
	}{
		{
			name: "bad apiVersion",
			yaml: `apiVersion: baaaad
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: &apis.FieldError{
				Message: "Invalid apiVersion format: must be 'v(digits).(digits)",
				Paths:   []string{"apiVersion"},
			},
		},
		/* TODO: Once we figure out how to differentiate between an empty agent and no agent specified...
				   {
				   			name: "empty agent",
				   			yaml: `apiVersion: v0.1
				   agent:
				   stages:
				     - name: A Working Stage
				       steps:
				         - command: echo
				           args:
		          - hello
		          - world
				   `,
				   			expectedError: &apis.FieldError{
				   				Message: "Invalid apiVersion format: must be 'v(digits).(digits)",
				   				Paths:   []string{"apiVersion"},
				   			},
				   		},
		*/
		{
			name: "agent with both image and label",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
  label: some-label
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: apis.ErrMultipleOneOf("label", "image").ViaField("agent"),
		},
		{
			name: "no stages",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
`,
			expectedError: apis.ErrMissingField("stages"),
		},
		{
			name: "no steps, stages, or parallel",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Broken Stage
`,
			expectedError: apis.ErrMissingOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name: "steps and stages",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Broken Stage
    steps:
      - command: echo
        arguments: ["hello","world"]
    stages:
      - name: Nested
        steps:
          - command: echo
            arguments: ['oops']
`,
			expectedError: apis.ErrMultipleOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name: "steps and parallel",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Broken Stage
    steps:
      - command: echo
        arguments: ["hello","world"]
    parallel:
      - name: Nested
        steps:
          - command: echo
            arguments: ['oops']
`,
			expectedError: apis.ErrMultipleOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name: "stages and parallel",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Broken Stage
    stages:
      - name: Nested
        steps:
          - command: echo
            arguments: ['oops']
    parallel:
      - name: Other Nested
        steps:
          - command: echo
            arguments: again
`,
			expectedError: apis.ErrMultipleOneOf("steps", "stages", "parallel").ViaFieldIndex("stages", 0),
		},
		{
			name: "step without command or step",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - args: ['hello','world']
`,
			expectedError: apis.ErrMissingOneOf("command", "step").ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name: "step with both command and step",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - command: echo
        step: some-step
`,
			expectedError: apis.ErrMultipleOneOf("command", "step").ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name: "step with command and options",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - command: echo
        options:
          someOptions: someValue
`,
			expectedError: (&apis.FieldError{
				Message: "Cannot set options for a command",
				Paths:   []string{"options"},
			}).ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name: "step with step and arguments",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    steps:
      - step: some-step
        args: ['some', 'args']
`,
			expectedError: (&apis.FieldError{
				Message: "Cannot set command-line arguments for a step",
				Paths:   []string{"args"},
			}).ViaFieldIndex("steps", 0).ViaFieldIndex("stages", 0),
		},
		{
			name: "no parent or stage agent",
			yaml: `apiVersion: v0.1
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "No agent specified for stage or for its parent(s)",
				Paths:   []string{"agent"},
			}).ViaFieldIndex("stages", 0),
		},
		{
			name: "top level timeout without time",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
options:
  timeout:
    unit: seconds
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options"),
		},
		{
			name: "stage timeout without time",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    options:
      timeout:
        unit: seconds
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "top level timeout with invalid unit",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
options:
  timeout:
    time: 5
    unit: years
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "years is not a valid time unit. Valid time units are seconds, minutes, hours, days",
				Paths:   []string{"unit"},
			}).ViaField("timeout").ViaField("options"),
		},
		{
			name: "stage timeout with invalid unit",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    options:
      timeout:
        time: 5
        unit: years
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "years is not a valid time unit. Valid time units are seconds, minutes, hours, days",
				Paths:   []string{"unit"},
			}).ViaField("timeout").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "top level timeout with invalid time",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
options:
  timeout:
    time: 0
    unit: minutes
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options"),
		},
		{
			name: "stage timeout with invalid time",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    options:
      timeout:
        time: 0
        unit: minutes
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "Timeout must be greater than zero",
				Paths:   []string{"time"},
			}).ViaField("timeout").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "top level retry with invalid count",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
options:
  retry: -5
stages:
  - name: A Working Stage
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "Retry count cannot be negative",
				Paths:   []string{"retry"},
			}).ViaField("options"),
		},
		{
			name: "stage retry with invalid count",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    options:
      retry: -5
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "Retry count cannot be negative",
				Paths:   []string{"retry"},
			}).ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "stash without name",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    options:
      stash:
        files: "foo/**/*"
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "The stash name must be provided",
				Paths:   []string{"name"},
			}).ViaField("stash").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "stash without files",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    options:
      stash:
        name: a-stash
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "files to stash must be provided",
				Paths:   []string{"files"},
			}).ViaField("stash").ViaField("options").ViaFieldIndex("stages", 0),
		},
		{
			name: "unstash without name",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: A Working Stage
    options:
      unstash:
        dir: some/dir
    steps:
      - command: echo
        args:
          - hello
          - world
`,
			expectedError: (&apis.FieldError{
				Message: "The unstash name must be provided",
				Paths:   []string{"name"},
			}).ViaField("unstash").ViaField("options").ViaFieldIndex("stages", 0),
		},
        {
			name: "blank stage name",
			yaml: `apiVersion: v0.1
agent:
  image: some-image
stages:
  - name: .-  ^ ö
    steps:
      - command: ls
`,
            expectedError: (&apis.FieldError{
                Message: "Stage name must contain at least one ASCII letter",
                Paths:   []string{"name"},
            }).ViaFieldIndex("stages", 0),
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, parseErr := ParseJenkinsfileYaml(tt.yaml)
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
		name           string
		input          string
		expected string
	}{
		{
			name: "unmodified",
			input: "unmodified",
			expected: "unmodified-suffix",
		},
		{
			name: "spaces",
			input: "A Simple Test.",
			expected: "a-simple-test-suffix",
		},
		{
			name: "no leading digits",
			input: "0123456789no-leading-digits",
			expected: "no-leading-digits-suffix",
		},
		{
			name: "no leading hyphens",
			input: "----no-leading-hyphens",
			expected: "no-leading-hyphens-suffix",
		},
		{
			name: "no consecutive hyphens",
			input: "no--consecutive- hyphens",
			expected: "no-consecutive-hyphens-suffix",
		},
		{
			name: "no trailing hyphens",
			input: "no-trailing-hyphens----",
			expected: "no-trailing-hyphens-suffix",
		},
		{
			name: "no symbols",
			input: "&$^#@(*&$^-whoops",
			expected: "whoops-suffix",
		},
		{
			name: "no unprintable characters",
			input: "a\n\t\x00b",
			expected: "ab-suffix",
		},
		{
			name: "no unicode",
			input: "japan-日本",
			expected: "japan-suffix",
		},
		{
			name: "no non-bmp characters",
			input: "happy 😃",
			expected: "happy-suffix",
		},
		{
			name: "truncated to 63",
			input:    "a0123456789012345678901234567890123456789012345678901234567890123456789",
			expected: "a0123456789012345678901234567890123456789012345678901234-suffix",
		},
		{
			name: "truncated to 62",
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
