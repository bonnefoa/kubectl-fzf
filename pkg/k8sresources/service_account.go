package k8sresources

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const ServiceAccountHeader = "Cluster Namespace Name Secrets Age Labels\n"

// ServiceAccount is the summary of a kubernetes service account
type ServiceAccount struct {
	ResourceMeta
	numberSecrets string
}

// NewServiceAccountFromRuntime builds a pod from informer result
func NewServiceAccountFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	s := &ServiceAccount{}
	s.FromRuntime(obj, config)
	return s
}

// FromRuntime builds object from the informer's result
func (s *ServiceAccount) FromRuntime(obj interface{}, config CtorConfig) {
	serviceAccount := obj.(*corev1.ServiceAccount)
	s.FromObjectMeta(serviceAccount.ObjectMeta)
	s.numberSecrets = strconv.Itoa(len(serviceAccount.Secrets))
}

// HasChanged returns true if the resource's dump needs to be updated
func (s *ServiceAccount) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (s *ServiceAccount) ToString() string {
	line := strings.Join([]string{
		s.cluster,
		s.namespace,
		s.name,
		s.numberSecrets,
		s.resourceAge(),
		s.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
