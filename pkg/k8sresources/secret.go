package k8sresources

import (
	"strconv"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
)

const SecretHeader = "Cluster Namespace Name Type Data Age Labels\n"

// Secret is the summary of a kubernetes secret
type Secret struct {
	ResourceMeta
	secretType string
	data       string
}

// NewSecretFromRuntime builds a secret from informer result
func NewSecretFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	s := &Secret{}
	s.FromRuntime(obj, config)
	return s
}

// FromRuntime builds object from the informer's result
func (s *Secret) FromRuntime(obj interface{}, config CtorConfig) {
	secret := obj.(*corev1.Secret)
	glog.V(19).Infof("Reading meta %#v", secret)
	s.FromObjectMeta(secret.ObjectMeta, config)
	s.secretType = string(secret.Type)
	s.data = strconv.Itoa(len(secret.Data))
}

// HasChanged returns true if the resource's dump needs to be updated
func (s *Secret) HasChanged(k K8sResource) bool {
	return true
}
