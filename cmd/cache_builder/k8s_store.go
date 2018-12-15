package main

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// K8sStore stores the current state of k8s resources
type K8sStore struct {
	data         map[string]K8sResource
	resourceCtor func() K8sResource
	resourceName string
	destFile     string
	currentFile  *os.File
}

// NewK8sStore creates a new store
func NewK8sStore(resourceCtor func() K8sResource,
	resourceName string, cacheDir string) (K8sStore, error) {
	k := K8sStore{}
	destFile := path.Join(cacheDir, resourceName)
	err := os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		return k, errors.Wrapf(err, "Error creating directory %s", cacheDir)
	}
	currentFile, err := os.Create(destFile)
	if err != nil {
		return k, errors.Wrapf(err, "Error creating file %s", destFile)
	}
	k.data = make(map[string]K8sResource, 0)
	k.resourceCtor = resourceCtor
	k.resourceName = resourceName
	k.destFile = destFile
	k.currentFile = currentFile
	currentFile.WriteString(k.resourceCtor().Header())
	return k, nil
}

// AddResource adds a new k8s object to the store
func (k *K8sStore) AddResource(obj interface{}) {
	key := resourceKey(obj)
	newObj := k.resourceCtor()
	newObj.FromRuntime(obj)
	glog.V(9).Infof("%s added: %s", k.resourceName, key)
	k.data[key] = newObj

	err := k.AppendNewObject(newObj)
	if err != nil {
		glog.Warningf("Error when appending new object to current state: %v", err)
	}
}

// DeleteResource removes an existing k8s object to the store
func (k *K8sStore) DeleteResource(obj interface{}) {
	key := resourceKey(obj)
	glog.V(9).Infof("%s deleted: %s", k.resourceName, key)
	delete(k.data, key)

	err := k.DumpFullState()
	if err != nil {
		glog.Warningf("Error when dumping state: %v", err)
	}
}

// UpdateResource update an existing k8s object
func (k *K8sStore) UpdateResource(oldObj, newObj interface{}) {
	key := resourceKey(newObj)
	k8sObj := k.resourceCtor()
	k8sObj.FromRuntime(newObj)
	if k8sObj.HasChanged(k.data[key]) {
		glog.V(9).Infof("%s changed: %s", k.resourceName, key)
		k.data[key] = k8sObj
		err := k.DumpFullState()
		if err != nil {
			glog.Warningf("Error when dumping state: %v", err)
		}
	}
}

// AppendNewObject appends a new object to the cache dump
func (k *K8sStore) AppendNewObject(resource K8sResource) error {
	_, err := k.currentFile.WriteString(resource.ToString())
	if err != nil {
		return err
	}
	err = k.currentFile.Sync()
	return err
}

// DumpFullState writes the full state to the cache file
func (k *K8sStore) DumpFullState() error {
	glog.V(8).Infof("Doing full dump %d %s", len(k.data), k.resourceName)
	tempFilePath := fmt.Sprintf("%s_", k.destFile)
	tempFile, err := os.Create(tempFilePath)
	glog.V(12).Infof("Creating temp file for full state %s", tempFile.Name())
	if err != nil {
		return errors.Wrapf(err, "Error creating temp file %s", tempFilePath)
	}
	w := bufio.NewWriter(tempFile)
	w.WriteString(k.resourceCtor().Header())
	for _, v := range k.data {
		_, err := w.WriteString(v.ToString())
		if err != nil {
			return errors.Wrapf(err, "Error writing bytes to file %s", tempFilePath)
		}
	}
	err = w.Flush()
	if err != nil {
		return errors.Wrapf(err, "Error flushing buffer")
	}

	err = tempFile.Sync()
	if err != nil {
		return errors.Wrapf(err, "Error syncing file")
	}

	glog.V(17).Infof("Closing file %s", k.currentFile.Name())
	k.currentFile.Close()
	err = os.Rename(tempFilePath, k.destFile)
	if err != nil {
		return errors.Wrapf(err, "Error moving file from %s to %s",
			tempFilePath, k.destFile)
	}
	k.currentFile = tempFile
	return nil
}
