package k8sresources

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"kubectlfzf/pkg/util"
)

const EndpointsHeader = "Cluster Namespace Name Age ReadyIps ReadyPods NotReadyIps NotReadyPods Labels\n"

// Endpoints is the summary of a kubernetes endpoints
type Endpoints struct {
	ResourceMeta
	readyIps     []string
	readyPods    []string
	notReadyIps  []string
	notReadyPods []string
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
			e.readyIps = append(e.readyIps, v.IP)
			if v.TargetRef != nil && v.TargetRef.Kind == "Pod" {
				e.readyPods = append(e.readyPods, v.TargetRef.Name)
			}
		}
		for _, v := range subsets.NotReadyAddresses {
			e.notReadyIps = append(e.notReadyIps, v.IP)
			if v.TargetRef != nil && v.TargetRef.Kind == "Pod" {
				e.notReadyPods = append(e.notReadyPods, v.TargetRef.Name)
			}
		}
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (e *Endpoints) HasChanged(k K8sResource) bool {
	oldE := k.(*Endpoints)
	return (util.StringSlicesEqual(e.readyIps, oldE.readyIps) ||
		util.StringSlicesEqual(e.readyPods, oldE.readyPods) ||
		util.StringSlicesEqual(e.notReadyIps, oldE.notReadyIps) ||
		util.StringSlicesEqual(e.notReadyIps, oldE.notReadyIps))
}

// ToString serializes the object to strings
func (e *Endpoints) ToString() string {
	line := strings.Join([]string{
		e.cluster,
		e.namespace,
		e.name,
		e.resourceAge(),
		util.JoinSlicesWithMaxOrNone(e.readyIps, 20, ","),
		util.JoinSlicesWithMaxOrNone(e.readyPods, 20, ","),
		util.JoinSlicesWithMaxOrNone(e.notReadyIps, 20, ","),
		util.JoinSlicesWithMaxOrNone(e.notReadyPods, 20, ","),
		e.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
