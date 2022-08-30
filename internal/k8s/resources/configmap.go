package resources

import (
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	corev1 "k8s.io/api/core/v1"
)

// ConfigMap is the summary of a kubernetes configMap
type ConfigMap struct {
	ResourceMeta
}

// NewConfigMapFromRuntime builds a pod from informer result
func NewConfigMapFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	c := &ConfigMap{}
	c.FromRuntime(obj, config)
	return c
}

// FromRuntime builds object from the informer's result
func (c *ConfigMap) FromRuntime(obj interface{}, config CtorConfig) {
	configMap := obj.(*corev1.ConfigMap)
	c.FromObjectMeta(configMap.ObjectMeta, config)
}

// HasChanged returns true if the resource's dump needs to be updated
func (c *ConfigMap) HasChanged(k K8sResource) bool {
	return true
}

// ToStrings serializes the object to strings
func (c *ConfigMap) ToStrings() []string {
	line := []string{
		c.Namespace,
		c.Name,
		c.resourceAge(),
		c.labelsString(),
	}
	return util.DumpLines(line)
}
