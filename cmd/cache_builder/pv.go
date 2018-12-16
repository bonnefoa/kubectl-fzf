package main

import (
	corev1 "k8s.io/api/core/v1"
)

const PersistentVolumeHeader = "Name Status StorageClass Zone Claim Age Labels\n"

// PersistentVolume is the summary of a kubernetes physical volume
type PersistentVolume struct {
	ResourceMeta
	status       string
	claim        string
	zone         string
	spec         string
	storageClass string
}

// NewPersistentVolumetime builds a pod from informer result
func NewPersistentVolumeFromRuntime(obj interface{}) K8sResource {
	p := &PersistentVolume{}
	p.FromRuntime(obj)
	return p
}

// FromRuntime builds object from the informer's result
func (pv *PersistentVolume) FromRuntime(obj interface{}) {
	pvFromRuntime := obj.(*corev1.PersistentVolume)
	pv.FromObjectMeta(pvFromRuntime.ObjectMeta)
	pv.status = string(pvFromRuntime.Status.Phase)
	var ok bool
	pv.zone, ok = pv.labels["failure-domain.beta.kubernetes.io/zone"]
	if !ok {
		pv.zone = "None"
	}
	pv.storageClass = pvFromRuntime.Spec.StorageClassName
	pv.claim = "None"
	if pvFromRuntime.Spec.ClaimRef != nil {
		pv.claim = pvFromRuntime.Spec.ClaimRef.Name
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (pv *PersistentVolume) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (pv *PersistentVolume) ToString() string {
	lst := []string{
		pv.name,
		pv.status,
		pv.storageClass,
		pv.zone,
		pv.claim,
		pv.resourceAge(),
		pv.labelsString(),
	}
	return DumpLine(lst)
}
