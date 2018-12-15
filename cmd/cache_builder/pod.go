package main

import (
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
)

// Pod is the summary of a kubernetes pod
type Pod struct {
	ResourceMeta
	hostIP     string
	podIP      string
	nodeName   string
	containers []string
	phase      string
}

func getPhase(p *corev1.Pod) string {
	for _, v := range p.Status.ContainerStatuses {
		if v.State.Waiting != nil && v.State.Waiting.Reason != "Completed" {
			return v.State.Waiting.Reason
		}
	}
	return string(p.Status.Phase)
}

// FromRuntime builds object from the informer's result
func (p *Pod) FromRuntime(obj interface{}) {
	pod := obj.(*corev1.Pod)
	glog.V(19).Infof("Reading meta %#v", pod)
	p.FromObjectMeta(pod.ObjectMeta)
	p.hostIP = pod.Status.HostIP
	p.podIP = pod.Status.PodIP
	p.nodeName = pod.Spec.NodeName
	p.phase = getPhase(pod)

	containers := pod.Spec.Containers
	containers = append(containers, pod.Spec.InitContainers...)
	p.containers = make([]string, len(containers))
	for k, v := range containers {
		p.containers[k] = v.Name
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (p *Pod) HasChanged(k K8sResource) bool {
	oldPod := k.(*Pod)
	return (p.podIP != oldPod.podIP ||
		p.phase != oldPod.phase ||
		StringMapsEqual(p.labels, oldPod.labels) ||
		p.nodeName != oldPod.nodeName)
}

// Header generates the csv header for the resource
func (p *Pod) Header() string {
	return "Namespace Name PodIp HostIp NodeName Phase Containers Age Labels\n"
}

// ToString serializes the object to strings
func (p *Pod) ToString() string {
	lst := []string{
		p.namespace,
		p.name,
		p.podIP,
		p.hostIP,
		p.nodeName,
		p.phase,
		JoinSlicesOrNone(p.containers, ","),
		p.resourceAge(),
		p.labelsString(),
	}
	return DumpLine(lst)
}
