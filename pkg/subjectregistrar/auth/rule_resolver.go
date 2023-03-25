package auth

import (
	"k8s.io/kubernetes/pkg/registry/rbac/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SubjectRegistrarRuleResolver struct {
	validation.AuthorizationRuleResolver

	client client.Client
}
