package preupgrade

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// Used to authenticate to create the client config
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"

	storagev3 "helm.sh/helm/v3/pkg/storage"
	driverv3 "helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	cfg       *rest.Config
	clientset *kubernetes.Clientset
)

var InstallRetry = wait.Backoff{
	Steps:    40,
	Duration: 5 * time.Second,
	Factor:   1.5,
	Jitter:   0.1,
}

func init() {
	cfg = ctrl.GetConfigOrDie()
	clientset, _ = kubernetes.NewForConfig(cfg)

}

func notFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func Validate(name, namespace string) bool {
	log.Info("Validating the release is present or not")

	clientV1, err := typev1.NewForConfig(cfg)
	if err != nil {
		log.Errorf("error getting kubeconfig %v", err)
		return false
	}

	storageBackendV3 := storagev3.Init(driverv3.NewSecrets(clientV1.Secrets(namespace)))
	releaseHistory, err := storageBackendV3.History(name)
	if err != nil {
		if notFoundErr(err) || len(releaseHistory) == 0 {
			log.Errorf("release with name %s not found", name)
		}
		return false
	}

	return true
}

func Do(name, namespace, image string) error {
	defer cleanup(namespace)

	sa := getServiceAccount(namespace)
	if _, err := clientset.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), sa, metav1.CreateOptions{}); err != nil {
		return err
	}

	role := getRole(namespace)
	if _, err := clientset.RbacV1().Roles(namespace).Create(context.TODO(), role, metav1.CreateOptions{}); err != nil {
		return err
	}

	rolebinding := getRoleBinding(namespace)
	if _, err := clientset.RbacV1().RoleBindings(namespace).Create(context.TODO(), rolebinding, metav1.CreateOptions{}); err != nil {
		return err
	}

	job := getJobObject(name, namespace, image)
	if _, err := clientset.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{}); err != nil {
		return err
	}

	_ = wait.ExponentialBackoff(InstallRetry, func() (done bool, err error) {
		completeJob, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), job.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, nil
		}
		if completeJob.Status.Succeeded > 0 {
			return true, nil
		}
		return false, nil
	})

	log.Info("Helm release updated, now you can upgrade the TVM :)")

	return nil
}

func cleanup(namespace string) {
	log.Info("Deleting the job")
	if err := clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Errorf("Error deleting the job %s, you need to manually delete it", jobName)
		}
	}

	if err := clientset.RbacV1().RoleBindings(namespace).Delete(context.TODO(), roleBindingName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Errorf("Error deleting the rolebinding %s, you need to manually delete it", roleBindingName)
		}
	}

	if err := clientset.RbacV1().Roles(namespace).Delete(context.TODO(), roleName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Errorf("Error deleting the role %s, you need to manually delete it", roleName)
		}
	}

	if err := clientset.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), serviceAccountName,
		metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Errorf("Error deleting the service account %s, you need to manually delete it", serviceAccountName)
		}
	}
}
