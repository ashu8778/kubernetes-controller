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

// MyCrdSpec defines the desired state of MyCrd
type MyCrdSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MyCrd. Edit mycrd_types.go to remove/update
	PodName      string `json:"podName"`
	ImageName    string `json:"imageName"`
	PodNamespace string `json:"podNamespace,omitempty"`
	PodCount     int    `json:"podCount"`
}

// MyCrdStatus defines the observed state of MyCrd
type MyCrdStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions    []MyCrdCondition `json:"conditions,omitempty"`
	AvailablePods int              `json:"availablePods,omitempty"`
}

// MyCrdCondition represents a condition of the MyCrd
type MyCrdCondition struct {
	Status string `json:"state,omitempty"`
	Reason string `json:"reason,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[-1].state`

// MyCrd is the Schema for the mycrds API
type MyCrd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyCrdSpec   `json:"spec,omitempty"`
	Status MyCrdStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MyCrdList contains a list of MyCrd
type MyCrdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyCrd `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyCrd{}, &MyCrdList{})
}
