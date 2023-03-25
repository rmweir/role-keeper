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
