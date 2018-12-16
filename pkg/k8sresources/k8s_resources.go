package k8sresources

import (
	"fmt"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sResource is the generic information of a k8s entity
type K8sResource interface {
	HasChanged(k K8sResource) bool
	ToString() string
	FromRuntime(obj interface{})
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
	duration := time.Now().Sub(r.creationTime)
	duration = duration.Round(time.Minute)
	if duration.Hours() > 30 {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
	hour := duration / time.Hour
	duration -= hour * time.Hour
	minute := duration / time.Minute
	return fmt.Sprintf("%02d:%02d", hour, minute)
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
	return util.JoinSlicesOrNone(els, ",")
}
