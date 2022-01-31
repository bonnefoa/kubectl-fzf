package k8sresources

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
	case "apiresource":
		return ResourceTypeApiResource
	case "configmaps":
	case "configmap":
		return ResourceTypeConfigMap
	case "cronjobs":
	case "cronjob":
		return ResourceTypeCronJob
	case "daemonsets":
	case "daemonset":
		return ResourceTypeDaemonSet
	case "deployments":
	case "deployment":
		return ResourceTypeDeployment
	case "endpoints":
	case "endpoint":
		return ResourceTypeEndpoints
	case "horizontalpodautoscalers":
	case "horizontalpodautoscaler":
	case "hpas":
	case "hpa":
		return ResourceTypeHorizontalPodAutoscaler
	case "ingress":
		return ResourceTypeIngress
	case "job":
	case "jobs":
		return ResourceTypeJob
	case "namespace":
	case "namespaces":
	case "ns":
		return ResourceTypeNamespace
	case "node":
	case "nodes":
		return ResourceTypeNode
	case "pods":
	case "pod":
	case "p":
		return ResourceTypePod
	case "persistentvolumes":
	case "persistentvolume":
	case "pv":
		return ResourceTypePersistentVolume
	case "persistentvolumeclaims":
	case "persistentvolumeclaim":
	case "pvc":
		return ResourceTypePersistentVolumeClaim
	case "replicasets":
	case "replicaset":
		return ResourceTypeReplicaSet
	case "secrets":
	case "secret":
		return ResourceTypeSecret
	case "services":
	case "service":
		return ResourceTypeService
	case "serviceaccounts":
	case "serviceaccount":
	case "sa":
		return ResourceTypeServiceAccount
	case "statefulsets":
	case "statefulset":
	case "sts":
		return ResourceTypeStatefulSet
	}
	return ResourceTypeUnknown
}
