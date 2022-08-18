package resources

import (
	"kubectlfzf/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

// Endpoint is the summary of a kubernetes endpoints
type Endpoints struct {
	ResourceMeta
	ReadyIps     []string
	ReadyPods    []string
	NotReadyIps  []string
	NotReadyPods []string
}

// NewEndpointsFromRuntime builds a k8s resource from informer result
func NewEndpointsFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	e := &Endpoints{}
	e.FromRuntime(obj, config)
	return e
}

// FromRuntime builds object from the informer's result
func (e *Endpoints) FromRuntime(obj interface{}, config CtorConfig) {
	endpoints := obj.(*corev1.Endpoints)
	e.FromObjectMeta(endpoints.ObjectMeta, config)
	for _, subsets := range endpoints.Subsets {
		for _, v := range subsets.Addresses {
			e.ReadyIps = append(e.ReadyIps, v.IP)
			if v.TargetRef != nil && v.TargetRef.Kind == "Pod" {
				e.ReadyPods = append(e.ReadyPods, v.TargetRef.Name)
			}
		}
		for _, v := range subsets.NotReadyAddresses {
			e.NotReadyIps = append(e.NotReadyIps, v.IP)
			if v.TargetRef != nil && v.TargetRef.Kind == "Pod" {
				e.NotReadyPods = append(e.NotReadyPods, v.TargetRef.Name)
			}
		}
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (e *Endpoints) HasChanged(k K8sResource) bool {
	oldE := k.(*Endpoints)
	return !(util.StringSlicesEqual(e.ReadyIps, oldE.ReadyIps) &&
		util.StringSlicesEqual(e.ReadyPods, oldE.ReadyPods) &&
		util.StringSlicesEqual(e.NotReadyIps, oldE.NotReadyIps) &&
		util.StringSlicesEqual(e.NotReadyIps, oldE.NotReadyIps))
}

// ToString serializes the object to strings
func (e *Endpoints) ToStrings() []string {
	line := []string{
		e.Cluster,
		e.Namespace,
		e.Name,
		e.resourceAge(),
		util.JoinSlicesWithMaxOrNone(e.ReadyIps, 20, ","),
		util.JoinSlicesWithMaxOrNone(e.ReadyPods, 20, ","),
		util.JoinSlicesWithMaxOrNone(e.NotReadyIps, 20, ","),
		util.JoinSlicesWithMaxOrNone(e.NotReadyPods, 20, ","),
		e.labelsString(),
	}
	return util.DumpLines(line)
}
