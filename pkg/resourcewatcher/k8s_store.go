package resourcewatcher

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type LabelKey struct {
	Namespace string
	Label     string
}

type LabelPair struct {
	Key         LabelKey
	Occurrences int
}

type LabelPairList []LabelPair

func (p LabelPairList) Len() int { return len(p) }
func (p LabelPairList) Less(i, j int) bool {
	if p[i].Occurrences == p[j].Occurrences {
		if p[i].Key.Namespace == p[j].Key.Namespace {
			return p[i].Key.Label < p[j].Key.Label
		}
		return p[i].Key.Namespace < p[j].Key.Namespace
	}
	return p[i].Occurrences > p[j].Occurrences
}
func (p LabelPairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// K8sStore stores the current state of k8s resources
type K8sStore struct {
	data         map[string]k8sresources.K8sResource
	labelMap     map[LabelKey]int
	resourceCtor func(obj interface{}, config k8sresources.CtorConfig) k8sresources.K8sResource
	ctorConfig   k8sresources.CtorConfig
	resourceType k8sresources.ResourceType
	currentFile  *os.File
	storeConfig  *k8sresources.StoreConfig
	firstWrite   bool

	dataMutex  sync.Mutex
	labelMutex sync.Mutex
	fileMutex  sync.Mutex

	labelToDump   bool
	lastFullDump  time.Time
	lastLabelDump time.Time
}

// NewK8sStore creates a new store
func NewK8sStore(ctx context.Context, cfg WatchConfig, storeConfig *k8sresources.StoreConfig, ctorConfig k8sresources.CtorConfig) *K8sStore {
	k := K8sStore{}
	k.data = make(map[string]k8sresources.K8sResource, 0)
	k.labelMap = make(map[LabelKey]int, 0)
	k.resourceCtor = cfg.resourceCtor
	k.resourceType = cfg.resourceType
	k.currentFile = nil
	k.storeConfig = storeConfig
	k.firstWrite = true
	k.ctorConfig = ctorConfig
	k.lastLabelDump = time.Time{}
	k.lastFullDump = time.Time{}

	return &k
}

func resourceKey(obj interface{}) (string, string, map[string]string) {
	name := "None"
	namespace := "None"
	var labels map[string]string
	switch v := obj.(type) {
	case metav1.ObjectMetaAccessor:
		o := v.GetObjectMeta()
		namespace = o.GetNamespace()
		name = o.GetName()
		labels = o.GetLabels()
	case *unstructured.Unstructured:
		metadata := v.Object["metadata"].(map[string]interface{})
		name = metadata["name"].(string)
		namespace = metadata["namespace"].(string)
		labels = metadata["labels"].(map[string]string)
	default:
		glog.Warningf("Unknown type %v", obj)
	}
	return fmt.Sprintf("%s_%s", namespace, name), namespace, labels
}

func (k *K8sStore) periodicLabelDump(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.labelMutex.Lock()
			if k.labelToDump {
				k.dumpLabel()
			}
			k.labelMutex.Unlock()
		}
	}
}

func (k *K8sStore) resetLabelMap() {
	k.labelMutex.Lock()
	k.labelMap = make(map[LabelKey]int, 0)
	k.labelMutex.Unlock()
}

func (k *K8sStore) updateLabelMap(namespace string, labels map[string]string, delta int) {
	k.labelMutex.Lock()
	for labelKey, labelValue := range labels {
		k.labelMap[LabelKey{namespace, fmt.Sprintf("%s=%s", labelKey, labelValue)}] += delta
	}
	k.labelMutex.Unlock()
}

// AddResourceList clears current state add the objects to the store.
// It will trigger a full dump
// This is used for polled resources, no need for mutex
func (k *K8sStore) AddResourceList(lstRuntime []runtime.Object) {
	k.data = make(map[string]k8sresources.K8sResource, 0)
	k.resetLabelMap()
	for _, runtimeObject := range lstRuntime {
		key, ns, labels := resourceKey(runtimeObject)
		resource := k.resourceCtor(runtimeObject, k.ctorConfig)
		k.data[key] = resource
		k.updateLabelMap(ns, labels, 1)
	}
	err := k.DumpFullState()
	if err != nil {
		glog.Warningf("Error when dumping state: %v", err)
	}
}

// AddResource adds a new k8s object to the store
func (k *K8sStore) AddResource(obj interface{}) {
	key, ns, labels := resourceKey(obj)
	newObj := k.resourceCtor(obj, k.ctorConfig)
	glog.V(11).Infof("%s added: %s", k.resourceType, key)
	k.dataMutex.Lock()
	k.data[key] = newObj
	k.dataMutex.Unlock()
	k.updateLabelMap(ns, labels, 1)

	err := k.AppendNewObject(newObj)
	if err != nil {
		glog.Warningf("Error when appending new object to current state: %v", err)
	}
}

// DeleteResource removes an existing k8s object to the store
func (k *K8sStore) DeleteResource(obj interface{}) {
	key := "Unknown"
	ns := "Unknown"
	var labels map[string]string
	switch v := obj.(type) {
	case cache.DeletedFinalStateUnknown:
		key, ns, labels = resourceKey(v.Obj)
	case unstructured.Unstructured:
	case metav1.ObjectMetaAccessor:
		key, ns, labels = resourceKey(obj)
	default:
		glog.V(6).Infof("Unknown object type %v", obj)
		return
	}
	glog.V(11).Infof("%s deleted: %s", k.resourceType, key)
	k.dataMutex.Lock()
	delete(k.data, key)
	k.dataMutex.Unlock()
	k.updateLabelMap(ns, labels, -1)

	err := k.DumpFullState()
	if err != nil {
		glog.Warningf("Error when dumping state: %v", err)
	}
}

// UpdateResource update an existing k8s object
func (k *K8sStore) UpdateResource(oldObj, newObj interface{}) {
	key, _, _ := resourceKey(newObj)
	k8sObj := k.resourceCtor(newObj, k.ctorConfig)
	k.dataMutex.Lock()
	if k8sObj.HasChanged(k.data[key]) {
		glog.V(11).Infof("%s changed: %s", k.resourceType, key)
		k.data[key] = k8sObj
		k.dataMutex.Unlock()
		// TODO Handle label diff
		// k.updateLabelMap(ns, labels, 1)
		err := k.DumpFullState()
		if err != nil {
			glog.Warningf("Error when dumping state: %v", err)
		}
	} else {
		k.dataMutex.Unlock()
	}
}

// AppendNewObject appends a new object to the cache dump
func (k *K8sStore) AppendNewObject(resource k8sresources.K8sResource) error {
	//	k.fileMutex.Lock()
	//	if k.currentFile == nil {
	//		var err error
	//		err = util.WriteStringToFile(resource.ToString(), k.destDir, k.resourceType, "resource")
	//		if err != nil {
	//			k.fileMutex.Unlock()
	//			return err
	//		}
	//		err = k.reopenCurrentFile()
	//		if err != nil {
	//			k.fileMutex.Unlock()
	//			return err
	//		}
	//		glog.Infof("Initial write of %s", k.currentFile.Name())
	//	}
	//	_, err := k.currentFile.WriteString(resource.ToString())
	//	k.fileMutex.Unlock()
	//	if err != nil {
	//		return err
	//	}
	//
	//	now := time.Now()
	//	k.labelMutex.Lock()
	//	delta := now.Sub(k.lastLabelDump)
	//	if delta < time.Second {
	//		k.labelToDump = true
	//	}
	//	k.labelMutex.Unlock()
	//	return nil
	return nil
}

func (k *K8sStore) dumpLabel() error {
	glog.V(8).Infof("Dump of label file %s", k.resourceType)
	//k.lastLabelDump = time.Now()
	//labelOutput, err := k.generateLabel()
	//if err != nil {
	//return errors.Wrapf(err, "Error generating label output")
	//}
	//err = util.WriteStringToFile(labelOutput, k.destDir, k.resourceType, "label")
	//if err != nil {
	//return errors.Wrapf(err, "Error writing label file")
	//}
	//k.labelToDump = false
	return nil
}

func (k *K8sStore) generateLabel() (string, error) {
	k.labelMutex.Lock()
	pl := make(LabelPairList, len(k.labelMap))
	i := 0
	for key, occurrences := range k.labelMap {
		pl[i] = LabelPair{key, occurrences}
		i++
	}
	k.labelMutex.Unlock()
	sort.Sort(pl)
	var res strings.Builder
	for _, pair := range pl {
		var str string
		if pair.Key.Namespace == "" {
			str = fmt.Sprintf("%s %s %d\n",
				k.ctorConfig.Cluster, pair.Key.Label, pair.Occurrences)
		} else {
			str = fmt.Sprintf("%s %s %s %d\n",
				k.ctorConfig.Cluster, pair.Key.Namespace, pair.Key.Label, pair.Occurrences)
		}
		_, err := res.WriteString(str)
		if err != nil {
			return "", errors.Wrapf(err, "Error writing string %s",
				str)
		}
	}
	return strings.Trim(res.String(), "\n"), nil
}

// DumpFullState writes the full state to the cache file
func (k *K8sStore) DumpFullState() error {
	glog.V(8).Infof("Dump full state of %s", k.resourceType)
	now := time.Now()
	delta := now.Sub(k.lastFullDump)
	if delta < k.storeConfig.TimeBetweenFullDump {
		glog.V(10).Infof("Last full dump for %s happened %s ago, ignoring it", k.resourceType, delta)
		return nil
	}
	k.lastFullDump = now
	glog.V(8).Infof("Doing full dump %d %s", len(k.data), k.resourceType)

	destFile := k.storeConfig.GetFilePath(k.resourceType)
	err := util.EncodeToFile(k.data, destFile)
	return err

	// err = k.reopenCurrentFile()
	// if err != nil {
	// 	return err
	// }
	// labelOutput, err := k.generateLabel()
	// if err != nil {
	// 	return errors.Wrapf(err, "Error generating label output")
	// }
	// err = util.WriteStringToFile(labelOutput, k.destDir, k.resourceType, "label")
	// return err
}
