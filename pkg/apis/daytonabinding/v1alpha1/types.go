/*
Copyright 2020 The Knative Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DaytonaBinding is a Knative-style Binding for injecting Github credentials
// compatible with ./pkg/github into any Kubernetes resource with a Pod Spec.
type DaytonaBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the DaytonaBinding (from the client).
	// +optional
	Spec DaytonaBindingSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the DaytonaBinding (from the controller).
	// +optional
	Status DaytonaBindingStatus `json:"status,omitempty"`
}

var (
	// Check that DaytonaBinding can be validated and defaulted.
	_ apis.Validatable   = (*DaytonaBinding)(nil)
	_ apis.Defaultable   = (*DaytonaBinding)(nil)
	_ kmeta.OwnerRefable = (*DaytonaBinding)(nil)
)

// DaytonaBindingSpec holds the desired state of the DaytonaBinding (from the client).
type DaytonaBindingSpec struct {
	// Subject holds a reference to the "pod speccable" Kubernetes resource which will
	// be bound with Github secret data.
	Subject tracker.Reference `json:"subject"`

	// The location of the Daytona image
	Image string `json:"image"`

	Auth string `json:"auth"`

	AuthMount string `json:"authMount"`

	SecretEnv string `json:"secretEnv"`

	TokenPath string `json:"tokenPath"`

	VaultAuthRole string `json:"vaultAuthRole"`

	SecretPath string `json:"secretPath"`

	VaultSecretsApp string `json:"vaultSecretsApp"`

	VaultSecretsGlobal string `json:"vaultSecretsGlobal"`
	
}

// DaytonaBindingStatus communicates the observed state of the DaytonaBinding (from the controller).
type DaytonaBindingStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DaytonaBindingList is a list of DaytonaBinding resources
type DaytonaBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DaytonaBinding `json:"items"`
}