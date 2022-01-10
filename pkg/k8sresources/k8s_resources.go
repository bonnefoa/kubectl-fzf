package k8sresources

import (
	"sort"
	"time"

	"kubectlfzf/pkg/util"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// K8sResource is the generic information of a k8s entity
type K8sResource interface {
	HasChanged(k K8sResource) bool
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
		glog.V(3).Infof("metadata.labels was not found in %#v", u.Object)
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
