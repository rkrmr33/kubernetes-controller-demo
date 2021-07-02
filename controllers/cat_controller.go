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

package controllers

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/rkrmr33/kubernetes-controller-demo/api/v1alpha1"
)

const (
	indexKey = ".metadata.controller"
)

var (
	apiGVStr = v1alpha1.GroupVersion.String()
)

// CatReconciler reconciles a Cat object
type CatReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=example.cats.io,resources=cats,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=example.cats.io,resources=cats/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=example.cats.io,resources=cats/finalizers,verbs=update
//+kubebuilder:rbac:groups=v1,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cat object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CatReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)
	logr.V(1).Info("reconciling cat")

	cat := &v1alpha1.Cat{}
	if err := r.Get(ctx, req.NamespacedName, cat); err != nil {
		logr.Error(err, "unable to get cat, ignoring and waiting for requeue")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cat.Status.Phase == v1alpha1.CatPhaseCompleted {
		logr.V(1).Info("cat in completed phase, nothing to do")

		return ctrl.Result{}, nil
	}

	dur, err := time.ParseDuration(string(*cat.Spec.Duration))
	if err != nil {
		logr.Error(err, "failed to parse cat duration", "duration", *cat.Spec.Duration)

		return ctrl.Result{}, err
	}

	if !cat.Status.LastCatPodFinishedTime.IsZero() || time.Now().Add(dur).Before(cat.Status.LastCatPodFinishedTime.Time) {
		schedTime := cat.Status.LastCatPodFinishedTime.Add(dur)
		timeLeft := time.Until(schedTime)
		logr.V(1).Info("still time until next can schedule", "time-left", timeLeft)

		return ctrl.Result{RequeueAfter: timeLeft}, nil // schedule again when it's time to start the next life
	}

	if cat.Status.LastCatPodName != "" {
		lastCatPod := &v1.Pod{}
		namespacedName := types.NamespacedName{
			Namespace: cat.Namespace,
			Name:      cat.Status.LastCatPodName,
		}
		if err := r.Get(ctx, namespacedName, lastCatPod); err != nil {
			logr.Error(err, "unable to get last cat pod")

			return ctrl.Result{}, err
		}

		if err := r.handleLastPod(ctx, cat, lastCatPod); err != nil {
			logr.Error(err, "unable to delete last cat pod")

			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}

	if cat.Status.CurrentLife == *cat.Spec.TotalLives {
		logr.V(1).Info("cat finished all its lives")
		cat.Status.Phase = v1alpha1.CatPhaseCompleted

		if err := r.Status().Update(ctx, cat); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	cat.Status.Phase = v1alpha1.CatPhaseRunning

	if cat.Status.LastCatPodPhase == v1.PodSucceeded || cat.Status.LastCatPodPhase == v1.PodFailed || cat.Status.LastCatPodPhase == "" {
		logr.V(1).Info("scheduling new cat pod")
		if err := r.spawnNewCatPod(ctx, cat); err != nil {
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}

	return ctrl.Result{}, r.Status().Update(ctx, cat)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CatReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1.Pod{}, indexKey, func(o client.Object) []string {
		p := o.(*v1.Pod)
		owner := metav1.GetControllerOf(p)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != apiGVStr || owner.Kind != "Cat" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Cat{}).
		Owns(&v1.Pod{}).
		Complete(r)
}

func (r *CatReconciler) handleLastPod(ctx context.Context, cat *v1alpha1.Cat, lastPod *v1.Pod) error {
	if lastPod.Status.Phase == cat.Status.LastCatPodPhase {
		return nil // ignore same state
	}

	switch lastPod.Status.Phase {
	case v1.PodFailed:
		cat.Status.Phase = v1alpha1.CatPhaseError
		cat.Status.LastCatPodFinishedTime = metav1.Now()
		cat.Status.Message = fmt.Sprintf("pod %s has failed with: %s", lastPod.Name, lastPod.Status.Message)
		return r.Delete(ctx, lastPod)
	case v1.PodSucceeded:
		cat.Status.CurrentLife++
		cat.Status.LastCatPodFinishedTime = metav1.Now()
		cat.Status.Message = fmt.Sprintf("life %d pod: %s finished successfully", cat.Status.CurrentLife, cat.Status.LastCatPodName)
		return r.Delete(ctx, lastPod)
	case v1.PodRunning:
		cat.Status.Message = fmt.Sprintf("life %d running pod: %s", cat.Status.CurrentLife, cat.Status.LastCatPodName)
	}

	return nil
}

func (r *CatReconciler) spawnNewCatPod(ctx context.Context, cat *v1alpha1.Cat) error {
	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-life-%d", cat.Name, cat.Status.CurrentLife),
			Namespace: cat.Namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "main",
					Image: "alpine:latest",
					Command: []string{
						"sh",
						"-c",
					},
					Args: []string{
						fmt.Sprintf("echo \"%s\"", *cat.Spec.Message),
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	cat.Status.LastCatPodName = p.Name
	cat.Status.LastCatPodPhase = v1.PodUnknown

	return r.Create(ctx, p)
}
