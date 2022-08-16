package resources

// CtorConfig is the configuration passed to all resource constructors
type CtorConfig struct {
	IgnoredNodeRoles map[string]bool
	Cluster          string
}

type ResourceCtor func(obj interface{}, config CtorConfig) K8sResource

func ResourceTypeToCtor(resourceType ResourceType) ResourceCtor {
	switch resourceType {
	case ResourceTypePod:
		return NewPodFromRuntime
	case ResourceTypeConfigMap:
		return NewConfigMapFromRuntime
	case ResourceTypeService:
		return NewServiceFromRuntime
	case ResourceTypeServiceAccount:
		return NewServiceAccountFromRuntime
	case ResourceTypeReplicaSet:
		return NewReplicaSetFromRuntime
	case ResourceTypeDaemonSet:
		return NewDaemonSetFromRuntime
	case ResourceTypeSecret:
		return NewSecretFromRuntime
	case ResourceTypeStatefulSet:
		return NewStatefulSetFromRuntime
	case ResourceTypeDeployment:
		return NewDeploymentFromRuntime
	case ResourceTypeEndpoints:
		return NewEndpointsFromRuntime
	case ResourceTypeIngress:
		return NewIngressFromRuntime
	case ResourceTypeCronJob:
		return NewCronJobFromRuntime
	case ResourceTypeJob:
		return NewJobFromRuntime
	case ResourceTypeHorizontalPodAutoscaler:
		return NewHorizontalPodAutoscalerFromRuntime
	case ResourceTypePersistentVolume:
		return NewPersistentVolumeFromRuntime
	case ResourceTypePersistentVolumeClaim:
		return NewPersistentVolumeClaimFromRuntime
	case ResourceTypeNode:
		return NewNodeFromRuntime
	case ResourceTypeNamespace:
		return NewNamespaceFromRuntime
	}
	return nil
}
