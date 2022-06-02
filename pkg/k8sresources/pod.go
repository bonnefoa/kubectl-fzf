package k8sresources

import (
	"fmt"
	"strings"

	"kubectlfzf/pkg/util"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// PodHeader is the header for pod files
const PodHeader = "Cluster Namespace Name PodIp HostIp NodeName Phase QOSClass Containers Tolerations Claims Age Labels FieldSelectors\n"

// Pod is the summary of a kubernetes pod
type Pod struct {
	ResourceMeta
	HostIP         string
	PodIP          string
	NodeName       string
	Tolerations    []string
	Containers     []string
	Claims         []string
	Phase          string
	FieldSelectors string
	QosClass       string
	Resource       string
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
	logrus.Debugf("Reading meta %#v", pod)
	p.FromObjectMeta(pod.ObjectMeta, config)
	p.HostIP = pod.Status.HostIP
	p.PodIP = pod.Status.PodIP
	spec := pod.Spec
	p.NodeName = spec.NodeName
	p.Phase = getPhase(pod)
	p.QosClass = string(pod.Status.QOSClass)

	fieldSelectors := make([]string, 0)
	if p.NodeName != "" {
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("spec.nodeName=%s", p.NodeName))
	}
	fieldSelectors = append(fieldSelectors, fmt.Sprintf("status.phase=%s", pod.Status.Phase))
	p.FieldSelectors = util.JoinSlicesOrNone(fieldSelectors, ",")

	containers := spec.Containers
	containers = append(containers, spec.InitContainers...)
	p.Containers = make([]string, len(containers))
	for k, v := range containers {
		p.Containers[k] = v.Name
	}

	volumes := spec.Volumes
	for _, v := range volumes {
		if v.PersistentVolumeClaim != nil {
			fullClaimName := fmt.Sprintf("%s/%s", p.ResourceMeta.Namespace,
				v.PersistentVolumeClaim.ClaimName)
			p.Claims = append(p.Claims, fullClaimName)
		}
	}
	tolerations := spec.Tolerations
	p.Tolerations = make([]string, 0)
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
		p.Tolerations = append(p.Tolerations, toleration)
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (p *Pod) HasChanged(k K8sResource) bool {
	oldPod := k.(*Pod)
	return (p.PodIP != oldPod.PodIP ||
		p.Phase != oldPod.Phase ||
		util.StringMapsEqual(p.Labels, oldPod.Labels) ||
		p.NodeName != oldPod.NodeName)
}

// ToString serializes the object to strings
func (p *Pod) ToString() string {
	lst := []string{
		p.Cluster,
		p.Namespace,
		p.Name,
		p.PodIP,
		p.HostIP,
		p.NodeName,
		p.Phase,
		p.QosClass,
		util.TruncateString(util.JoinSlicesOrNone(p.Containers, ","), 300),
		util.JoinSlicesOrNone(p.Tolerations, ","),
		util.JoinSlicesOrNone(p.Claims, ","),
		p.resourceAge(),
		p.labelsString(),
		p.FieldSelectors,
	}
	return util.DumpLine(lst)
}
