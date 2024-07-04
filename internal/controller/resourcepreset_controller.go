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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	presetterv1 "github.com/xamma/presetter/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ResourcePresetReconciler reconciles a ResourcePreset object
type ResourcePresetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;watch;get;create;update;patch;delete
// +kubebuilder:rbac:groups=presetter.xamma.dev,resources=resourcepresets,verbs=list;watch;get;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=list;watch;get

// Reconcile is part of the main Kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ResourcePresetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch deployment
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		// not found
		if errors.IsNotFound(err) {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		// other stuff lol
		logger.Error(err, "Unable to fetch Deployment.")
		return ctrl.Result{}, err
	}

	// get presetname from the label
	presetName, ok := deployment.Labels["presetter.xamma.dev/preset"]
	if !ok {
		return ctrl.Result{}, nil
	}

	// Fetch resourcepreset
	preset := &presetterv1.ResourcePreset{}
	if err := r.Get(ctx, types.NamespacedName{Name: presetName, Namespace: deployment.Namespace}, preset); err != nil {
		logger.Error(err, "Unable to fetch ResourcePreset.")
		return ctrl.Result{}, err
	}

	// Update deploy pod template
	updated := false
	for i := range deployment.Spec.Template.Spec.Containers {
		container := &deployment.Spec.Template.Spec.Containers[i]

		// Apply CPU and Memory Requests
		if container.Resources.Requests == nil {
			container.Resources.Requests = corev1.ResourceList{}
		}
		if _, exists := container.Resources.Requests[corev1.ResourceCPU]; !exists {
			container.Resources.Requests[corev1.ResourceCPU] = preset.Spec.CPURequests
		}
		if _, exists := container.Resources.Requests[corev1.ResourceMemory]; !exists {
			container.Resources.Requests[corev1.ResourceMemory] = preset.Spec.MemoryRequests
		}

		// Apply CPU and Memory Limits
		if container.Resources.Limits == nil {
			container.Resources.Limits = corev1.ResourceList{}
		}
		if _, exists := container.Resources.Limits[corev1.ResourceCPU]; !exists {
			container.Resources.Limits[corev1.ResourceCPU] = preset.Spec.CPULimits
		}
		if _, exists := container.Resources.Limits[corev1.ResourceMemory]; !exists {
			container.Resources.Limits[corev1.ResourceMemory] = preset.Spec.MemoryLimits
		}

		updated = true
	}

	// controller conflicts...
	if updated {
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			err := r.Update(ctx, deployment)
			if err == nil {
				break
			}
			if errors.IsConflict(err) {
				logger.Info("Conflict detected, retrying...")
				if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
					logger.Error(err, "Unable to re-fetch Deployment.")
					return ctrl.Result{}, err
				}
				continue
			}
			// whatever..
			logger.Error(err, "Unable to update Deployment with ResourcePreset.")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourcePresetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
