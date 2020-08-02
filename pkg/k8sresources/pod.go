package k8sresources

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"kubectlfzf/pkg/util"
)

// PodHeader is the header for pod files
const PodHeader = "Cluster Namespace Name PodIp HostIp NodeName Phase Containers Tolerations Claims Age Labels FieldSelectors\n"

// Pod is the summary of a kubernetes pod
type Pod struct {
	ResourceMeta
	hostIP         string
	podIP          string
	nodeName       string
	tolerations    []string
	containers     []string
	claims         []string
	phase          string
	fieldSelectors string
}

func getPhase(p *corev1.Pod) string {
	for _, v := range p.Status.InitContainerStatuses {
		if v.State.Waiting != nil && v.State.Waiting.Reason != "" {
			return fmt.Sprintf("Init:%s", v.State.Waiting.Reason)
		}
		if v.State.Terminated != nil && v.State.Terminated.Reason != "Completed" {
			return fmt.Sprintf("Init:%s", v.State.Terminated.Reason)
		}
	}
	for _, v := range p.Status.ContainerStatuses {
		if v.State.Waiting != nil && v.State.Waiting.Reason != "" {
			return v.State.Waiting.Reason
		}
		if v.State.Terminated != nil && v.State.Terminated.Reason != "Completed" {
			return v.State.Terminated.Reason
		}
	}
	for _, v := range p.Status.Conditions {
		if v.Status != "True" && v.Reason != "" {
			return v.Reason
		}
	}
	if p.Status.Reason != "" {
		return p.Status.Reason
	}
	return string(p.Status.Phase)
}

// NewPodFromRuntime builds a pod from informer result
func NewPodFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	p := &Pod{}
	p.FromRuntime(obj, config)
	return p
}

// FromRuntime builds object from the informer's result
func (p *Pod) FromRuntime(obj interface{}, config CtorConfig) {
	pod := obj.(*corev1.Pod)
	glog.V(19).Infof("Reading meta %#v", pod)
	p.FromObjectMeta(pod.ObjectMeta, config)
	p.hostIP = pod.Status.HostIP
	p.podIP = pod.Status.PodIP
	spec := pod.Spec
	p.nodeName = spec.NodeName
	p.phase = getPhase(pod)

	fieldSelectors := make([]string, 0)
	if p.nodeName != "" {
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("spec.nodeName=%s", p.nodeName))
	}
	fieldSelectors = append(fieldSelectors, fmt.Sprintf("status.phase=%s", pod.Status.Phase))
	p.fieldSelectors = util.JoinSlicesOrNone(fieldSelectors, ",")

	containers := spec.Containers
	containers = append(containers, spec.InitContainers...)
	p.containers = make([]string, len(containers))
	for k, v := range containers {
		p.containers[k] = v.Name
	}

	volumes := spec.Volumes
	for _, v := range volumes {
		if v.PersistentVolumeClaim != nil {
			fullClaimName := fmt.Sprintf("%s/%s", p.ResourceMeta.namespace,
				v.PersistentVolumeClaim.ClaimName)
			p.claims = append(p.claims, fullClaimName)
		}
	}
	tolerations := spec.Tolerations
	p.tolerations = make([]string, 0)
	for _, v := range tolerations {
		if strings.HasPrefix(v.Key, "node.kubernetes.io") {
			continue
		}
		var toleration string
		if v.Operator == "Equal" {
			toleration = fmt.Sprintf("%s=%s:%s", v.Key, v.Value, v.Effect)
		} else if v.Key == "" {
			toleration = "Exists"
		} else {
			toleration = fmt.Sprintf("%s:%s", v.Key, v.Effect)
		}
		p.tolerations = append(p.tolerations, toleration)
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (p *Pod) HasChanged(k K8sResource) bool {
	oldPod := k.(*Pod)
	return (p.podIP != oldPod.podIP ||
		p.phase != oldPod.phase ||
		util.StringMapsEqual(p.labels, oldPod.labels) ||
		p.nodeName != oldPod.nodeName)
}

// ToString serializes the object to strings
func (p *Pod) ToString() string {
	lst := []string{
		p.cluster,
		p.namespace,
		p.name,
		p.podIP,
		p.hostIP,
		p.nodeName,
		p.phase,
		util.TruncateString(util.JoinSlicesOrNone(p.containers, ","), 300),
		util.JoinSlicesOrNone(p.tolerations, ","),
		util.JoinSlicesOrNone(p.claims, ","),
		p.resourceAge(),
		p.labelsString(),
		p.fieldSelectors,
	}
	return util.DumpLine(lst)
}
