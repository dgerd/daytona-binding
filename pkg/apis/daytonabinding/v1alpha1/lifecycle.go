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

	"knative.dev/pkg/ptr"

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

// InitializeConditions initializes Ready and subconditions to Unknown.
func (dbs *DaytonaBindingStatus) InitializeConditions() {
	daytonaCondSet.Manage(dbs).InitializeConditions()
}

// MarkBindingUnavailable marks when the DaytonaBinding CRD is not Ready with a reason.
func (dbs *DaytonaBindingStatus) MarkBindingUnavailable(reason, message string) {
	daytonaCondSet.Manage(dbs).MarkFalse(
		DaytonaBindingConditionReady, reason, message)
}

// MarkBindingAvailable marks when the DaytonaBinding CRD is Ready.
func (dbs *DaytonaBindingStatus) MarkBindingAvailable() {
	daytonaCondSet.Manage(dbs).MarkTrue(DaytonaBindingConditionReady)
}

// Do implements the logic of injecting all of the Daytona content into the Pod.
func (db *DaytonaBinding) Do(ctx context.Context, pod *duckv1.WithPodable) {
	// First undo so that we can just unconditionally append below.
	db.Undo(ctx, pod)

	// Add daytona secrets volume.
	volume := corev1.Volume{
		Name: daytona.SecretVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: daytona.Medium,
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)

	volumeMount := []corev1.VolumeMount{{
		Name:      daytona.SecretVolumeName,
		MountPath: db.Spec.MountPath,
	}}
	// Add daytona to the init containers section.
	container := corev1.Container{
		Name: daytona.ContainerName,
		Env:  daytonaEnv(db),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                ptr.Int64(daytona.RunAsUser),
			AllowPrivilegeEscalation: ptr.Bool(false),
		},
		VolumeMounts: volumeMount,
		Image:        db.Spec.Image,
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, container)

	// Add volume mount to the user container. As users can customize the container name
	// and sidecars can vary we look this up by the presence of the `K_REVISION` Environment
	// Variable. This is a hack, but works.
	for i, c := range pod.Spec.Containers {
		for _, e := range c.Env {
			// This container is the user container. Modify and then exit.
			if e.Name == "K_REVISION" {
				pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, volumeMount[0])
				return
			}
		}
	}
}

// Undo implements the logic of removing all of the Daytona content from the Pod.
func (db *DaytonaBinding) Undo(ctx context.Context, pod *duckv1.WithPodable) {

	// Remove Daytona Volume
	for i, v := range pod.Spec.Volumes {
		if v.Name == daytona.SecretVolumeName {
			pod.Spec.Volumes = append(pod.Spec.Volumes[:i], pod.Spec.Volumes[i+1:]...)
			break
		}
	}

	// Remove Daytona InitContainer
	for i, c := range pod.Spec.InitContainers {
		if c.Name == daytona.ContainerName {
			pod.Spec.InitContainers = append(pod.Spec.InitContainers[:i], pod.Spec.InitContainers[i+1:]...)
			break
		}
	}

	// Remove Volume from user container
	for i, c := range pod.Spec.Containers {
		for j, e := range c.Env {
			// This container is the user container. Remove and then exit.
			if e.Name == "K_REVISION" {
				pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts[:j], pod.Spec.Containers[i].VolumeMounts[j+1:]...)
				return
			}
		}
	}
}

func daytonaEnv(db *DaytonaBinding) []corev1.EnvVar {
	vars := []corev1.EnvVar{
		// always use a Kubernetes Auth Backend
		{
			Name:  "K8S_AUTH",
			Value: "true",
		},
		// auth mount/name is required
		{
			Name:  "K8S_AUTH_MOUNT",
			Value: db.Spec.AuthMount,
		},
	}
	// optional vars
	if db.Spec.AuthRole != "" {
		vars = append(vars, corev1.EnvVar{
			Name:  "VAULT_AUTH_ROLE",
			Value: db.Spec.AuthRole,
		})
	}
	if db.Spec.TokenPath != "" {
		vars = append(vars, corev1.EnvVar{
			Name:  "TOKEN_PATH",
			Value: db.Spec.TokenPath,
		})
	}
	if db.Spec.SecretPath != "" {
		vars = append(vars, corev1.EnvVar{
			Name:  "SECRET_PATH",
			Value: db.Spec.SecretPath,
		})
	}
	if db.Spec.VaultSecretsApp != "" {
		vars = append(vars, corev1.EnvVar{
			Name:  "VAULT_SECRETS_APP",
			Value: db.Spec.VaultSecretsApp,
		})
	}
	if db.Spec.VaultSecretsGlobal != "" {
		vars = append(vars, corev1.EnvVar{
			Name:  "VAULT_SECRETS_GLOBAL",
			Value: db.Spec.VaultSecretsGlobal,
		})
	}
	return vars
}
