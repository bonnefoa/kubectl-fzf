package resources

import (
	"fmt"
	"kubectlfzf/pkg/util"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

const HorizontalPodAutoscalerHeader = "Cluster Namespace Name Reference Targets MinPods MaxPods Replicas Age Labels\n"

// HorizontalPodAutoscaler is the summary of a kubernetes horizontal pod autoscaler
type HorizontalPodAutoscaler struct {
	ResourceMeta
	Reference       string
	Targets         string
	MinPods         string
	MaxPods         string
	CurrentReplicas string
}

// NewHorizontalPodAutoscalerFromRuntime builds a pod from informer result
func NewHorizontalPodAutoscalerFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	h := &HorizontalPodAutoscaler{}
	h.FromRuntime(obj, config)
	return h
}

// FromRuntime builds object from the informer's result
func (h *HorizontalPodAutoscaler) FromRuntime(obj interface{}, config CtorConfig) {
	hpa := obj.(*autoscalingv1.HorizontalPodAutoscaler)
	h.FromObjectMeta(hpa.ObjectMeta, config)
	h.Reference = fmt.Sprintf("%s/%s",
		hpa.Spec.ScaleTargetRef.Kind,
		hpa.Spec.ScaleTargetRef.Name)
	h.MinPods = "None"
	if hpa.Spec.MinReplicas != nil {
		h.MinPods = fmt.Sprintf("%d", *hpa.Spec.MinReplicas)
	}
	h.MaxPods = fmt.Sprintf("%d", hpa.Spec.MaxReplicas)
	h.CurrentReplicas = fmt.Sprintf("%d", hpa.Status.CurrentReplicas)
}

// HasChanged returns true if the resource'h dump needs to be updated
func (h *HorizontalPodAutoscaler) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (h *HorizontalPodAutoscaler) ToStrings() []string {
	line := []string{
		h.Cluster,
		h.Namespace,
		h.Name,
		h.Reference,
		h.Targets,
		h.MinPods,
		h.MaxPods,
		h.CurrentReplicas,
		h.resourceAge(),
		h.labelsString(),
	}
	return util.DumpLine(line)
}
