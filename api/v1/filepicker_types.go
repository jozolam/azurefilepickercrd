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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FilePickerSpec defines the desired state of FilePicker
type FilePickerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	FileName  string `json:"file,omitempty"`
	Container string `json:"container,omitempty"`
	Account   string `json:"account,omitempty"`
}

// FilePickerStatus defines the observed state of FilePicker
type FilePickerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ReconciledAt int64 `json:"reconciledAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FilePicker is the Schema for the filepickers API
type FilePicker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FilePickerSpec   `json:"spec,omitempty"`
	Status FilePickerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FilePickerList contains a list of FilePicker
type FilePickerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FilePicker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FilePicker{}, &FilePickerList{})
}
