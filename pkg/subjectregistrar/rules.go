package subjectregistrar

import (
	"context"
	"fmt"
	cattlerbacv1 "github.com/rmweir/role-keeper/api/v1"
	"github.com/sirupsen/logrus"
	k8srbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func updateRulesForRoles(ctx context.Context, sr *cattlerbacv1.SubjectRegistrar, c client.Client) {
	appliedRules := make(map[string]bool)
	for _, rule := range sr.Status.AppliedRules {
		appliedRules[fmt.Sprintf("%s/%s", rule.Namespace, rule.String())] = true
	}

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

		role := k8srbacv1.Role{}
		if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, &role); err != nil {
			logrus.Errorf("error getting role [%s:%s]", role.Namespace, role.Name)
			continue
		}

		if _, ok := appliedRules[fmt.Sprintf("%s/%s", ns, role.String())]; ok {
			continue
		}

	}
}

func createSubjectRegistrar(ns string, role k8srbacv1.Role) cattlerbacv1.SubjectRegistrar {
	return cattlerbacv1.SubjectRegistrar{
		Spec: cattlerbacv1.SubjectRegistrarSpec{},
	}
}
