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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
	var srr rbacv1.SubjectRegistrarList
	if err := r.List(ctx, &srr, client.MatchingFields{"spec.subjectID": "a"}, client.MatchingFields{"spec.subjectKind": ""}); err != nil {
		return ctrl.Result{}, err
	}
	// TODO(user): your logic here

	return ctrl.Result{}, nil
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
		return []string{srr.Spec.SubjectKind, srr.Spec.SubjectKind}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.SubjectRoleRequest{}).
		Complete(r)
}
