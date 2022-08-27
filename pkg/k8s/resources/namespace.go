package resources

import (
	"kubectlfzf/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

// Namespace is the summary of a kubernetes configMap
type Namespace struct {
	ResourceMeta
}

// NewNamespaceFromRuntime builds a pod from informer result
func NewNamespaceFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	c := &Namespace{}
	c.FromRuntime(obj, config)
	return c
}

// FromRuntime builds object from the informer's result
func (c *Namespace) FromRuntime(obj interface{}, config CtorConfig) {
	configMap := obj.(*corev1.Namespace)
	c.FromObjectMeta(configMap.ObjectMeta, config)
}

// HasChanged returns true if the resource's dump needs to be updated
func (c *Namespace) HasChanged(k K8sResource) bool {
	return false
}

// ToString serializes the object to strings
func (c *Namespace) ToStrings() []string {
	line := []string{
		c.Name,
		c.resourceAge(),
		c.labelsString(),
	}
	return util.DumpLines(line)
}
