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
	switch r {
	case ResourceTypeNode:
		return false
	case ResourceTypeService:
		return true
	case ResourceTypeEndpoints:
		return true
	case ResourceTypePersistentVolumeClaim:
		return true
	case ResourceTypeSecret:
		return true
	case ResourceTypeConfigMap:
		return true
	case ResourceTypeNamespace:
		return false
	case ResourceTypeServiceAccount:
		return true
	case ResourceTypePersistentVolume:
		return false
	case ResourceTypePod:
		return true
	case ResourceTypeDaemonSet:
		return true
	case ResourceTypeReplicaSet:
		return true
	case ResourceTypeStatefulSet:
		return true
	case ResourceTypeDeployment:
		return true
	case ResourceTypeHorizontalPodAutoscaler:
		return true
	case ResourceTypeJob:
		return true
	case ResourceTypeCronJob:
		return true
	case ResourceTypeApiResource:
		return false
	}
	return false
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
	case "no":
		fallthrough
	case "node":
		fallthrough
	case "nodes":
		return ResourceTypeNode
	case "svc":
		fallthrough
	case "service":
		fallthrough
	case "services":
		return ResourceTypeService
	case "ep":
		fallthrough
	case "endpoint":
		fallthrough
	case "endpoints":
		return ResourceTypeEndpoints
	case "pvc":
		fallthrough
	case "persistentvolumeclaim":
		fallthrough
	case "persistentvolumeclaims":
		return ResourceTypePersistentVolumeClaim
	case "secret":
		fallthrough
	case "secrets":
		return ResourceTypeSecret
	case "cm":
		fallthrough
	case "configmap":
		fallthrough
	case "configmaps":
		return ResourceTypeConfigMap
	case "ns":
		fallthrough
	case "namespace":
		fallthrough
	case "namespaces":
		return ResourceTypeNamespace
	case "sa":
		fallthrough
	case "serviceaccount":
		fallthrough
	case "serviceaccounts":
		return ResourceTypeServiceAccount
	case "pv":
		fallthrough
	case "persistentvolume":
		fallthrough
	case "persistentvolumes":
		return ResourceTypePersistentVolume
	case "po":
		fallthrough
	case "pod":
		fallthrough
	case "pods":
		return ResourceTypePod
	case "ds":
		fallthrough
	case "daemonset":
		fallthrough
	case "daemonsets":
		return ResourceTypeDaemonSet
	case "rs":
		fallthrough
	case "replicaset":
		fallthrough
	case "replicasets":
		return ResourceTypeReplicaSet
	case "sts":
		fallthrough
	case "statefulset":
		fallthrough
	case "statefulsets":
		return ResourceTypeStatefulSet
	case "deploy":
		fallthrough
	case "deployment":
		fallthrough
	case "deployments":
		return ResourceTypeDeployment
	case "hpa":
		fallthrough
	case "horizontalpodautoscaler":
		fallthrough
	case "horizontalpodautoscalers":
		return ResourceTypeHorizontalPodAutoscaler
	case "job":
		fallthrough
	case "jobs":
		return ResourceTypeJob
	case "cj":
		fallthrough
	case "cronjob":
		fallthrough
	case "cronjobs":
		return ResourceTypeCronJob
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
	logrus.Debugf("Getting resource type from %s, '%s', %d", cmdUse, args, len(args))
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
