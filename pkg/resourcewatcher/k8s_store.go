package resourcewatcher

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/k8sresources"
	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// K8sStore stores the current state of k8s resources
type K8sStore struct {
	data         map[string]k8sresources.K8sResource
	resourceCtor func(obj interface{}, config k8sresources.CtorConfig) k8sresources.K8sResource
	ctorConfig   k8sresources.CtorConfig
	resourceName string
	destFileName string
	currentFile  *os.File
	lastFullDump time.Time
	storeConfig  StoreConfig
	firstWrite   bool
}

// StoreConfig defines parameters used for the cache location
type StoreConfig struct {
	Cluster             string
	CacheDir            string
	TimeBetweenFullDump time.Duration
}

// NewK8sStore creates a new store
func NewK8sStore(cfg WatchConfig, storeConfig StoreConfig, ctorConfig k8sresources.CtorConfig, namespace string) (K8sStore, error) {
	k := K8sStore{}
	destFileName := util.GetDestFileName(storeConfig.CacheDir, storeConfig.Cluster, cfg.resourceName)
	currentFile, err := ioutil.TempFile("", cfg.resourceName)
	if err != nil {
		return k, errors.Wrapf(err, "Error creating file for %s", cfg.resourceName)
	}
	k.data = make(map[string]k8sresources.K8sResource, 0)
	k.resourceCtor = cfg.resourceCtor
	k.resourceName = cfg.resourceName
	k.destFileName = destFileName
	if namespace != "" {
		k.destFileName = fmt.Sprintf("%s_ns_%s", destFileName, namespace)
	}
	k.currentFile = currentFile
	k.lastFullDump = time.Time{}
	k.storeConfig = storeConfig
	k.firstWrite = true
	k.ctorConfig = ctorConfig

	util.WriteHeaderFile(cfg.header, destFileName)

	return k, nil
}

func resourceKey(obj interface{}) string {
	name := "None"
	namespace := "None"
	switch v := obj.(type) {
	case metav1.ObjectMetaAccessor:
		o := v.GetObjectMeta()
		namespace = o.GetNamespace()
		name = o.GetName()
	case *unstructured.Unstructured:
		metadata := v.Object["metadata"].(map[string]interface{})
		name = metadata["name"].(string)
		namespace = metadata["namespace"].(string)
	default:
		glog.Warningf("Unknown type %v", obj)
	}
	return fmt.Sprintf("%s_%s", namespace, name)
}

// AddResourceList clears current state add the objects to the store.
// It will trigger a full dump
func (k *K8sStore) AddResourceList(lstRuntime []runtime.Object) {
	k.data = make(map[string]k8sresources.K8sResource, 0)
	for _, runtimeObject := range lstRuntime {
		key := resourceKey(runtimeObject)
		resource := k.resourceCtor(runtimeObject, k.ctorConfig)
		k.data[key] = resource
	}
	glog.Infof("Writing new state of %s", k.destFileName)
	err := k.DumpFullState()
	if err != nil {
		glog.Warningf("Error when dumping state: %v", err)
	}
}

// AddResource adds a new k8s object to the store
func (k *K8sStore) AddResource(obj interface{}) {
	key := resourceKey(obj)
	newObj := k.resourceCtor(obj, k.ctorConfig)
	glog.V(11).Infof("%s added: %s", k.resourceName, key)
	k.data[key] = newObj

	err := k.AppendNewObject(newObj)
	if err != nil {
		glog.Warningf("Error when appending new object to current state: %v", err)
	}
}

// DeleteResource removes an existing k8s object to the store
func (k *K8sStore) DeleteResource(obj interface{}) {
	key := "Unknown"
	switch v := obj.(type) {
	case cache.DeletedFinalStateUnknown:
		key = resourceKey(v.Obj)
	case unstructured.Unstructured:
	case metav1.ObjectMetaAccessor:
		key = resourceKey(obj)
	default:
		glog.V(6).Infof("Unknown object type %v", obj)
	}
	glog.V(11).Infof("%s deleted: %s", k.resourceName, key)
	delete(k.data, key)

	err := k.DumpFullState()
	if err != nil {
		glog.Warningf("Error when dumping state: %v", err)
	}
}

// UpdateResource update an existing k8s object
func (k *K8sStore) UpdateResource(oldObj, newObj interface{}) {
	key := resourceKey(newObj)
	k8sObj := k.resourceCtor(newObj, k.ctorConfig)
	if k8sObj.HasChanged(k.data[key]) {
		glog.V(11).Infof("%s changed: %s", k.resourceName, key)
		k.data[key] = k8sObj
		err := k.DumpFullState()
		if err != nil {
			glog.Warningf("Error when dumping state: %v", err)
		}
	}
}

// AppendNewObject appends a new object to the cache dump
func (k *K8sStore) AppendNewObject(resource k8sresources.K8sResource) error {
	if k.firstWrite {
		k.firstWrite = false
		err := os.Rename(k.currentFile.Name(), k.destFileName)
		if err != nil {
			return err
		}
		glog.Infof("Initial write of %s", k.destFileName)
	}
	_, err := k.currentFile.WriteString(resource.ToString())
	if err != nil {
		return err
	}
	return nil
}

func (k *K8sStore) writeDataInFile(tempFile *os.File) error {
	w := bufio.NewWriter(tempFile)
	keys := make([]string, len(k.data))
	i := 0
	for key := range k.data {
		keys[i] = key
		i = i + 1
	}
	sort.Strings(keys)
	for _, key := range keys {
		v := k.data[key]
		_, err := w.WriteString(v.ToString())
		if err != nil {
			return errors.Wrapf(err, "Error writing bytes to file %s",
				tempFile.Name())
		}
	}
	err := w.Flush()
	if err != nil {
		return errors.Wrapf(err, "Error flushing buffer")
	}
	err = tempFile.Sync()
	if err != nil {
		return errors.Wrapf(err, "Error syncing file")
	}
	return nil
}

// DumpFullState writes the full state to the cache file
func (k *K8sStore) DumpFullState() error {
	now := time.Now()
	delta := now.Sub(k.lastFullDump)
	if delta < k.storeConfig.TimeBetweenFullDump {
		glog.V(10).Infof("Last full dump for %s happened %s ago, ignoring it", k.resourceName, delta)
		return nil
	}
	k.lastFullDump = now
	glog.V(8).Infof("Doing full dump %d %s", len(k.data), k.resourceName)
	tempFile, err := ioutil.TempFile("", k.resourceName)
	if err != nil {
		return errors.Wrapf(err, "Error creating temp file for resource %s",
			k.resourceName)
	}
	glog.V(12).Infof("Created temp file for full state %s", tempFile.Name())
	err = k.writeDataInFile(tempFile)
	if err != nil {
		return err
	}

	glog.V(17).Infof("Closing file %s", k.currentFile.Name())
	k.currentFile.Close()
	err = os.Rename(tempFile.Name(), k.destFileName)
	if err != nil {
		return errors.Wrapf(err, "Error moving file from %s to %s",
			tempFile.Name(), k.destFileName)
	}
	k.currentFile = tempFile
	return nil
}
