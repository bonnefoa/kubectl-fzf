package resources

import (
	"fmt"
	"strings"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// Pod is the summary of a kubernetes pod
type Pod struct {
	ResourceMeta
	HostIP      string
	PodIP       string
	NodeName    string
	Tolerations []string
	Containers  []string
	Claims      []string
	Phase       string
	QosClass    string
	Resource    string
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
	logrus.Tracef("Reading meta %#v", pod)
	p.FromObjectMeta(pod.ObjectMeta, config)
	p.HostIP = pod.Status.HostIP
	p.PodIP = pod.Status.PodIP
	spec := pod.Spec
	p.NodeName = spec.NodeName
	p.Phase = getPhase(pod)
	p.QosClass = string(pod.Status.QOSClass)

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

func (p *Pod) GetFieldSelectors() map[string]string {
	return map[string]string{
		"spec.nodeName": p.NodeName,
		"status.phase":  p.Phase}
}

// ToString serializes the object to strings
func (p *Pod) ToStrings() []string {
	lst := []string{
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
	}
	return util.DumpLines(lst)
}
