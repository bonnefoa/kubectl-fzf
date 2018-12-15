package main

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ConfigMap is the summary of a kubernetes configMap
type ConfigMap struct {
	ResourceMeta
}

// FromRuntime builds object from the informer's result
func (c *ConfigMap) FromRuntime(obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	c.FromObjectMeta(configMap.ObjectMeta)
}

// HasChanged returns true if the resource's dump needs to be updated
func (c *ConfigMap) HasChanged(k K8sResource) bool {
	return true
}

// Header generates the csv header for the resource
func (c *ConfigMap) Header() string {
	return "Namespace Name Age Labels\n"
}

// ToString serializes the object to strings
func (c *ConfigMap) ToString() string {
	line := strings.Join([]string{c.namespace,
		c.name,
		c.resourceAge(),
		c.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
