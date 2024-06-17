//go:build !ignore_autogenerated

/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Broom) DeepCopyInto(out *Broom) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Broom.
func (in *Broom) DeepCopy() *Broom {
	if in == nil {
		return nil
	}
	out := new(Broom)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Broom) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BroomAdjustment) DeepCopyInto(out *BroomAdjustment) {
	*out = *in
	out.MaxLimit = in.MaxLimit.DeepCopy()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BroomAdjustment.
func (in *BroomAdjustment) DeepCopy() *BroomAdjustment {
	if in == nil {
		return nil
	}
	out := new(BroomAdjustment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BroomList) DeepCopyInto(out *BroomList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Broom, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BroomList.
func (in *BroomList) DeepCopy() *BroomList {
	if in == nil {
		return nil
	}
	out := new(BroomList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BroomList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BroomSlackWebhook) DeepCopyInto(out *BroomSlackWebhook) {
	*out = *in
	out.Secret = in.Secret
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BroomSlackWebhook.
func (in *BroomSlackWebhook) DeepCopy() *BroomSlackWebhook {
	if in == nil {
		return nil
	}
	out := new(BroomSlackWebhook)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BroomSlackWebhookSecret) DeepCopyInto(out *BroomSlackWebhookSecret) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BroomSlackWebhookSecret.
func (in *BroomSlackWebhookSecret) DeepCopy() *BroomSlackWebhookSecret {
	if in == nil {
		return nil
	}
	out := new(BroomSlackWebhookSecret)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BroomSpec) DeepCopyInto(out *BroomSpec) {
	*out = *in
	in.Target.DeepCopyInto(&out.Target)
	in.Adjustment.DeepCopyInto(&out.Adjustment)
	out.SlackWebhook = in.SlackWebhook
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BroomSpec.
func (in *BroomSpec) DeepCopy() *BroomSpec {
	if in == nil {
		return nil
	}
	out := new(BroomSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BroomStatus) DeepCopyInto(out *BroomStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BroomStatus.
func (in *BroomStatus) DeepCopy() *BroomStatus {
	if in == nil {
		return nil
	}
	out := new(BroomStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BroomTarget) DeepCopyInto(out *BroomTarget) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BroomTarget.
func (in *BroomTarget) DeepCopy() *BroomTarget {
	if in == nil {
		return nil
	}
	out := new(BroomTarget)
	in.DeepCopyInto(out)
	return out
}
