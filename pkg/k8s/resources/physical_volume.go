package resources

import (
	"fmt"

	"kubectlfzf/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

// PersistentVolumeHeader is the header for pvc csv
const PersistentVolumeHeader = "Cluster Name Status StorageClass Zone Claim Volume Affinities Age Labels\n"

// PersistentVolume is the summary of a kubernetes persistent volume
type PersistentVolume struct {
	ResourceMeta
	Status       string
	Claim        string
	Volume       string
	Zone         string
	Spec         string
	Affinities   []string
	StorageClass string
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
	pv.Status = string(pvFromRuntime.Status.Phase)
	var ok bool
	pv.Zone, ok = pv.Labels["failure-domain.beta.kubernetes.io/zone"]
	if !ok {
		pv.Zone = "None"
	}
	spec := pvFromRuntime.Spec
	if spec.AWSElasticBlockStore != nil {
		pv.Volume = util.LastURLPart(spec.AWSElasticBlockStore.VolumeID)
	} else if spec.GCEPersistentDisk != nil {
		pv.Volume = spec.GCEPersistentDisk.PDName
	}
	pv.StorageClass = spec.StorageClassName
	pv.Claim = "None"
	if pvFromRuntime.Spec.ClaimRef != nil {
		pv.Claim = fmt.Sprintf("%s/%s", spec.ClaimRef.Namespace, spec.ClaimRef.Name)
	}
	if pvFromRuntime.Spec.NodeAffinity != nil {
		for _, term := range pvFromRuntime.Spec.NodeAffinity.Required.NodeSelectorTerms {
			for _, expression := range term.MatchExpressions {
				affinity := fmt.Sprintf("%s:%s:%s", expression.Key,
					expression.Operator, util.JoinSlicesOrNone(expression.Values, ";"))
				pv.Affinities = append(pv.Affinities, affinity)
			}
		}
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (pv *PersistentVolume) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (pv *PersistentVolume) ToStrings() []string {
	lst := []string{
		pv.Cluster,
		pv.Name,
		pv.Status,
		pv.StorageClass,
		pv.Zone,
		pv.Claim,
		pv.Volume,
		util.JoinSlicesOrNone(pv.Affinities, ","),
		pv.resourceAge(),
		pv.labelsString(),
	}
	return util.DumpLine(lst)
}
