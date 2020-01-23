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
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"

	"github.com/dgerd/daytona-binding/pkg/daytona"
)

const (
	// DaytonaBindingConditionReady is set when the binding has been applied to the subjects.
	DaytonaBindingConditionReady = apis.ConditionReady
)

var daytonaCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (db *DaytonaBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("DaytonaBinding")
}

// GetSubject implements Bindable
func (db *DaytonaBinding) GetSubject() tracker.Reference {
	return db.Spec.Subject
}

// GetBindingStatus implements Bindable
func (db *DaytonaBinding) GetBindingStatus() duck.BindableStatus {
	return &db.Status
}

// SetObservedGeneration implements BindableStatus
func (dbs *DaytonaBindingStatus) SetObservedGeneration(gen int64) {
	dbs.ObservedGeneration = gen
}

func (dbs *DaytonaBindingStatus) InitializeConditions() {
	daytonaCondSet.Manage(dbs).InitializeConditions()
}

func (dbs *DaytonaBindingStatus) MarkBindingUnavailable(reason, message string) {
	daytonaCondSet.Manage(dbs).MarkFalse(
		DaytonaBindingConditionReady, reason, message)
}

func (dbs *DaytonaBindingStatus) MarkBindingAvailable() {
	daytonaCondSet.Manage(dbs).MarkTrue(DaytonaBindingConditionReady)
}

func (db *DaytonaBinding) Do(ctx context.Context, ps *duckv1.WithPod) {

	// First undo so that we can just unconditionally append below.
	db.Undo(ctx, ps)

	// Make sure the PodSpec has a Volume like this:
	volume := corev1.Volume{
		Name: github.VolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: db.Spec.Secret.Name,
			},
		},
	}
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, volume)

	// Make sure that each [init]container in the PodSpec has a VolumeMount like this:
	volumeMount := corev1.VolumeMount{
		Name:      github.VolumeName,
		ReadOnly:  true,
		MountPath: github.MountPath,
	}
	spec := ps.Spec.Template.Spec
	for i := range spec.InitContainers {
		spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts, volumeMount)
	}
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMount)
	}
}

func (db *DaytonaBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	// Make sure the PodSpec does NOT have the github volume.
	for i, v := range spec.Volumes {
		if v.Name == github.VolumeName {
			ps.Spec.Template.Spec.Volumes = append(spec.Volumes[:i], spec.Volumes[i+1:]...)
			break
		}
	}

	// Make sure that none of the [init]containers have the github volume mount
	for i, c := range spec.InitContainers {
		for j, vm := range c.VolumeMounts {
			if vm.Name == github.VolumeName {
				spec.InitContainers[i].VolumeMounts = append(c.VolumeMounts[:j], c.VolumeMounts[j+1:]...)
				break
			}
		}
	}
	for i, c := range spec.Containers {
		for j, vm := range c.VolumeMounts {
			if vm.Name == github.VolumeName {
				spec.Containers[i].VolumeMounts = append(c.VolumeMounts[:j], c.VolumeMounts[j+1:]...)
				break
			}
		}
	}
}
