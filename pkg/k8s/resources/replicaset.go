package resources

import (
	"strconv"

	"kubectlfzf/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
)

// ReplicaSet is the summary of a kubernetes replicaSet
type ReplicaSet struct {
	ResourceMeta
	Replicas          string
	ReadyReplicas     string
	AvailableReplicas string
	Selectors         []string
}

// NewReplicaSetFromRuntime builds a k8sresource from informer result
func NewReplicaSetFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	p := &ReplicaSet{}
	p.FromRuntime(obj, config)
	return p
}

// FromRuntime builds object from the informer's result
func (r *ReplicaSet) FromRuntime(obj interface{}, config CtorConfig) {
	replicaSet := obj.(*appsv1.ReplicaSet)
	r.FromObjectMeta(replicaSet.ObjectMeta, config)
	r.Replicas = strconv.Itoa(int(replicaSet.Status.Replicas))
	r.ReadyReplicas = strconv.Itoa(int(replicaSet.Status.ReadyReplicas))
	r.AvailableReplicas = strconv.Itoa(int(replicaSet.Status.AvailableReplicas))
	r.Selectors = util.JoinStringMap(replicaSet.Spec.Selector.MatchLabels,
		ExcludedLabels, "=")
}

// HasChanged returns true if the resource'r dump needs to be updated
func (r *ReplicaSet) HasChanged(k K8sResource) bool {
	oldRs := k.(*ReplicaSet)
	return (r.Replicas != oldRs.Replicas ||
		r.ReadyReplicas != oldRs.ReadyReplicas ||
		r.AvailableReplicas != oldRs.AvailableReplicas ||
		util.StringSlicesEqual(r.Selectors, oldRs.Selectors) ||
		util.StringMapsEqual(r.Labels, oldRs.Labels))
}

// ToString serializes the object to strings
func (r *ReplicaSet) ToStrings() []string {
	selectorList := util.JoinSlicesOrNone(r.Selectors, ",")
	line := []string{
		r.Namespace,
		r.Name,
		r.Replicas,
		r.AvailableReplicas,
		r.ReadyReplicas,
		selectorList,
		r.resourceAge(),
		r.labelsString(),
	}
	return util.DumpLines(line)
}
