package preupgrade

import (
	authv1 "k8s.io/api/authorization/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	JobName = "tvm-upgrade-hook"
)

func getJobObject(name, namespace, serviceaccount string) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      JobName,
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
					ServiceAccountName: serviceaccount,
					Containers: []corev1.Container{{
						Name:  "pre-release-change",
						Image: "gcr.io/amazing-chalice-243510/operator-webhook-init:updatecrd",
						Command: []string{
							"./operator-init",
							"--upgrade",
							"true",
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

func getSelfSubjectAccessReview(namespace, verb, resource string) *authv1.SelfSubjectAccessReview {
	return &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     "",
				Resource:  resource,
			},
		},
	}
}
