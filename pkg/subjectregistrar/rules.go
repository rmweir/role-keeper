package subjectregistrar

//TODO will probably want to delete this file and keep structure simple for now
import (
	"context"
	"fmt"
	"reflect"
	"strings"

	cattlerbacv1 "github.com/rmweir/role-keeper/api/v1"
	"github.com/sirupsen/logrus"
	k8srbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateRulesForRoles(ctx context.Context, sr *cattlerbacv1.SubjectRegistrar, c client.Client) bool {
	appliedRules := make(map[string]bool)
	for _, rule := range sr.Status.AppliedRules {
		appliedRules[fmt.Sprintf("%s/%s", rule.Namespace, rule.String())] = true
	}
	updatedRolesRules := make(map[string]bool)
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

		addMissingRules(ns, updatedRolesRules, role.Rules)
	}
	return reflect.DeepEqual(appliedRules, updatedRolesRules)
}

func addMissingRules(ns string, applied map[string]bool, rules []k8srbacv1.PolicyRule) {
	for _, rule := range rules {
		applied[fmt.Sprintf("%s/%s", ns, rule.String())] = true
	}
}
