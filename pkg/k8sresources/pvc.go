package k8sresources

import (
	corev1 "k8s.io/api/core/v1"
)

const PersistentVolumeClaimHeader = "Cluster Namespace Name Status Capacity VolumeName StorageClass Age Labels\n"

// PersistentVolumeClaim is the summary of a kubernetes physical volume claim
type PersistentVolumeClaim struct {
	ResourceMeta
	status       string
	volumeName   string
	capacity     string
	storageClass string
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
	pvc.status = string(pvcFromRuntime.Status.Phase)
	if pvcFromRuntime.Spec.StorageClassName != nil {
		pvc.storageClass = *pvcFromRuntime.Spec.StorageClassName
	}
	pvc.volumeName = pvcFromRuntime.Spec.VolumeName
	quantity := pvcFromRuntime.Status.Capacity["storage"]
	pvc.capacity = quantity.String()
}

// HasChanged returns true if the resource's dump needs to be updated
func (pvc *PersistentVolumeClaim) HasChanged(k K8sResource) bool {
	return true
}
