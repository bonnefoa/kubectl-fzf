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
	object       K8sResource
	resourceName string
	destFile     string
	currentFile  *os.File
}

// NewK8sStore creates a new store
func NewK8sStore(object K8sResource, resourceName string, cacheDir string) (K8sStore, error) {
	k8sStore := K8sStore{}
	destFile := path.Join(cacheDir, resourceName)
	err := os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		return k8sStore, errors.Wrapf(err, "Error creating directory %s", cacheDir)
	}
	currentFile, err := os.Create(destFile)
	if err != nil {
		return k8sStore, errors.Wrapf(err, "Error creating file %s", destFile)
	}
	k8sStore.data = make(map[string]K8sResource, 0)
	k8sStore.object = object
	k8sStore.resourceName = resourceName
	k8sStore.destFile = destFile
	k8sStore.currentFile = currentFile
	return k8sStore, nil
}

// AddResource adds a new k8s object to the store
func (k8sStore K8sStore) AddResource(obj interface{}) {
	key := resourceKey(obj)
	newObj := k8sStore.object
	newObj.FromRuntime(obj)
	glog.V(9).Infof("Object added: %s", key)
	k8sStore.data[key] = newObj

	err := k8sStore.AppendNewObject(newObj)
	if err != nil {
		glog.Warningf("Error when appending new object to current state: %v", err)
	}
}

// DeleteResource removes an existing k8s object to the store
func (k8sStore K8sStore) DeleteResource(obj interface{}) {
	key := resourceKey(obj)
	glog.V(9).Infof("Object deleted: %s", key)
	delete(k8sStore.data, key)

	err := k8sStore.DumpFullState()
	if err != nil {
		glog.Warningf("Error when dumping state: %v", err)
	}
}

// UpdateResource update an existing k8s object
func (k8sStore K8sStore) UpdateResource(oldObj, newObj interface{}) {
	key := resourceKey(newObj)
	glog.V(9).Infof("Object changed: %s", key)
	k8sObj := k8sStore.object
	k8sObj.FromRuntime(newObj)
	k8sStore.data[key] = k8sObj
}

// AppendNewObject appends a new object to the cache dump
func (k8sStore K8sStore) AppendNewObject(resource K8sResource) error {
	_, err := k8sStore.currentFile.WriteString(resource.ToString())
	if err != nil {
		return err
	}
	err = k8sStore.currentFile.Sync()
	return err
}

// DumpFullState writes the full state to the cache file
func (k8sStore K8sStore) DumpFullState() error {
	tempFilePath := fmt.Sprintf("%s_", k8sStore.destFile)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return errors.Wrapf(err, "Error creating temp file %s", tempFilePath)
	}
	w := bufio.NewWriter(tempFile)
	for _, v := range k8sStore.data {
		_, err := w.WriteString(v.ToString())
		if err != nil {
			return errors.Wrapf(err, "Error writing bytes to file %s", tempFilePath)
		}
	}
	err = w.Flush()
	if err != nil {
		return errors.Wrapf(err, "Error flushing buffer")
	}
	k8sStore.currentFile.Close()
	err = os.Rename(tempFilePath, k8sStore.destFile)
	if err != nil {
		return errors.Wrapf(err, "Error moving file from %s to %s",
			tempFilePath, k8sStore.destFile)
	}
	k8sStore.currentFile = tempFile
	return nil
}
