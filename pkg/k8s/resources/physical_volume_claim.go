package resources

import (
	"kubectlfzf/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

// PersistentVolumeClaim is the summary of a kubernetes physical volume claim
type PersistentVolumeClaim struct {
	ResourceMeta
	Status       string
	VolumeName   string
	Capacity     string
	StorageClass string
}

// NewPersistentVolumeClaimFromRuntime builds a pod from informer result
func NewPersistentVolumeClaimFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	p := &PersistentVolumeClaim{}
	p.FromRuntime(obj, config)
	return p
}

// FromRuntime builds object from the informer's result
func (pvc *PersistentVolumeClaim) FromRuntime(obj interface{}, config CtorConfig) {
	pvcFromRuntime := obj.(*corev1.PersistentVolumeClaim)
	pvc.FromObjectMeta(pvcFromRuntime.ObjectMeta, config)
	pvc.Status = string(pvcFromRuntime.Status.Phase)
	if pvcFromRuntime.Spec.StorageClassName != nil {
		pvc.StorageClass = *pvcFromRuntime.Spec.StorageClassName
	}
	pvc.VolumeName = pvcFromRuntime.Spec.VolumeName
	quantity := pvcFromRuntime.Status.Capacity["storage"]
	pvc.Capacity = quantity.String()
}

// HasChanged returns true if the resource's dump needs to be updated
func (pvc *PersistentVolumeClaim) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (pvc *PersistentVolumeClaim) ToStrings() []string {
	lst := []string{
		pvc.Namespace,
		pvc.Name,
		pvc.Status,
		pvc.Capacity,
		pvc.VolumeName,
		pvc.StorageClass,
		pvc.resourceAge(),
		pvc.labelsString(),
	}
	return util.DumpLines(lst)
}
