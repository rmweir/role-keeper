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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SubjectRoleRequestSpec defines the desired state of SubjectRoleRequest
type SubjectRoleRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SubjectID   string    `json:"subjectID,omitempty"`
	SubjectKind string    `json:"subjectKind,omitempty"`
	Operation   Operation `json:"operation,omitempty"`
}

// +kubebuilder:validation:Enum=AddRole;RemoveRole

type Operation string

const (
	AddRole    Operation = "AddRole"
	RemoveRole Operation = "RemoveRole"
)

// SubjectRoleRequestStatus defines the observed state of SubjectRoleRequest
type SubjectRoleRequestStatus struct {
	// +optional
	Status RequestStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:Enum=Successful;Failure

type RequestStatus string

const (
	Success RequestStatus = "Successful"
	Failure RequestStatus = "Failure"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SubjectRoleRequest is the Schema for the subjectrolerequests API
type SubjectRoleRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubjectRoleRequestSpec   `json:"spec,omitempty"`
	Status SubjectRoleRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SubjectRoleRequestList contains a list of SubjectRoleRequest
type SubjectRoleRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubjectRoleRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubjectRoleRequest{}, &SubjectRoleRequestList{})
}
