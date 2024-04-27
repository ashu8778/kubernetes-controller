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

package controller

import (
	"context"
	"fmt"
	"time"

	mygroupv1 "github.com/ashu8778/kubernetes-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MyCrdReconciler reconciles a MyCrd object
type MyCrdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mygroup.my.domain,resources=mycrds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mygroup.my.domain,resources=mycrds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mygroup.my.domain,resources=mycrds/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyCrd object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *MyCrdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	myCrdResource := &mygroupv1.MyCrd{}

	// Check if new MyCrd resource exists
	err := r.Get(ctx, req.NamespacedName, myCrdResource)
	// If the resource is not found/deleted
	if errors.IsNotFound(err) {
		fmt.Println("Resource not found.")
		return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil
	} else if err != nil {
		// All other errors
		retry := 3
		fmt.Println("Cannot get crd. retrying...")
		if retry > 0 {
			retry--
			return ctrl.Result{}, err
		} else {
			return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil
		}

	}

	// If no errors - Get MyCrd resource returns no error. Proceed to create dependent resource
	for idx := 0; idx < myCrdResource.Spec.PodCount; idx++ {
		// Dependent resource properties
		newPodName := fmt.Sprintf("%v-%v", myCrdResource.Spec.PodName, idx)
		newPodNamespace := myCrdResource.Spec.PodNamespace
		newPodImage := myCrdResource.Spec.ImageName

		// Check if dependent resource already exists.
		err = r.Get(ctx, types.NamespacedName{Name: newPodName, Namespace: newPodNamespace}, &corev1.Pod{})
		if err == nil {
			fmt.Printf("%v exists in %v namespace. It will not be created...\n", newPodName, newPodNamespace)
			return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil
		}
		if err != nil {
			fmt.Printf("%v does not exist in %v namespace. It will be created...\n", newPodName, newPodNamespace)
			// Added owner reference on dependent resource
			blockOwnerDeletion := true
			fmt.Printf("owner ref is ....... %+v=\n", []metav1.OwnerReference{
				{APIVersion: myCrdResource.APIVersion, Kind: myCrdResource.Kind, Name: myCrdResource.Name, BlockOwnerDeletion: &blockOwnerDeletion, UID: myCrdResource.UID},
			})
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      newPodName,
					Namespace: newPodNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{APIVersion: myCrdResource.APIVersion, Kind: myCrdResource.Kind, Name: myCrdResource.Name, BlockOwnerDeletion: &blockOwnerDeletion, UID: myCrdResource.UID},
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  newPodName,
							Image: newPodImage,
						},
					},
				},
			}

			fmt.Println("Creating pod...")
			err = r.Create(ctx, pod)
			if err != nil {
				fmt.Println("Error while creating pod")
				return ctrl.Result{}, err
			} else {
				fmt.Println("Pod created.")
			}

		}
	}

	return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *MyCrdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1.MyCrd{}).
		Complete(r)
}
