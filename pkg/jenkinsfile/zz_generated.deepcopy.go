// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package jenkinsfile

import (
	syntax "github.com/jenkins-x/jx/pkg/tekton/syntax"
	v1 "k8s.io/api/core/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CreateJenkinsfileArguments) DeepCopyInto(out *CreateJenkinsfileArguments) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CreateJenkinsfileArguments.
func (in *CreateJenkinsfileArguments) DeepCopy() *CreateJenkinsfileArguments {
	if in == nil {
		return nil
	}
	out := new(CreateJenkinsfileArguments)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CreatePipelineArguments) DeepCopyInto(out *CreatePipelineArguments) {
	*out = *in
	if in.Lifecycles != nil {
		in, out := &in.Lifecycles, &out.Lifecycles
		*out = new(PipelineLifecycles)
		(*in).DeepCopyInto(*out)
	}
	if in.PodTemplates != nil {
		in, out := &in.PodTemplates, &out.PodTemplates
		*out = make(map[string]*v1.Pod, len(*in))
		for key, val := range *in {
			var outVal *v1.Pod
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(v1.Pod)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CreatePipelineArguments.
func (in *CreatePipelineArguments) DeepCopy() *CreatePipelineArguments {
	if in == nil {
		return nil
	}
	out := new(CreatePipelineArguments)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImportFile) DeepCopyInto(out *ImportFile) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImportFile.
func (in *ImportFile) DeepCopy() *ImportFile {
	if in == nil {
		return nil
	}
	out := new(ImportFile)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Module) DeepCopyInto(out *Module) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Module.
func (in *Module) DeepCopy() *Module {
	if in == nil {
		return nil
	}
	out := new(Module)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Modules) DeepCopyInto(out *Modules) {
	*out = *in
	if in.Modules != nil {
		in, out := &in.Modules, &out.Modules
		*out = make([]*Module, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Module)
				**out = **in
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Modules.
func (in *Modules) DeepCopy() *Modules {
	if in == nil {
		return nil
	}
	out := new(Modules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamedLifecycle) DeepCopyInto(out *NamedLifecycle) {
	*out = *in
	if in.Lifecycle != nil {
		in, out := &in.Lifecycle, &out.Lifecycle
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamedLifecycle.
func (in *NamedLifecycle) DeepCopy() *NamedLifecycle {
	if in == nil {
		return nil
	}
	out := new(NamedLifecycle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineConfig) DeepCopyInto(out *PipelineConfig) {
	*out = *in
	if in.Extends != nil {
		in, out := &in.Extends, &out.Extends
		*out = new(PipelineExtends)
		**out = **in
	}
	if in.Agent != nil {
		in, out := &in.Agent, &out.Agent
		*out = new(syntax.Agent)
		**out = **in
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Pipelines.DeepCopyInto(&out.Pipelines)
	if in.ContainerOptions != nil {
		in, out := &in.ContainerOptions, &out.ContainerOptions
		*out = new(v1.Container)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineConfig.
func (in *PipelineConfig) DeepCopy() *PipelineConfig {
	if in == nil {
		return nil
	}
	out := new(PipelineConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineExtends) DeepCopyInto(out *PipelineExtends) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineExtends.
func (in *PipelineExtends) DeepCopy() *PipelineExtends {
	if in == nil {
		return nil
	}
	out := new(PipelineExtends)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineLifecycle) DeepCopyInto(out *PipelineLifecycle) {
	*out = *in
	if in.Steps != nil {
		in, out := &in.Steps, &out.Steps
		*out = make([]*syntax.Step, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(syntax.Step)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PreSteps != nil {
		in, out := &in.PreSteps, &out.PreSteps
		*out = make([]*syntax.Step, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(syntax.Step)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineLifecycle.
func (in *PipelineLifecycle) DeepCopy() *PipelineLifecycle {
	if in == nil {
		return nil
	}
	out := new(PipelineLifecycle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in PipelineLifecycleArray) DeepCopyInto(out *PipelineLifecycleArray) {
	{
		in := &in
		*out = make(PipelineLifecycleArray, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineLifecycleArray.
func (in PipelineLifecycleArray) DeepCopy() PipelineLifecycleArray {
	if in == nil {
		return nil
	}
	out := new(PipelineLifecycleArray)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineLifecycles) DeepCopyInto(out *PipelineLifecycles) {
	*out = *in
	if in.Setup != nil {
		in, out := &in.Setup, &out.Setup
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.SetVersion != nil {
		in, out := &in.SetVersion, &out.SetVersion
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.PreBuild != nil {
		in, out := &in.PreBuild, &out.PreBuild
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.Build != nil {
		in, out := &in.Build, &out.Build
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.PostBuild != nil {
		in, out := &in.PostBuild, &out.PostBuild
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.Promote != nil {
		in, out := &in.Promote, &out.Promote
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.Pipeline != nil {
		in, out := &in.Pipeline, &out.Pipeline
		*out = new(syntax.ParsedPipeline)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineLifecycles.
func (in *PipelineLifecycles) DeepCopy() *PipelineLifecycles {
	if in == nil {
		return nil
	}
	out := new(PipelineLifecycles)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pipelines) DeepCopyInto(out *Pipelines) {
	*out = *in
	if in.PullRequest != nil {
		in, out := &in.PullRequest, &out.PullRequest
		*out = new(PipelineLifecycles)
		(*in).DeepCopyInto(*out)
	}
	if in.Release != nil {
		in, out := &in.Release, &out.Release
		*out = new(PipelineLifecycles)
		(*in).DeepCopyInto(*out)
	}
	if in.Feature != nil {
		in, out := &in.Feature, &out.Feature
		*out = new(PipelineLifecycles)
		(*in).DeepCopyInto(*out)
	}
	if in.Post != nil {
		in, out := &in.Post, &out.Post
		*out = new(PipelineLifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.Overrides != nil {
		in, out := &in.Overrides, &out.Overrides
		*out = make([]*syntax.PipelineOverride, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(syntax.PipelineOverride)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Default != nil {
		in, out := &in.Default, &out.Default
		*out = new(syntax.ParsedPipeline)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pipelines.
func (in *Pipelines) DeepCopy() *Pipelines {
	if in == nil {
		return nil
	}
	out := new(Pipelines)
	in.DeepCopyInto(out)
	return out
}
