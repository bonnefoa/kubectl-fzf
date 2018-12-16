package main

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
)

const DeploymentHeader = "Namespace Name Age Labels\n"

// Deployment is the summary of a kubernetes deployment
type Deployment struct {
	ResourceMeta
}

// NewDeploymentFromRuntime builds a k8sresource from informer result
func NewDeploymentFromRuntime(obj interface{}) K8sResource {
	d := &Deployment{}
	d.FromRuntime(obj)
	return d
}

// FromRuntime builds object from the informer's result
func (s *Deployment) FromRuntime(obj interface{}) {
	deployment := obj.(*appsv1.Deployment)
	s.FromObjectMeta(deployment.ObjectMeta)
}

// HasChanged returns true if the resource's dump needs to be updated
func (s *Deployment) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (s *Deployment) ToString() string {
	line := strings.Join([]string{s.namespace,
		s.name,
		s.resourceAge(),
		s.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
