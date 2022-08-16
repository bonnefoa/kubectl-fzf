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
	GetCluster() string
	GetLabels() map[string]string
	GetFieldSelectors() map[string]string

	HasChanged(k K8sResource) bool
	ToStrings() []string
	FromRuntime(obj interface{}, config CtorConfig)
}

// ResourceMeta is the generic information of a k8s entity
type ResourceMeta struct {
	Cluster      string
	Name         string
	Namespace    string // Namespace can be None
	Labels       map[string]string
	CreationTime time.Time
}

func (r *ResourceMeta) GetCluster() string {
	return r.Cluster
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
	r.Cluster = config.Cluster
	r.Labels = meta.Labels
	r.CreationTime = meta.CreationTimestamp.Time
}

// FromDynamicMeta copies meta information to the object
func (r *ResourceMeta) FromDynamicMeta(u *unstructured.Unstructured, config CtorConfig) {
	metadata := u.Object["metadata"].(map[string]interface{})
	r.Name = metadata["name"].(string)
	r.Namespace = metadata["namespace"].(string)
	r.Cluster = config.Cluster
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
	switch r {
	case ResourceTypeApiResource:
		return APIResourceHeader
	case ResourceTypeConfigMap:
		return ConfigMapHeader
	case ResourceTypeCronJob:
		return CronJobHeader
	case ResourceTypeDaemonSet:
		return DaemonSetHeader
	case ResourceTypeDeployment:
		return DeploymentHeader
	case ResourceTypeEndpoints:
		return EndpointsHeader
	case ResourceTypeHorizontalPodAutoscaler:
		return HorizontalPodAutoscalerHeader
	case ResourceTypeIngress:
		return IngressHeader
	case ResourceTypeJob:
		return JobHeader
	case ResourceTypeNamespace:
		return NamespaceHeader
	case ResourceTypeNode:
		return NodeHeader
	case ResourceTypePod:
		return PodHeader
	case ResourceTypePersistentVolume:
		return PersistentVolumeHeader
	case ResourceTypePersistentVolumeClaim:
		return PersistentVolumeClaimHeader
	case ResourceTypeReplicaSet:
		return ReplicaSetHeader
	case ResourceTypeSecret:
		return SecretHeader
	case ResourceTypeService:
		return ServiceHeader
	case ResourceTypeServiceAccount:
		return ServiceAccountHeader
	case ResourceTypeStatefulSet:
		return StatefulSetHeader
	default:
		return "Unknown"
	}
}
