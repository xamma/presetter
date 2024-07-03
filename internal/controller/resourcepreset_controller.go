package controller

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	presetterv1 "github.com/xamma/presetter/api/v1"
	corev1 "k8s.io/api/core/v1"
)

// ResourcePresetReconciler reconciles a ResourcePreset object
type ResourcePresetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ResourcePresetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Pod
	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if errors.IsNotFound(err) {
			// If the Pod is not found, ignore the error
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		// Log other errors
		logger.Error(err, "Unable to fetch Pod.")
		return ctrl.Result{}, err
	}

	// Check if the Pod has the correct label
	presetName, ok := pod.Labels["presetter.xamma.dev/preset"]
	if !ok {
		// If the label is not found, do nothing
		return ctrl.Result{}, nil
	}

	// Fetch the ResourcePreset from the same namespace as the Pod
	preset := &presetterv1.ResourcePreset{}
	if err := r.Get(ctx, types.NamespacedName{Name: presetName, Namespace: req.Namespace}, preset); err != nil {
		// If the ResourcePreset is not found, log the error but don't return it
		logger.Error(err, "Unable to fetch ResourcePreset.")
		return ctrl.Result{}, err
	}

	// Define the name for the new Pod
	newPodName := fmt.Sprintf("%s-new", pod.Name)

	// Check if the new Pod already exists
	newPod := &corev1.Pod{}
	if err := r.Get(ctx, types.NamespacedName{Name: newPodName, Namespace: req.Namespace}, newPod); err == nil {
		// New Pod already exists, nothing more to do
		return ctrl.Result{}, nil
	} else if !errors.IsNotFound(err) {
		// Some other error occurred while checking for the new Pod
		logger.Error(err, "Unable to check if new Pod exists.")
		return ctrl.Result{}, err
	}

	// Create a new Pod with the updated ResourcePreset
	newPod = pod.DeepCopy() // Create a copy of the existing Pod
	r.applyPresetToPod(newPod, preset)
	newPod.Name = newPodName                                       // Give the new Pod a unique name
	newPod.ResourceVersion = ""                                    // Clear the resource version to create a new Pod
	delete(newPod.ObjectMeta.Labels, "presetter.xamma.dev/preset") // Remove the label to avoid conflicts

	// Create the new Pod
	if err := r.Create(ctx, newPod); err != nil {
		logger.Error(err, "Unable to create new Pod with ResourcePreset.")
		return ctrl.Result{}, err
	}

	// Wait for a short time to allow the new Pod to be fully created before deleting the old Pod
	// Note: This is a simple delay. In production, you might want to use more sophisticated checks.
	// time.Sleep(30 * time.Second)

	// Delete the old Pod
	if err := r.Delete(ctx, pod); err != nil {
		logger.Error(err, "Unable to delete old Pod.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourcePresetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}). // Ensure we're watching Pods
		Complete(r)
}

// applyPresetToPod applies the ResourcePreset to the Pod.
func (r *ResourcePresetReconciler) applyPresetToPod(pod *corev1.Pod, preset *presetterv1.ResourcePreset) {
	if pod.Spec.Containers == nil || len(pod.Spec.Containers) == 0 {
		return // No containers in the pod
	}

	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]

		// Apply CPU and Memory Requests
		if container.Resources.Requests == nil {
			container.Resources.Requests = corev1.ResourceList{}
		}
		container.Resources.Requests[corev1.ResourceCPU] = preset.Spec.CPURequests
		container.Resources.Requests[corev1.ResourceMemory] = preset.Spec.MemoryRequests

		// Apply CPU and Memory Limits
		if container.Resources.Limits == nil {
			container.Resources.Limits = corev1.ResourceList{}
		}
		container.Resources.Limits[corev1.ResourceCPU] = preset.Spec.CPULimits
		container.Resources.Limits[corev1.ResourceMemory] = preset.Spec.MemoryLimits
	}
}
