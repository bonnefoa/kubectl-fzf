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

func ResourceToHeader(r ResourceType) []string {
	replicaSetHeader := []string{"Namespace", "Name", "Replicas", "AvailableReplicas", "ReadyReplicas", "Selector", "Age", "Labels"}
	apiResourceHeader := []string{"Name", "Shortnames", "ApiVersion", "Namespaced", "Kind"}
	configMapHeader := []string{"Namespace", "Name", "Age", "Labels"}
	cronJobHeader := []string{"Namespace", "Name", "Schedule", "LastSchedule", "Containers", "Age", "Labels"}
	daemonSetHeader := []string{"Namespace", "Name", "Desired", "Current", "Ready", "LabelSelector", "Containers", "Age", "Labels"}
	deploymentHeader := []string{"Namespace", "Name", "Desired", "Current", "Up-to-date", "Available", "Age", "Labels"}
	endpointsHeader := []string{"Namespace", "Name", "Age", "ReadyIps", "ReadyPods", "NotReadyIps", "NotReadyPods", "Labels"}
	horizontalPodAutoscalerHeader := []string{"Namespace", "Name", "Reference", "Targets", "MinPods", "MaxPods", "Replicas", "Age", "Labels"}
	ingressHeader := []string{"Namespace", "Name", "Address", "Age", "Labels"}
	jobHeader := []string{"Namespace", "Name", "Completions", "Containers", "Age", "Labels"}
	namespaceHeader := []string{"Name", "Age", "Labels"}
	nodeHeader := []string{"Name", "Roles", "Status", "InstanceType", "Zone", "InternalIp", "Taints", "InstanceID", "Age", "Labels"}
	podHeader := []string{"Namespace", "Name", "PodIp", "HostIp", "NodeName", "Phase", "QOSClass", "Containers", "Tolerations", "Claims", "Age", "Labels"}
	persistentVolumeHeader := []string{"Name", "Status", "StorageClass", "Zone", "Claim", "Volume", "Affinities", "Age", "Labels"}
	persistentVolumeClaimHeader := []string{"Namespace", "Name", "Status", "Capacity", "VolumeName", "StorageClass", "Age", "Labels"}
	secretHeader := []string{"Namespace", "Name", "Type", "Data", "Age", "Labels"}
	serviceHeader := []string{"Namespace", "Name", "Type", "ClusterIp", "Ports", "Selector", "Age", "Labels"}
	serviceAccountHeader := []string{"Namespace", "Name", "Secrets", "Age", "Labels"}
	statefulSetHeader := []string{"Namespace", "Name", "Replicas", "Selector", "Age", "Labels"}
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
		return []string{"Unknown"}
	}
}
