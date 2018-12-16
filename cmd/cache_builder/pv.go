package main

import (
	corev1 "k8s.io/api/core/v1"
)

// PersistentVolume is the summary of a kubernetes physical volume
type PersistentVolume struct {
	ResourceMeta
	status       string
	claim        string
	zone         string
	spec         string
	storageClass string
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

// Header generates the csv header for the resource
func (pv *PersistentVolume) Header() string {
	return "Name Status StorageClass Zone Claim Age Labels\n"
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
