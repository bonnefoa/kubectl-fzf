package main

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Endpoints is the summary of a kubernetes endpoints
type Endpoints struct {
	ResourceMeta
	readyIps     []string
	readyPods    []string
	notReadyIps  []string
	notReadyPods []string
}

// FromRuntime builds object from the informer's result
func (e *Endpoints) FromRuntime(obj interface{}) {
	endpoints := obj.(*corev1.Endpoints)
	e.FromObjectMeta(endpoints.ObjectMeta)
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
	return (StringSlicesEqual(e.readyIps, oldE.readyIps) ||
		StringSlicesEqual(e.readyPods, oldE.readyPods) ||
		StringSlicesEqual(e.notReadyIps, oldE.notReadyIps) ||
		StringSlicesEqual(e.notReadyIps, oldE.notReadyIps))
}

// Header generates the csv header for the resource
func (e *Endpoints) Header() string {
	return "Namespace Name Age ReadyIps ReadyPods NotReadyIps NotReadyPods Labels\n"
}

// ToString serializes the object to strings
func (e *Endpoints) ToString() string {
	line := strings.Join([]string{e.namespace,
		e.name,
		e.resourceAge(),
		JoinSlicesWithMaxOrNone(e.readyIps, 20, ","),
		JoinSlicesWithMaxOrNone(e.readyPods, 20, ","),
		JoinSlicesWithMaxOrNone(e.notReadyIps, 20, ","),
		JoinSlicesWithMaxOrNone(e.notReadyPods, 20, ","),
		e.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
