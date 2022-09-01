package resources

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type UnknownResourceError struct {
	ResourceStr string
}

func (u UnknownResourceError) Error() string {
	return fmt.Sprintf("Resource %s is unknown", u.ResourceStr)
}

type ResourceType int64

const (
	ResourceTypeApiResource ResourceType = iota
	ResourceTypeConfigMap
	ResourceTypeCronJob
	ResourceTypeDaemonSet
	ResourceTypeDeployment
	ResourceTypeEndpoints
	ResourceTypeHorizontalPodAutoscaler
	ResourceTypeIngress
	ResourceTypeJob
	ResourceTypeNamespace
	ResourceTypeNode
	ResourceTypePod
	ResourceTypePersistentVolume
	ResourceTypePersistentVolumeClaim
	ResourceTypeReplicaSet
	ResourceTypeSecret
	ResourceTypeService
	ResourceTypeServiceAccount
	ResourceTypeStatefulSet
	ResourceTypeUnknown
)

func (r ResourceType) IsNamespaced() bool {
	return true
}

func (r ResourceType) String() string {
	switch r {
	case ResourceTypeApiResource:
		return "apiresources"
	case ResourceTypeConfigMap:
		return "configmaps"
	case ResourceTypeCronJob:
		return "cronjobs"
	case ResourceTypeDaemonSet:
		return "daemonsets"
	case ResourceTypeDeployment:
		return "deployments"
	case ResourceTypeEndpoints:
		return "endpoints"
	case ResourceTypeHorizontalPodAutoscaler:
		return "horizontalpodautoscalers"
	case ResourceTypeIngress:
		return "ingresses"
	case ResourceTypeJob:
		return "jobs"
	case ResourceTypeNamespace:
		return "namespaces"
	case ResourceTypeNode:
		return "nodes"
	case ResourceTypePod:
		return "pods"
	case ResourceTypePersistentVolume:
		return "persistentvolumes"
	case ResourceTypePersistentVolumeClaim:
		return "persistentvolumeclaims"
	case ResourceTypeReplicaSet:
		return "replicasets"
	case ResourceTypeSecret:
		return "secrets"
	case ResourceTypeService:
		return "services"
	case ResourceTypeServiceAccount:
		return "serviceaccounts"
	case ResourceTypeStatefulSet:
		return "statefulsets"
	}
	return "unknown"
}

func ParseResourceType(s string) ResourceType {
	switch s {
	case "apiresources":
		fallthrough
	case "apiresource":
		return ResourceTypeApiResource
	case "cm":
		fallthrough
	case "configmaps":
		fallthrough
	case "configmap":
		return ResourceTypeConfigMap
	case "cronjobs":
		fallthrough
	case "cronjob":
		return ResourceTypeCronJob
	case "daemonsets":
		fallthrough
	case "daemonset":
		return ResourceTypeDaemonSet
	case "deployments":
		fallthrough
	case "deployment":
		return ResourceTypeDeployment
	case "endpoints":
		fallthrough
	case "endpoint":
		return ResourceTypeEndpoints
	case "horizontalpodautoscalers":
		fallthrough
	case "horizontalpodautoscaler":
		fallthrough
	case "hpas":
		fallthrough
	case "hpa":
		return ResourceTypeHorizontalPodAutoscaler
	case "ingress":
		return ResourceTypeIngress
	case "job":
		fallthrough
	case "jobs":
		return ResourceTypeJob
	case "namespace":
		fallthrough
	case "namespaces":
		fallthrough
	case "ns":
		return ResourceTypeNamespace
	case "node":
		fallthrough
	case "nodes":
		return ResourceTypeNode
	case "po":
		fallthrough
	case "pods":
		fallthrough
	case "pod":
		fallthrough
	case "p":
		return ResourceTypePod
	case "persistentvolumes":
		fallthrough
	case "persistentvolume":
		fallthrough
	case "pv":
		return ResourceTypePersistentVolume
	case "persistentvolumeclaims":
		fallthrough
	case "persistentvolumeclaim":
		fallthrough
	case "pvc":
		return ResourceTypePersistentVolumeClaim
	case "replicasets":
		fallthrough
	case "replicaset":
		return ResourceTypeReplicaSet
	case "secrets":
		fallthrough
	case "secret":
		return ResourceTypeSecret
	case "services":
		fallthrough
	case "service":
		return ResourceTypeService
	case "serviceaccounts":
		fallthrough
	case "serviceaccount":
		fallthrough
	case "sa":
		return ResourceTypeServiceAccount
	case "statefulsets":
		fallthrough
	case "statefulset":
		fallthrough
	case "statefulsets.apps":
		fallthrough
	case "sts":
		return ResourceTypeStatefulSet
	}
	return ResourceTypeUnknown
}

func GetResourceSetFromSlice(resourceSlice []string) (map[ResourceType]bool, error) {
	res := make(map[ResourceType]bool, 0)
	for _, resourceStr := range resourceSlice {
		r := ParseResourceType(resourceStr)
		if r == ResourceTypeUnknown {
			return nil, UnknownResourceError{resourceStr}
		}
		res[r] = true
	}
	return res, nil
}

func GetResourceType(cmdUse string, args []string) ResourceType {
	logrus.Debugf("Getting resource type from '%s', %d", args, len(args))
	resourceType := ResourceTypeApiResource
	if cmdUse == "logs" {
		return ResourceTypePod
	}
	if cmdUse == "exec" {
		return ResourceTypePod
	}
	// No resource type or we have only
	// get ''#
	if len(args) <= 1 {
		return resourceType
	}
	for _, arg := range args {
		resourceType = ParseResourceType(arg)
		if resourceType != ResourceTypeUnknown {
			return resourceType
		}
	}
	return ResourceTypeUnknown
}
