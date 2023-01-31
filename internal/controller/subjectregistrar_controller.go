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
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	rbacv1 "github.com/rmweir/role-keeper/api/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// SubjectRegistrarReconciler reconciles a SubjectRegistrar object
type SubjectRegistrarReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=rbac.cattle.io,resources=subjectregistrars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.cattle.io,resources=subjectregistrars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rbac.cattle.io,resources=subjectregistrars/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SubjectRegistrar object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SubjectRegistrarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	ns := req.Namespace
	id := req.Name
	var sr rbacv1.SubjectRegistrar
	if err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: id}, &sr); err != nil {
		return ctrl.Result{}, err
	}

	updated, err := r.UpdateRulesForRoles(ctx, sr)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	if updated {
		return ctrl.Result{}, nil
	}

	updated, err = r.processAddQueue(ctx, sr)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	if updated {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *SubjectRegistrarReconciler) processAddQueue(ctx context.Context, sr rbacv1.SubjectRegistrar) (bool, error) {
	var update bool
	if len(sr.Status.AddQueue) == 0 {
		return update, nil
	}
	for _, srrID := range sr.Status.AddQueue {
		parts := strings.Split(srrID, ":")
		if len(parts) != 2 {
			logrus.Errorf("improper format for SubjectRoleRequest ID [%s] in SubjectRegistrar [%s:%s] AddRoles field."+
				" Should be of format \"<namespace>:<name>\"", srrID, sr.Namespace, sr.Name)
		}
		val := sr.Status.AppliedRoles[srrID] + 1
		sr.Status.AppliedRoles[srrID] = val
		update = true
	}
	(&sr).Status.AddQueue = []string{}
	if err := r.Client.Status().Update(ctx, &sr); err != nil {
		return false, fmt.Errorf("")
	}
	return true, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubjectRegistrarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &rbacv1.SubjectRegistrar{}, "status.rolesApplied", func(obj client.Object) []string {
		srr := obj.(*rbacv1.SubjectRegistrar)
		var roleNames []string
		for key := range srr.Status.AppliedRoles {
			roleNames = append(roleNames, key)
		}
		return roleNames
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.SubjectRegistrar{}).
		Owns(&v1.Role{}).
		Watches(&source.Kind{Type: &v1.Role{}},
			handler.EnqueueRequestsFromMapFunc(
				r.roleToSRR)).
		Complete(r)
}

func (r *SubjectRegistrarReconciler) roleToSRR(object client.Object) []reconcile.Request {
	role := object.(*v1.Role)
	fullRoleId := fmt.Sprintf("%s:%s", role.Namespace, role.Name)
	var srs rbacv1.SubjectRegistrarList
	if err := r.Client.List(context.Background(), &srs, client.MatchingFields{"status.rolesApplied": fullRoleId}); err != nil {
		logrus.Errorf("error listing SubjectRegistrars that match status.roleApplied field for value [%s]", fullRoleId)
		return nil
	}

	var result []reconcile.Request
	for _, sr := range srs.Items {
		result = append(result, reconcile.Request{
			NamespacedName: types2.NamespacedName{
				Namespace: sr.Namespace,
				Name:      sr.Name,
			},
		})
	}
	return result
}

func (r *SubjectRegistrarReconciler) UpdateRulesForRoles(ctx context.Context, sr rbacv1.SubjectRegistrar) (bool, error) {
	appliedRules := make(map[string]rbacv1.AppliedRule)
	for _, rule := range sr.Status.AppliedRules {
		appliedRules[fmt.Sprintf("%s/%s", rule.Namespace, rule.String())] = rule
	}

	updatedRolesRules := make(map[string]rbacv1.AppliedRule)
	for roleID := range sr.Status.AppliedRoles {
		parts := strings.Split(roleID, ":")
		if len(parts) != 0 && len(parts) != 2 {
			logrus.Errorf("cannot parse role [%s] for subjectRegistrar [%s:%s]. Role name should be of format"+
				" \"<namespace>:<id>\" or \"<id>\"", sr.Namespace, sr.Name, parts[0])
			continue
		}
		var ns, name string
		if len(parts) == 1 {
			name = parts[0]
		} else {
			ns, name = parts[0], parts[1]
		}

		role := v1.Role{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, &role); err != nil {
			logrus.Errorf("error getting role [%s:%s]", role.Namespace, role.Name)
			continue
		}

		addMissingRules(ns, updatedRolesRules, role.Rules)
	}
	if reflect.DeepEqual(appliedRules, updatedRolesRules) {
		return false, nil
	}
	var result []rbacv1.AppliedRule
	for _, appliedRule := range updatedRolesRules {
		result = append(result, appliedRule)
	}
	sr.Status.AppliedRules = result
	if err := r.Client.Status().Update(ctx, &sr); err != nil {
		return false, err
	}
	return true, nil
}

func addMissingRules(ns string, applied map[string]rbacv1.AppliedRule, rules []v1.PolicyRule) {
	for _, rule := range rules {
		applied[fmt.Sprintf("%s/%s", ns, rule.String())] = rbacv1.AppliedRule{Namespace: ns, PolicyRule: rule}
	}
}
