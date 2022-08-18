package resources

import (
	"fmt"

	"kubectlfzf/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

// Service is the summary of a kubernetes service
type Service struct {
	ResourceMeta
	ServiceType string
	ClusterIP   string
	Ports       []string
	Selectors   []string
}

// NewServiceFromRuntime builds a pod from informer result
func NewServiceFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	s := &Service{}
	s.FromRuntime(obj, config)
	return s
}

// FromRuntime builds object from the informer's result
func (s *Service) FromRuntime(obj interface{}, config CtorConfig) {
	service := obj.(*corev1.Service)
	s.FromObjectMeta(service.ObjectMeta, config)
	s.ServiceType = string(service.Spec.Type)
	s.ClusterIP = service.Spec.ClusterIP
	s.Ports = make([]string, len(service.Spec.Ports))
	for k, v := range service.Spec.Ports {
		if v.NodePort > 0 {
			s.Ports[k] = fmt.Sprintf("%s:%d/%d", v.Name, v.Port, v.NodePort)
		} else {
			s.Ports[k] = fmt.Sprintf("%s:%d", v.Name, v.Port)
		}
	}
	s.Selectors = util.JoinStringMap(service.Spec.Selector, ExcludedLabels, "=")
}

// HasChanged returns true if the resource's dump needs to be updated
func (s *Service) HasChanged(k K8sResource) bool {
	oldService := k.(*Service)
	return (util.StringSlicesEqual(s.Ports, oldService.Ports) ||
		util.StringSlicesEqual(s.Selectors, oldService.Selectors) ||
		util.StringMapsEqual(s.Labels, oldService.Labels))
}

// ToString serializes the object to strings
func (s *Service) ToStrings() []string {
	portList := util.JoinSlicesOrNone(s.Ports, ",")
	selectorList := util.JoinSlicesOrNone(s.Selectors, ",")
	line := []string{
		s.Cluster,
		s.Namespace,
		s.Name,
		s.ServiceType,
		s.ClusterIP,
		portList,
		selectorList,
		s.resourceAge(),
		s.labelsString(),
	}
	return util.DumpLines(line)
}
