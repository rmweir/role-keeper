//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023.

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

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppliedRule) DeepCopyInto(out *AppliedRule) {
	*out = *in
	in.PolicyRule.DeepCopyInto(&out.PolicyRule)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppliedRule.
func (in *AppliedRule) DeepCopy() *AppliedRule {
	if in == nil {
		return nil
	}
	out := new(AppliedRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRegistrar) DeepCopyInto(out *SubjectRegistrar) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRegistrar.
func (in *SubjectRegistrar) DeepCopy() *SubjectRegistrar {
	if in == nil {
		return nil
	}
	out := new(SubjectRegistrar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubjectRegistrar) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRegistrarList) DeepCopyInto(out *SubjectRegistrarList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SubjectRegistrar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRegistrarList.
func (in *SubjectRegistrarList) DeepCopy() *SubjectRegistrarList {
	if in == nil {
		return nil
	}
	out := new(SubjectRegistrarList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubjectRegistrarList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRegistrarSpec) DeepCopyInto(out *SubjectRegistrarSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRegistrarSpec.
func (in *SubjectRegistrarSpec) DeepCopy() *SubjectRegistrarSpec {
	if in == nil {
		return nil
	}
	out := new(SubjectRegistrarSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRegistrarStatus) DeepCopyInto(out *SubjectRegistrarStatus) {
	*out = *in
	if in.AppliedRoles != nil {
		in, out := &in.AppliedRoles, &out.AppliedRoles
		*out = make(map[string]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AppliedRules != nil {
		in, out := &in.AppliedRules, &out.AppliedRules
		*out = make([]AppliedRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AddQueue != nil {
		in, out := &in.AddQueue, &out.AddQueue
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RemoveQueue != nil {
		in, out := &in.RemoveQueue, &out.RemoveQueue
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRegistrarStatus.
func (in *SubjectRegistrarStatus) DeepCopy() *SubjectRegistrarStatus {
	if in == nil {
		return nil
	}
	out := new(SubjectRegistrarStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRoleRequest) DeepCopyInto(out *SubjectRoleRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRoleRequest.
func (in *SubjectRoleRequest) DeepCopy() *SubjectRoleRequest {
	if in == nil {
		return nil
	}
	out := new(SubjectRoleRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubjectRoleRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRoleRequestList) DeepCopyInto(out *SubjectRoleRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SubjectRoleRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRoleRequestList.
func (in *SubjectRoleRequestList) DeepCopy() *SubjectRoleRequestList {
	if in == nil {
		return nil
	}
	out := new(SubjectRoleRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubjectRoleRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRoleRequestSpec) DeepCopyInto(out *SubjectRoleRequestSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRoleRequestSpec.
func (in *SubjectRoleRequestSpec) DeepCopy() *SubjectRoleRequestSpec {
	if in == nil {
		return nil
	}
	out := new(SubjectRoleRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectRoleRequestStatus) DeepCopyInto(out *SubjectRoleRequestStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectRoleRequestStatus.
func (in *SubjectRoleRequestStatus) DeepCopy() *SubjectRoleRequestStatus {
	if in == nil {
		return nil
	}
	out := new(SubjectRoleRequestStatus)
	in.DeepCopyInto(out)
	return out
}
