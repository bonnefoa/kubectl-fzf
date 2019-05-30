package k8sresources

import (
	"sort"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sResource is the generic information of a k8s entity
type K8sResource interface {
	HasChanged(k K8sResource) bool
	ToString() string
	FromRuntime(obj interface{}, config CtorConfig)
}

// ResourceMeta is the generic information of a k8s entity
type ResourceMeta struct {
	name         string
	namespace    string
	labels       map[string]string
	creationTime time.Time
}

// FromObjectMeta copies meta information to the object
func (r *ResourceMeta) FromObjectMeta(meta metav1.ObjectMeta) {
	r.name = meta.Name
	r.namespace = meta.Namespace
	r.labels = meta.Labels
	r.creationTime = meta.CreationTimestamp.Time
}

func (r *ResourceMeta) resourceAge() string {
	return util.TimeToAge(r.creationTime)
}

// ExcludedLabels is a list of excluded label/selector from the dump
var ExcludedLabels = map[string]string{"pod-template-generation": "",
	"app.kubernetes.io/name": "", "controller-revision-hash": "",
	"app.kubernetes.io/managed-by": "", "pod-template-hash": "",
	"statefulset.kubernetes.io/pod-name": "",
	"controler-uid":                      ""}

func (r *ResourceMeta) labelsString() string {
	if len(r.labels) == 0 {
		return "None"
	}
	els := util.JoinStringMap(r.labels, ExcludedLabels, "=")
	sort.Strings(els)
	return util.JoinSlicesOrNone(els, ",")
}
