package k8sresources

import (
	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	betav1 "k8s.io/api/extensions/v1beta1"
)

const IngressHeader = "Namespace Name Address Age Labels\n"

// Ingress is the summary of a kubernetes ingress
type Ingress struct {
	ResourceMeta
	address []string
}

// NewIngressFromRuntime builds a pod from informer result
func NewIngressFromRuntime(obj interface{}) K8sResource {
	p := &Ingress{}
	p.FromRuntime(obj)
	return p
}

// FromRuntime builds object from the informer's result
func (ingress *Ingress) FromRuntime(obj interface{}) {
	ingressFromRuntime := obj.(*betav1.Ingress)
	ingress.FromObjectMeta(ingressFromRuntime.ObjectMeta)
	for _, lb := range ingressFromRuntime.Status.LoadBalancer.Ingress {
		ingress.address = append(ingress.address, lb.Hostname)
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (ingress *Ingress) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (ingress *Ingress) ToString() string {
	addressList := util.JoinSlicesOrNone(ingress.address, ",")
	lst := []string{
		ingress.namespace,
		ingress.name,
		addressList,
		ingress.resourceAge(),
		ingress.labelsString(),
	}
	return util.DumpLine(lst)
}
