package resources

import (
	"sort"
	"time"

	"kubectlfzf/pkg/util"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// K8sResource is the generic information of a k8s entity
type K8sResource interface {
	GetNamespace() string
	GetLabels() map[string]string
	GetFieldSelectors() map[string]string

	HasChanged(k K8sResource) bool
	ToStrings() []string
	FromRuntime(obj interface{}, config CtorConfig)
}

// ResourceMeta is the generic information of a k8s entity
type ResourceMeta struct {
	Name         string
	Namespace    string // Namespace can be None
	Labels       map[string]string
	CreationTime time.Time
}

func (r *ResourceMeta) GetNamespace() string {
	return r.Namespace
}

func (r *ResourceMeta) GetFieldSelectors() map[string]string {
	return nil
}

func (r *ResourceMeta) GetLabels() map[string]string {
	return r.Labels
}

// FromObjectMeta copies meta information to the object
func (r *ResourceMeta) FromObjectMeta(meta metav1.ObjectMeta, config CtorConfig) {
	r.Name = meta.Name
	r.Namespace = meta.Namespace
	r.Labels = meta.Labels
	r.CreationTime = meta.CreationTimestamp.Time
}

// FromDynamicMeta copies meta information to the object
func (r *ResourceMeta) FromDynamicMeta(u *unstructured.Unstructured, config CtorConfig) {
	metadata := u.Object["metadata"].(map[string]interface{})
	r.Name = metadata["name"].(string)
	r.Namespace = metadata["namespace"].(string)
	var err error
	var found bool
	r.Labels, found, err = unstructured.NestedStringMap(u.Object, "metadata", "labels")
	util.FatalIf(err)
	if !found {
		logrus.Debugf("metadata.labels was not found in %#v", u.Object)
	}
	r.CreationTime, err = time.Parse(time.RFC3339, metadata["creationTimestamp"].(string))
	util.FatalIf(err)
}

func (r *ResourceMeta) resourceAge() string {
	return util.TimeToAge(r.CreationTime)
}

// ExcludedLabels is a list of excluded label/selector from the dump
var ExcludedLabels = map[string]string{"pod-template-generation": "",
	"app.kubernetes.io/name": "", "controller-revision-hash": "",
	"app.kubernetes.io/managed-by": "", "pod-template-hash": "",
	"statefulset.kubernetes.io/pod-name": "",
	"controler-uid":                      ""}

func (r *ResourceMeta) labelsString() string {
	if len(r.Labels) == 0 {
		return "None"
	}
	els := util.JoinStringMap(r.Labels, ExcludedLabels, "=")
	sort.Strings(els)
	return util.JoinSlicesOrNone(els, ",")
}

func ResourceToHeader(r ResourceType) string {
	replicaSetHeader := "Namespace\tName\tReplicas\tAvailableReplicas\tReadyReplicas\tSelector\tAge\tLabels"
	apiResourceHeader := "Name\tShortnames\tApiVersion\tNamespaced\tKind"
	configMapHeader := "Namespace\tName\tAge\tLabels"
	cronJobHeader := "Namespace\tName\tSchedule\tLastSchedule\tContainers\tAge\tLabels"
	daemonSetHeader := "Namespace\tName\tDesired\tCurrent\tReady\tLabelSelector\tContainers\tAge\tLabels"
	deploymentHeader := "Namespace\tName\tDesired\tCurrent\tUp-to-date\tAvailable\tAge\tLabels"
	endpointsHeader := "Namespace\tName\tAge\tReadyIps\tReadyPods\tNotReadyIps\tNotReadyPods\tLabels"
	horizontalPodAutoscalerHeader := "Namespace\tName\tReference\tTargets\tMinPods\tMaxPods\tReplicas\tAge\tLabels"
	ingressHeader := "Namespace\tName\tAddress\tAge\tLabels"
	jobHeader := "Namespace\tName\tCompletions\tContainers\tAge\tLabels"
	namespaceHeader := "Name\tAge\tLabels"
	nodeHeader := "Name\tRoles\tStatus\tInstanceType\tZone\tInternalIp\tTaints\tInstanceID\tAge\tLabels"
	podHeader := "Namespace\tName\tPodIp\tHostIp\tNodeName\tPhase\tQOSClass\tContainers\tTolerations\tClaims\tAge\tLabels"
	persistentVolumeHeader := "Name\tStatus\tStorageClass\tZone\tClaim\tVolume\tAffinities\tAge\tLabels"
	persistentVolumeClaimHeader := "Namespace\tName\tStatus\tCapacity\tVolumeName\tStorageClass\tAge\tLabels"
	secretHeader := "Namespace\tName\tType\tData\tAge\tLabels"
	serviceHeader := "Namespace\tName\tType\tClusterIp\tPorts\tSelector\tAge\tLabels"
	serviceAccountHeader := "Namespace\tName\tSecrets\tAge\tLabels"
	statefulSetHeader := "Namespace\tName\tReplicas\tSelector\tAge\tLabels"
	switch r {
	case ResourceTypeApiResource:
		return apiResourceHeader
	case ResourceTypeConfigMap:
		return configMapHeader
	case ResourceTypeCronJob:
		return cronJobHeader
	case ResourceTypeDaemonSet:
		return daemonSetHeader
	case ResourceTypeDeployment:
		return deploymentHeader
	case ResourceTypeEndpoints:
		return endpointsHeader
	case ResourceTypeHorizontalPodAutoscaler:
		return horizontalPodAutoscalerHeader
	case ResourceTypeIngress:
		return ingressHeader
	case ResourceTypeJob:
		return jobHeader
	case ResourceTypeNamespace:
		return namespaceHeader
	case ResourceTypeNode:
		return nodeHeader
	case ResourceTypePod:
		return podHeader
	case ResourceTypePersistentVolume:
		return persistentVolumeHeader
	case ResourceTypePersistentVolumeClaim:
		return persistentVolumeClaimHeader
	case ResourceTypeReplicaSet:
		return replicaSetHeader
	case ResourceTypeSecret:
		return secretHeader
	case ResourceTypeService:
		return serviceHeader
	case ResourceTypeServiceAccount:
		return serviceAccountHeader
	case ResourceTypeStatefulSet:
		return statefulSetHeader
	default:
		return "Unknown"
	}
}
