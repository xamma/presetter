/*
Copyright 2024 Max Bickel.

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourcePresetSpec defines the desired state of ResourcePreset
type ResourcePresetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	CPURequests    resource.Quantity `json:"cpuRequests"`
	CPULimits      resource.Quantity `json:"cpuLimits"`
	MemoryRequests resource.Quantity `json:"memoryRequests"`
	MemoryLimits   resource.Quantity `json:"memoryLimits"`
}

// ResourcePresetStatus defines the observed state of ResourcePreset
type ResourcePresetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ResourcePreset is the Schema for the resourcepresets API
type ResourcePreset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourcePresetSpec   `json:"spec,omitempty"`
	Status ResourcePresetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourcePresetList contains a list of ResourcePreset
type ResourcePresetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourcePreset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourcePreset{}, &ResourcePresetList{})
}
