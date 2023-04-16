package auth

import (
	"context"
	"fmt"
	cattlerbacv1 "github.com/rmweir/role-keeper/api/v1"
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/kubernetes/pkg/registry/rbac/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SubjectRegistrarRuleResolver struct {
	validation.AuthorizationRuleResolver

	client client.Client
}

type rbacGetter struct {
	client client.Client
}

func NewSubjectRegistrarRuleResolver(client client.Client) *SubjectRegistrarRuleResolver {
	getter := &rbacGetter{client: client}
	return &SubjectRegistrarRuleResolver{
		AuthorizationRuleResolver: validation.NewDefaultRuleResolver(getter, getter, getter, getter),
		client:                    client,
	}
}

//TODO: have to make RulesFor, might be possible to use same as default

func (s *SubjectRegistrarRuleResolver) VisitRulesFor(user user.Info, namespace string, visitor func(source fmt.Stringer, rule *rbacv1.PolicyRule, err error) bool) {
	userID := user.GetName()
	sr := cattlerbacv1.SubjectRegistrar{}
	var err error
	if err = s.client.Get(context.Background(), client.ObjectKey{Name: userID}, &sr); err != nil {
		logrus.Errorf("failed to get SubjectRegistrar [%s] while visiting rules: %v", userID, err)
	}
	if len(sr.Status.AppliedRoles) == 0 {
		return
	}
	appliedRolesForNS := sr.Status.AppliedRoles[namespace]
	if len(appliedRolesForNS) == 0 {
		return
	}
	for roleRefKey, val := range sr.Status.AppliedRoles[namespace] {
		if val < 1 {
			continue
		}
		var roleRef rbacv1.RoleRef
		if err = json.Unmarshal([]byte(roleRefKey), &roleRef); err != nil {
			logrus.Errorf("failed to unmarshal roleRef [%s]: %v", roleRefKey, err)
		}
		s.visitAll(namespace, roleRef, visitor)
	}
}

func (s *SubjectRegistrarRuleResolver) visitAll(namespace string, roleRef rbacv1.RoleRef, visitor func(source fmt.Stringer, rule *rbacv1.PolicyRule, err error) bool) {
	rules, err := s.GetRoleReferenceRules(roleRef, namespace)
	if err != nil {
		visitor(nil, nil, err)
	}
	for _, rule := range rules {
		visitor(nil, &rule, nil)
	}
}

func (r *rbacGetter) GetRole(namespace, name string) (*rbacv1.Role, error) {
	var role rbacv1.Role
	err := r.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *rbacGetter) ListRoleBindings(namespace string) ([]*rbacv1.RoleBinding, error) {
	var rbList rbacv1.RoleBindingList
	err := r.client.List(context.TODO(), &rbList, client.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	rbs := make([]*rbacv1.RoleBinding, len(rbList.Items))
	for index, val := range rbList.Items {
		rbs[index] = &val
	}
	return rbs, nil
}

func (r *rbacGetter) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	var cr rbacv1.ClusterRole
	err := r.client.Get(context.TODO(), client.ObjectKey{Name: name}, &cr)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (r *rbacGetter) ListClusterRoleBindings() ([]*rbacv1.ClusterRoleBinding, error) {
	var crbList rbacv1.ClusterRoleBindingList
	err := r.client.List(context.TODO(), &crbList)
	if err != nil {
		return nil, err
	}
	crbs := make([]*rbacv1.ClusterRoleBinding, len(crbList.Items))
	for index, val := range crbList.Items {
		crbs[index] = &val
	}
	return crbs, nil
}
