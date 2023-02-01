/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	rbacv1 "github.com/rmweir/role-keeper/api/v1"
)

// SubjectRoleRequestReconciler reconciles a SubjectRoleRequest object
type SubjectRoleRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=rbac.cattle.io,resources=subjectrolerequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.cattle.io,resources=subjectrolerequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rbac.cattle.io,resources=subjectrolerequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SubjectRoleRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SubjectRoleRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var srr rbacv1.SubjectRoleRequest
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &srr); err != nil {
		return ctrl.Result{Requeue: true}, nil
	}

	sr, err := r.getCorrespondingSubjectRegistrar(ctx, srr)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	requeued, err := r.addToQueue(ctx, srr, sr)
	if err != nil {
		return ctrl.Result{Requeue: true}, nil
	}

	if requeued {
		// has already been requeued as a consequence of updating the SRR or corresponding SR
		return ctrl.Result{}, nil
	}

	waitingInQueue := shouldWaitForQueueExit(srr, sr)
	if waitingInQueue {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SubjectRoleRequestReconciler) setSuccess(ctx context.Context, srr rbacv1.SubjectRoleRequest) error {
	if srr.Status.Status == rbacv1.Success {
		return nil
	}
	srr.Status.Status = rbacv1.Success
	return r.Client.Status().Update(ctx, &srr)
}

func (r *SubjectRoleRequestReconciler) getCorrespondingSubjectRegistrar(ctx context.Context, srr rbacv1.SubjectRoleRequest) (rbacv1.SubjectRegistrar, error) {
	var srs rbacv1.SubjectRegistrarList
	if err := r.List(ctx, &srs, client.MatchingFields{"spec.subjectID": srr.Spec.SubjectID}, client.MatchingFields{"spec.subjectKind": srr.Spec.SubjectKind}); err != nil {
		return rbacv1.SubjectRegistrar{}, err
	}

	if len(srs.Items) == 0 {
		// TODO: revisit what should happen if no corresponding SR is found- should probably error and move this out of this function
		return rbacv1.SubjectRegistrar{}, fmt.Errorf("no SubjectRegistrarFound corresponding to SubjectRoleRequest's SubjectID [%s] and SubjectKind [%s]", srr.Spec.SubjectID, srr.Spec.SubjectKind)
	}

	if len(srs.Items) > 1 {
		// TODO: should probably tie name/id to these instead of fields so that it isn't possible to have duplicates
		return rbacv1.SubjectRegistrar{}, fmt.Errorf("found more than 1 SubjectRegistrar for SubjectID [%s] and SubjectKind [%s]", srr.Spec.SubjectID, srr.Spec.SubjectKind)
	}

	return srs.Items[0], nil
}

func (r *SubjectRoleRequestReconciler) addToQueue(ctx context.Context, srr rbacv1.SubjectRoleRequest, sr rbacv1.SubjectRegistrar) (bool, error) {
	if srr.Status.Status != "" {
		return false, nil
	}

	updateSR, err := addSubjectRoleRequestToQueue(srr, sr)
	if err != nil {
		return false, err
	}

	if updateSR {
		if err := r.Client.Status().Update(ctx, &sr); err != nil {
			return false, err
		}
		return true, nil
	}

	srr.Status.Status = rbacv1.InQueue
	if err = r.Client.Status().Update(ctx, &srr); err != nil {
		return false, fmt.Errorf("failed to update SubjectRoleRequest [%s:%s] status to \"InQueue\"", srr.Namespace, srr.Name)
	}
	return true, nil
}

func shouldWaitForQueueExit(srr rbacv1.SubjectRoleRequest, sr rbacv1.SubjectRegistrar) bool {
	if srr.Status.Status != rbacv1.InQueue {
		return false
	}
	for _, srrInQueue := range sr.Status.AddQueue {
		if getSubjectRoleRequestQueueKey(sr) == srrInQueue {
			return true
		}
	}
	return false
}

func addSubjectRoleRequestToQueue(srr rbacv1.SubjectRoleRequest, sr rbacv1.SubjectRegistrar) (bool, error) {
	if srr.Spec.Operation != rbacv1.AddRole {
		// TODO: maybe turn these into a sort of noop error
		return false, nil
	}
	if srr.Status.Status == rbacv1.Success {
		// TODO: maybe turn these into a sort of noop error
		return false, nil
	}
	if srr.Status.Status == rbacv1.InQueue {
		// TODO: maybe turn these into a sort of noop error
		return false, nil
	}

	queueKey := getSubjectRoleRequestQueueKey(sr)
	var isQueued bool
	// check if already in queue
	for _, srrInQueue := range sr.Status.AddQueue {
		if srrInQueue == queueKey {
			isQueued = true
			break
		}
	}

	if !isQueued {
		(&sr).Status.AddQueue = append(sr.Status.AddQueue, queueKey)
	}

	return true, nil
}

func getSubjectRoleRequestQueueKey(sr rbacv1.SubjectRegistrar) string {
	return fmt.Sprintf("%s:%s", sr.Namespace, sr.Name)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubjectRoleRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &rbacv1.SubjectRoleRequest{}, "spec.subjectID", func(obj client.Object) []string {
		srr := obj.(*rbacv1.SubjectRoleRequest)
		return []string{srr.Spec.SubjectID}
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &rbacv1.SubjectRoleRequest{}, "spec.subjectKind", func(obj client.Object) []string {
		srr := obj.(*rbacv1.SubjectRoleRequest)
		return []string{srr.Spec.SubjectKind}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.SubjectRoleRequest{}).
		Complete(r)
}
