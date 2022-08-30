package resources

import (
	"strconv"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"

	appsv1 "k8s.io/api/apps/v1"
)

// Deployment is the summary of a kubernetes deployment
type Deployment struct {
	ResourceMeta
	DesiredReplicas   string
	AvailableReplicas string
	UpdatedReplicas   string
	CurrentReplicas   string
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
	d.FromObjectMeta(deployment.ObjectMeta, config)

	status := deployment.Status
	d.DesiredReplicas = "1"
	if deployment.Spec.Replicas != nil {
		d.DesiredReplicas = strconv.Itoa(int(*deployment.Spec.Replicas))
	}
	d.CurrentReplicas = strconv.Itoa(int(status.Replicas))
	d.UpdatedReplicas = strconv.Itoa(int(status.UpdatedReplicas))
	d.AvailableReplicas = strconv.Itoa(int(status.AvailableReplicas))
}

// HasChanged returns true if the resource's dump needs to be updated
func (d *Deployment) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (d *Deployment) ToStrings() []string {
	line := []string{
		d.Namespace,
		d.Name,
		d.DesiredReplicas,
		d.CurrentReplicas,
		d.UpdatedReplicas,
		d.AvailableReplicas,
		d.resourceAge(),
		d.labelsString(),
	}
	return util.DumpLines(line)
}
