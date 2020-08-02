package k8sresources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"kubectlfzf/pkg/util"
)

// PersistentVolumeHeader is the header for pvc csv
const PersistentVolumeHeader = "Name Status StorageClass Zone Claim Volume Affinities Age Labels\n"

// PersistentVolume is the summary of a kubernetes physical volume
type PersistentVolume struct {
	ResourceMeta
	status       string
	claim        string
	volume       string
	zone         string
	spec         string
	affinities   []string
	storageClass string
}

// NewPersistentVolumeFromRuntime builds a pod from informer result
func NewPersistentVolumeFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	p := &PersistentVolume{}
	p.FromRuntime(obj, config)
	return p
}

// FromRuntime builds object from the informer's result
func (pv *PersistentVolume) FromRuntime(obj interface{}, config CtorConfig) {
	pvFromRuntime := obj.(*corev1.PersistentVolume)
	pv.FromObjectMeta(pvFromRuntime.ObjectMeta, config)
	pv.status = string(pvFromRuntime.Status.Phase)
	var ok bool
	pv.zone, ok = pv.labels["failure-domain.beta.kubernetes.io/zone"]
	if !ok {
		pv.zone = "None"
	}
	spec := pvFromRuntime.Spec
	if spec.AWSElasticBlockStore != nil {
		pv.volume = util.LastURLPart(spec.AWSElasticBlockStore.VolumeID)
	} else if spec.GCEPersistentDisk != nil {
		pv.volume = spec.GCEPersistentDisk.PDName
	}
	pv.storageClass = spec.StorageClassName
	pv.claim = "None"
	if pvFromRuntime.Spec.ClaimRef != nil {
		pv.claim = fmt.Sprintf("%s/%s", spec.ClaimRef.Namespace, spec.ClaimRef.Name)
	}
	if pvFromRuntime.Spec.NodeAffinity != nil {
		for _, term := range pvFromRuntime.Spec.NodeAffinity.Required.NodeSelectorTerms {
			for _, expression := range term.MatchExpressions {
				affinity := fmt.Sprintf("%s:%s:%s", expression.Key,
					expression.Operator, util.JoinSlicesOrNone(expression.Values, ";"))
				pv.affinities = append(pv.affinities, affinity)
			}
		}
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (pv *PersistentVolume) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (pv *PersistentVolume) ToString() string {
	lst := []string{
		pv.cluster,
		pv.name,
		pv.status,
		pv.storageClass,
		pv.zone,
		pv.claim,
		pv.volume,
		util.JoinSlicesOrNone(pv.affinities, ","),
		pv.resourceAge(),
		pv.labelsString(),
	}
	return util.DumpLine(lst)
}
