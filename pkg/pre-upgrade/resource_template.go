package preupgrade

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	jobName            = "tvm-upgrade-hook"
	serviceAccountName = "tvm-upgrade-service-account"
	roleName           = "tvm-upgrade-role"
	roleBindingName    = "tvm-upgrade-rolebinding"
)

func getServiceAccount(namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}
}

func getRole(namespace string) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Verbs:     []string{"get", "list", "update"},
			Resources: []string{"secrets"},
		}},
	}
}

func getRoleBinding(namespace string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      serviceAccountName,
			Namespace: namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}
}

func getJobObject(name, namespace, imageRegistry string) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "Helm",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "Helm",
					},
					Annotations: nil,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{{
						Name:  "pre-release-change",
						Image: imageRegistry + "/operator-webhook-init:v2.1.0",
						Command: []string{
							"./operator-init",
							"--upgrade",
						},
						Env: []corev1.EnvVar{{
							Name:  "INSTALL_NAMESPACE",
							Value: namespace,
						},
							{
								Name:  "RELEASE_NAME",
								Value: name,
							},
						},
						ImagePullPolicy: "Always",
					}},
					RestartPolicy: "Never",
				},
			},
		},
	}

}
