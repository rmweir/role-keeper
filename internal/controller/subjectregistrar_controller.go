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
	"strings"
	"time"

	errors2 "github.com/pkg/errors"
	rbacv1 "github.com/rmweir/role-keeper/api/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	RolesAppliedIndexField = "status.rolesApplied"
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
	name := req.Name
	var sr rbacv1.SubjectRegistrar
	if err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, &sr); err != nil {
		return ctrl.Result{}, err
	}

	updated, err := r.UpdateRulesForRoles(ctx, sr)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	if updated {
		return ctrl.Result{}, nil
	}

	waitingOnSubjectRoleRequestStatus, updated, err := r.processAddQueue(ctx, sr)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	if updated {
		return ctrl.Result{}, nil
	}
	if waitingOnSubjectRoleRequestStatus {
		requeuTime := 5 * time.Second
		logrus.Debugf("waiting on SubjectRoleRequest(s) status. Requeueing SubjectRegistrar [%s:%s] after [%s]"+
			" has elapsed", sr.Namespace, sr.Name, requeuTime)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SubjectRegistrarReconciler) processAddQueue(ctx context.Context, sr rbacv1.SubjectRegistrar) (bool, bool, error) {
	srID := fmt.Sprintf("%s:%s", sr.Namespace, sr.Name)
	if len(sr.Status.AddQueue) == 0 {
		logrus.Debugf("skipping processing addQueue for SubjectRegistrar [%s] because it is empty", srID)
		return false, false, nil
	}

	var waitingOnSubjectRoleRequestStatusToChange []string
	for _, srrID := range sr.Status.AddQueue {
		srr, validationErr, err := r.getSubjectRoleRequestFromID(ctx, srrID)
		if err != nil {
			return false, false, err
		}

		if validationErr != nil {
			logrus.Errorf("invalid SubjectRoleRequest ID [%s]. Removing from SubjectRegistrar [%s] addQueue", srrID, srID)
			continue
		}

		if srr.Status.Status != rbacv1.InQueue {
			logrus.Debugf("waiting for SubjectRoleRequest startus to change to [%s] before its role can be applied"+
				" to SubjectRegistrar [%s]", srrID, srID)
			waitingOnSubjectRoleRequestStatusToChange = append(waitingOnSubjectRoleRequestStatusToChange, srrID)
			continue
		}

		validationErr, err = r.validateRoleFromID(ctx, fmt.Sprintf(srr.Spec.Role))
		if err != nil {
			return false, false, err
		}
		if validationErr != nil {
			logrus.Errorf("invalid v1.Role ID [%s] on SubjectRoleRequest [%s]. Removing from SubjectRegistrar [%s] addQueue", srr.Spec.Role, srrID, srID)
		}

		if err = r.writeErrToSubjectRoleRequest(ctx, srr, validationErr); err != nil {
			return false, false, err
		}

		appliedRoleKey := fmt.Sprintf("%s:%s", srr.Spec.TargetNamespace, srr.Spec.Role)
		if sr.Status.AppliedRoles == nil {
			sr.Status.AppliedRoles = make(map[string]int)
		}
		val := sr.Status.AppliedRoles[appliedRoleKey] + 1
		sr.Status.AppliedRoles[appliedRoleKey] = val
	}
	if len(sr.Status.AddQueue) == len(waitingOnSubjectRoleRequestStatusToChange) {
		return len(waitingOnSubjectRoleRequestStatusToChange) > 0, false, nil
	}
	sr.Status.AddQueue = waitingOnSubjectRoleRequestStatusToChange
	if err := r.Client.Status().Update(ctx, &sr); err != nil {
		return len(waitingOnSubjectRoleRequestStatusToChange) > 0, false, fmt.Errorf("")
	}
	return len(waitingOnSubjectRoleRequestStatusToChange) > 0, true, nil
}

func (r *SubjectRegistrarReconciler) writeErrToSubjectRoleRequest(ctx context.Context, srr rbacv1.SubjectRoleRequest, err error) error {
	if err == nil {
		return nil
	}
	logrus.Debugf("writing error [%v] to SubjectRoleRequest [%s:%s]", err, srr.Namespace, srr.Name)
	srr.Status.Status = rbacv1.Failure
	srr.Status.FailureMessage = err.Error()
	return errors2.WithStack(r.Client.Update(ctx, &srr))
}

func (r *SubjectRegistrarReconciler) validateRoleFromID(ctx context.Context, id string) (error, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid role ID [%s]. ID should be of the format <roleNamespace:roleName>", id), nil
	}

	var role v1.Role
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: parts[0], Name: parts[1]}, &role)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("invalid role [%s]: %w", id, err), nil
		}
		return nil, err
	}
	return nil, nil
}

func (r *SubjectRegistrarReconciler) getSubjectRoleRequestFromID(ctx context.Context, srrID string) (rbacv1.SubjectRoleRequest, error, error) {
	parts := strings.Split(srrID, ":")
	if len(parts) != 2 {
		return rbacv1.SubjectRoleRequest{}, fmt.Errorf("improper format for SubjectRoleRequest ID [%s] AddRoles field."+
			" Should be of format \"<namespace>:<name>\"", srrID), nil
	}
	var srr rbacv1.SubjectRoleRequest
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: parts[0], Name: parts[1]}, &srr); err != nil {
		if errors.IsNotFound(err) {
			return rbacv1.SubjectRoleRequest{}, fmt.Errorf("SubjectRoleRequest [%s] addQueue, does not exist."+
				" Removing from addQueue", srrID), nil
		}
		return rbacv1.SubjectRoleRequest{}, nil, errors2.WithStack(err)
	}
	return srr, nil, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubjectRegistrarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &rbacv1.SubjectRegistrar{}, RolesAppliedIndexField, func(obj client.Object) []string {
		sr := obj.(*rbacv1.SubjectRegistrar)
		var roleNames []string
		for key := range sr.Status.AppliedRoles {
			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				logrus.Debugf("indexer [%s] could not parse AppliedRole ID [%s] from SubjectRegistrar [%s:%s]. AppliedRoleID should be of the format "+
					"<targetNamespace>:<roleNamespace>:<roleName>", RolesAppliedIndexField, key, sr.Namespace, sr.Name)
				continue
			}
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

// TODO: consider removing rules from SubjectRegistrar
func (r *SubjectRegistrarReconciler) UpdateRulesForRoles(ctx context.Context, sr rbacv1.SubjectRegistrar) (bool, error) {
	appliedRules := make(map[string]rbacv1.AppliedRule)
	for _, rule := range sr.Status.AppliedRules {
		appliedRules[fmt.Sprintf("%s/%s", rule.Namespace, rule.String())] = rule
	}

	updatedRolesRules := make(map[string]rbacv1.AppliedRule)
	for roleID := range sr.Status.AppliedRoles {
		parts := strings.Split(roleID, ":")
		if len(parts) != 3 {
			logrus.Errorf("cannot parse role [%s] for subjectRegistrar [%s:%s]. Role name should be of format"+
				" \"<targetNamespace>:<roleNamespace>:<roleName>\"", roleID, sr.Namespace, sr.Name)
			continue
		}
		targetNS, ns, name := parts[0], parts[1], parts[2]

		role := v1.Role{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, &role); err != nil {
			logrus.Errorf("error getting role [%s:%s]", ns, name)
			continue
		}

		addMissingRules(targetNS, updatedRolesRules, role.Rules)
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
