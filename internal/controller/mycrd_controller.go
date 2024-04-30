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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	// log := log.FromContext(ctx)
	// log.Info("Reconciliation triggered for resource", "Name", req.Name, "Namespace", req.Namespace)

	myCrdResource := &mygroupv1.MyCrd{}

	fmt.Printf("NamespacedName is : %+v\n", req.NamespacedName)
	// Check if new MyCrd resource exists

	err := r.Get(ctx, req.NamespacedName, myCrdResource)
	// If the resource is not found/deleted
	if k8serrors.IsNotFound(err) {
		fmt.Println("Resource not found.")
		return ctrl.Result{}, nil
	} else if err != nil {
		// All other errors
		fmt.Println("Something went wrong")
		return ctrl.Result{}, err
	}

	myCrdResource.Status.AvailablePods = 0

	err = r.createPods(ctx, myCrdResource)

	if err != nil {
		fmt.Println("Received error from createPods function.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true, RequeueAfter: 8 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyCrdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1.MyCrd{}).
		Complete(r)
}

// Added function to the reconciler to create dependent pods according to the MyCrd specs
func (r *MyCrdReconciler) createPods(ctx context.Context, myCrdResource *mygroupv1.MyCrd) error {

	for idx := 0; idx < myCrdResource.Spec.PodCount; idx++ {

		// Dependent resource properties
		newPodName := fmt.Sprintf("%v-%v", myCrdResource.Spec.PodName, idx)
		newPodNamespace := myCrdResource.Spec.PodNamespace
		newPodImage := myCrdResource.Spec.ImageName
		newPod := &corev1.Pod{}

		// Check if dependent resource already exists.
		err := r.Get(ctx, types.NamespacedName{Name: newPodName, Namespace: newPodNamespace}, newPod)

		// Dependent resource does not exists.
		if k8serrors.IsNotFound(err) {

			fmt.Printf("%v does not exist in %v namespace. It will be created. \n", newPodName, newPodNamespace)

			// Pod does not exists in namespace - create

			// Pod specs
			blockOwnerDeletion := true
			fmt.Printf("owner ref is: %+v=\n", []metav1.OwnerReference{
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

			// Call create api
			fmt.Println("Calling create api.")
			err = r.Create(ctx, pod)

			if err != nil {
				fmt.Println("Error in create api - pod")
				return err
			}
			return nil
		} else if err != nil {
			fmt.Println("Something went wrong while getting pod.")
			return err
		}

		// if pods are existing - cannot create pod
		fmt.Printf("%v exists in %v namespace. It will not be created.\n", newPodName, newPodNamespace)

		if r.isOwner(myCrdResource, newPod) {
			myCrdResource.Status.AvailablePods++
		}
	}

	// Status is success if available pods count is same as spec PodCount
	if myCrdResource.Status.AvailablePods == myCrdResource.Spec.PodCount {
		// Check if no spec conditions
		if myCrdResource.Status.Conditions == nil {
			myCrdResource.Status.Conditions = append(myCrdResource.Status.Conditions, mygroupv1.MyCrdCondition{Status: "Success"})
		} else {
			if myCrdResource.Status.Conditions[len(myCrdResource.Status.Conditions)-1].Status != "Success" {
				// Change status to Success
				myCrdResource.Status.Conditions = append(myCrdResource.Status.Conditions, mygroupv1.MyCrdCondition{Status: "Success"})
			}
		}

	} else {
		// Status is pending if desired state is not reached
		if myCrdResource.Status.Conditions == nil {
			myCrdResource.Status.Conditions = append(myCrdResource.Status.Conditions, mygroupv1.MyCrdCondition{Status: "Pending"})
		} else {
			if myCrdResource.Status.Conditions[len(myCrdResource.Status.Conditions)-1].Status != "Pending" {
				// Change status to Pending
				myCrdResource.Status.Conditions = append(myCrdResource.Status.Conditions, mygroupv1.MyCrdCondition{Status: "Pending"})
			}
		}
	}

	if err := r.Status().Update(ctx, myCrdResource); err != nil {
		fmt.Println("Cannot update pod details.")
	}

	return nil
}

func (r *MyCrdReconciler) isOwner(myCrdResource *mygroupv1.MyCrd, pod *corev1.Pod) bool {
	if pod.ObjectMeta.OwnerReferences == nil {
		return false
	} else {
		for _, podOwnerRef := range pod.ObjectMeta.OwnerReferences {
			if podOwnerRef.UID == myCrdResource.ObjectMeta.UID {
				fmt.Println("Pod is owned by MyCrd.")
				return true
			}
		}
	}
	return false
}
