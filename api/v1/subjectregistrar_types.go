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
	v1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SubjectRegistrarSpec defines the desired state of SubjectRegistrar
type SubjectRegistrarSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SubjectRegistrar. Edit subjectregistrar_types.go to remove/update
	SubjectID   string `json:"subjectID,omitempty"`
	SubjectKind string `json:"subjectKind,omitempty"`
}

type AppliedRule struct {
	Namespace string  `json:"namespace,omitempty"`
	Rule      v1.Rule `json:"rule,omitempty"`
}

// SubjectRegistrarStatus defines the observed state of SubjectRegistrar
type SubjectRegistrarStatus struct {
	AppliedRoles map[string]int `json:"appliedRoles,omitempty"`
	AppliedRules []AppliedRule  `json:"appliedRules,omitempty"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SubjectRegistrar is the Schema for the subjectregistrars API
type SubjectRegistrar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubjectRegistrarSpec   `json:"spec,omitempty"`
	Status SubjectRegistrarStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SubjectRegistrarList contains a list of SubjectRegistrar
type SubjectRegistrarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubjectRegistrar `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubjectRegistrar{}, &SubjectRegistrarList{})
}
