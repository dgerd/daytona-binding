/*
Copyright 2018 The Knative Authors

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
)

// +genduck

// PodSpecable is implemented by types containing a PodTemplateSpec
// in the manner of ReplicaSet, Deployment, DaemonSet, StatefulSet.
type Podable corev1.PodSpec

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WithPod is the shell that demonstrates how PodSpecable types wrap
// a PodSpec.
type WithPodable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Podable `json:"spec,omitempty"`
}

// Assert that we implement the interfaces necessary to
// use duck.VerifyType.
var (
	_ duck.Populatable   = (*WithPodable)(nil)
	_ duck.Implementable = (*Podable)(nil)
	_ apis.Listable      = (*WithPodable)(nil)
)

// GetFullType implements duck.Implementable
func (*Podable) GetFullType() duck.Populatable {
	return &WithPodable{}
}

// Populate implements duck.Populatable
func (t *WithPodable) Populate() {
	spec := Podable{
			Containers: []corev1.Container{{
				Name:  "container-name",
				Image: "container-image:latest",
			}},
		}
	t.Spec = spec
}

// GetListType implements apis.Listable
func (*WithPodable) GetListType() runtime.Object {
	return &WithPodableList{}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WithPodableList is a list of WithPod resources
type WithPodableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WithPodable `json:"items"`
}
