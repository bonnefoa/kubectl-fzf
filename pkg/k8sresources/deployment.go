package k8sresources

import (
	"strconv"

	"kubectlfzf/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentHeader is the header file for deployment
const DeploymentHeader = "Namespace Name Desired Current Up-to-date Available Age Labels\n"

// Deployment is the summary of a kubernetes deployment
type Deployment struct {
	ResourceMeta
	desiredReplicas   string
	availableReplicas string
	updatedReplicas   string
	currentReplicas   string
}

// NewDeploymentFromRuntime builds a k8sresource from informer result
func NewDeploymentFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	d := &Deployment{}
	d.FromRuntime(obj, config)
	return d
}

// FromRuntime builds object from the informer's result
func (d *Deployment) FromRuntime(obj interface{}, config CtorConfig) {
	deployment := obj.(*appsv1.Deployment)
	d.FromObjectMeta(deployment.ObjectMeta)

	status := deployment.Status
	d.desiredReplicas = "1"
	if deployment.Spec.Replicas != nil {
		d.desiredReplicas = strconv.Itoa(int(*deployment.Spec.Replicas))
	}
	d.currentReplicas = strconv.Itoa(int(status.Replicas))
	d.updatedReplicas = strconv.Itoa(int(status.UpdatedReplicas))
	d.availableReplicas = strconv.Itoa(int(status.AvailableReplicas))
}

// HasChanged returns true if the resource's dump needs to be updated
func (d *Deployment) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (d *Deployment) ToString() string {
	line := []string{d.namespace,
		d.name,
		d.desiredReplicas,
		d.currentReplicas,
		d.updatedReplicas,
		d.availableReplicas,
		d.resourceAge(),
		d.labelsString(),
	}
	return util.DumpLine(line)
}
