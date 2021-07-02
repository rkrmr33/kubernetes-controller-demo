/*
Copyright 2021.

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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DurationString string

type CatPhase string

const (
	// CatPhaseRunning indicates that the cat is currently alive
	CatPhaseRunning CatPhase = "Alive"

	// CatPhaseCompleted indicates that the cat finished running
	CatPhaseCompleted

	// CatPhaseError indicates that the cat is at some error state
	CatPhaseError
)

// CatSpec defines the desired state of Cat
type CatSpec struct {

	// Most primitive go types may be used with the exception of numbers (only int32, int64 for integers, and resource.Quantity, for decimals)
	// When an primitive type is optional you should probably make it of pointer type to distinguish zero value and nil.

	// +kubebuilder:validation:Minimum=0
	// +optional

	// TotalLives is the total number of times a cat pod will be created.
	// Default is: 9
	TotalLives *int32 `json:"totalLives,omitempty"`

	// +kubebuilder:validation:MaxLength=128
	// +optional

	// Message is what the cat would say. Default is "hello, world!"
	Message *string `json:"message,omitempty"`

	// +optional

	// Duration the total duration per life. Default is 5s
	Duration *DurationString `json:"duration,omitempty"`
}

// CatStatus defines the observed state of Cat
type CatStatus struct {
	// LastCatPodName is the name of the last pod that was created by this cat
	LastCatPodName string `json:"lastCatPodName,omitempty"`

	// LastCatPodFinishedTime the time the last cat pod completed
	LastCatPodFinishedTime metav1.Time `json:"lastCatPodFinishedTime,omitempty"`

	// LastCatPodPhase last pod phase
	LastCatPodPhase v1.PodPhase `json:"lastCatPodPhase,omitempty"`

	// Phase the current phase of the cat
	Phase CatPhase `json:"phase,omitempty"`

	// Message is a description of the current phase
	Message string `json:"message,omitempty"`

	// CurrentLife the current life number
	CurrentLife int32 `json:"currentLife,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cat is the Schema for the cats API
type Cat struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatSpec   `json:"spec,omitempty"`
	Status CatStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CatList contains a list of Cat
type CatList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cat `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cat{}, &CatList{})
}
