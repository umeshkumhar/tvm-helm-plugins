package preupgrade

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Used to authenticate to create the client config
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	storagev3 "helm.sh/helm/v3/pkg/storage"
	driverv3 "helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	err       error
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
	kubeconfig := fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, _ = kubernetes.NewForConfig(cfg)
}

func notFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func Validate(name, namespace, serviceaccount string) bool {
	log.Info("Validating the sufficient permissions to run the pre upgrade job")

	clientV1, err := typev1.NewForConfig(cfg)
	if err != nil {
		log.Errorf("error getting kubeconfig %v", err)
		return false
	}

	validUser := validateUserPermissions(namespace)
	validSA := validateServiceAccountPermissions(namespace, serviceaccount)

	if !validUser || !validSA {
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

func Do(name, namespace, serviceaccount string) error {
	defer cleanup(namespace)

	job := getJobObject(name, namespace, serviceaccount)
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
	if err := clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), JobName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Errorf("Error deleting the job %s, you need to manually delete it", JobName)
		}
	}
}

func validateServiceAccountPermissions(namespace, serviceAccount string) bool {
	var valid = true
	user := "system:serviceaccount:" + namespace + ":" + serviceAccount
	verbs := []string{"list", "get", "update"}

	cfg.Impersonate = rest.ImpersonationConfig{
		UserName: user,
	}

	clset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	for i := range verbs {
		sar := getSelfSubjectAccessReview(namespace, verbs[i], "secret")
		resp, err := clset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), sar, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("error running authentication on service account %v %v", serviceAccount, err)
			return false
		}

		if resp.Status.Allowed {
			log.Infof("You have %s secret permission", verbs[i])
		} else {
			valid = false
			log.Errorf("Service Account %s doesn't have permission to %s secret", serviceAccount, verbs[i])
		}
	}

	return valid
}

func validateUserPermissions(namespace string) bool {
	var valid = true
	verbs := []string{"create", "delete"}
	for i := range verbs {
		sar := getSelfSubjectAccessReview(namespace, verbs[i], "job")

		resp, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), sar, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("error running authentication on User %v", err)
			return false
		}

		if resp.Status.Allowed {
			log.Infof("You have %s job permission", verbs[i])
		} else {
			valid = false
			log.Errorf("User doesn't have permission to %s job", verbs[i])
		}
	}

	return valid
}
