package resourcewatcher

import (
	"context"
	"regexp"
	"time"

	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/util"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"

	// Import for oidc auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/cache"
)

// ResourceWatcher contains rest clients for a given kubernetes context
type ResourceWatcher struct {
	namespaces  []string // List of namespaces filtered using excludedNamespaces
	cancelFuncs []context.CancelFunc
	storeConfig *store.StoreConfig

	watchedResourcesSet    map[resources.ResourceType]bool
	excludedResourcesSet   map[resources.ResourceType]bool
	excludedNamespaces     []*regexp.Regexp
	watchedNamespaces      []*regexp.Regexp
	namespacePollingPeriod time.Duration
	nodePollingPeriod      time.Duration
	ctorConfig             resources.CtorConfig

	//stores []*store.Store
}

// WatchConfig provides the configuration to watch a specific kubernetes resource
type WatchConfig struct {
	resourceType      resources.ResourceType
	getter            cache.Getter
	runtimeObject     runtime.Object
	hasNamespace      bool
	splitByNamespaces bool
	pollingPeriod     time.Duration
}

// NewResourceWatcher creates a new resource watcher on a given cluster
func NewResourceWatcher(cluster string, resourceWatcherCli ResourceWatcherCli, storeConfig *store.StoreConfig) (*ResourceWatcher, error) {
	excludedNamespaces, err := util.StringSliceToRegexps(resourceWatcherCli.excludedNamespaces)
	if err != nil {
		return nil, err
	}
	watchedNamespaces, err := util.StringSliceToRegexps(resourceWatcherCli.watchedNamespaces)
	if err != nil {
		return nil, err
	}
	ignoredNodeRoles := util.StringSliceToSet(resourceWatcherCli.ignoredNodeRoles)
	excludedResources, err := resources.GetResourceSetFromSlice(resourceWatcherCli.excludedResources)
	if err != nil {
		return nil, err
	}
	watchedResources, err := resources.GetResourceSetFromSlice(resourceWatcherCli.watchedResources)
	if err != nil {
		return nil, err
	}
	resourceWatcher := ResourceWatcher{
		storeConfig:            storeConfig,
		excludedResourcesSet:   excludedResources,
		watchedResourcesSet:    watchedResources,
		excludedNamespaces:     excludedNamespaces,
		watchedNamespaces:      watchedNamespaces,
		nodePollingPeriod:      resourceWatcherCli.nodePollingPeriod,
		namespacePollingPeriod: resourceWatcherCli.namespacePollingPeriod,
		ctorConfig: resources.CtorConfig{
			IgnoredNodeRoles: ignoredNodeRoles,
			Cluster:          cluster,
		},
	}
	return &resourceWatcher, nil
}

// Start begins the watch/poll of a given k8s resource
func (r *ResourceWatcher) Start(parentCtx context.Context, cfg WatchConfig) *store.Store {
	ctx, cancel := context.WithCancel(parentCtx)
	r.cancelFuncs = append(r.cancelFuncs, cancel)
	store := store.NewStore(ctx, r.storeConfig, r.ctorConfig, cfg.resourceType)
	if cfg.pollingPeriod > 0 {
		go r.pollResource(ctx, cfg, store)
	} else if cfg.splitByNamespaces {
		logrus.Infof("Starting watcher for ns %v, resource %s", r.namespaces, cfg.resourceType)
		go r.watchResource(ctx, cfg, store, r.namespaces)
	} else {
		go r.watchResource(ctx, cfg, store, []string{""})
	}
	return store
}

// Stop closes the watch/poll process of a k8s resource
func (r *ResourceWatcher) Stop() {
	logrus.Infof("Stopping %d resource watcher", len(r.cancelFuncs))
	for _, cancel := range r.cancelFuncs {
		cancel()
	}
}

// GetWatchConfigs creates the list of k8s to watch
func (r *ResourceWatcher) GetWatchConfigs() ([]WatchConfig, error) {
	clientset, err := r.storeConfig.GetClientset()
	if err != nil {
		return nil, err
	}
	coreGetter := clientset.CoreV1().RESTClient()
	appsGetter := clientset.AppsV1().RESTClient()
	autoscalingGetter := clientset.AutoscalingV1().RESTClient()
	networkingGetter := clientset.NetworkingV1().RESTClient()
	batchGetter := clientset.BatchV1().RESTClient()
	allWatchConfigs := []WatchConfig{
		{resources.ResourceTypePod, coreGetter, &corev1.Pod{}, true, false, 0},
		{resources.ResourceTypeConfigMap, coreGetter, &corev1.ConfigMap{}, true, false, 0},
		{resources.ResourceTypeService, coreGetter, &corev1.Service{}, true, false, 0},
		{resources.ResourceTypeServiceAccount, coreGetter, &corev1.ServiceAccount{}, true, false, 0},
		{resources.ResourceTypeReplicaSet, appsGetter, &appsv1.ReplicaSet{}, true, false, 0},
		{resources.ResourceTypeDaemonSet, appsGetter, &appsv1.DaemonSet{}, true, false, 0},
		{resources.ResourceTypeSecret, coreGetter, &corev1.Secret{}, true, false, 0},
		{resources.ResourceTypeStatefulSet, appsGetter, &appsv1.StatefulSet{}, true, false, 0},
		{resources.ResourceTypeDeployment, appsGetter, &appsv1.Deployment{}, true, false, 0},
		{resources.ResourceTypeEndpoints, coreGetter, &corev1.Endpoints{}, true, false, 0},
		{resources.ResourceTypeIngress, networkingGetter, &networkingv1.Ingress{}, true, false, 0},
		{resources.ResourceTypeCronJob, batchGetter, &batchv1.CronJob{}, true, false, 0},
		{resources.ResourceTypeJob, batchGetter, &batchv1.Job{}, true, false, 0},
		{resources.ResourceTypeHorizontalPodAutoscaler, autoscalingGetter, &autoscalingv1.HorizontalPodAutoscaler{}, true, false, 0},
		{resources.ResourceTypePersistentVolume, coreGetter, &corev1.PersistentVolume{}, false, false, 0},
		{resources.ResourceTypePersistentVolumeClaim, coreGetter, &corev1.PersistentVolumeClaim{}, true, false, 0},
		{resources.ResourceTypeNode, coreGetter, &corev1.Node{}, false, false, r.nodePollingPeriod},
		{resources.ResourceTypeNamespace, coreGetter, &corev1.Namespace{}, false, false, r.namespacePollingPeriod},
	}
	watchConfigs := []WatchConfig{}
	for _, w := range allWatchConfigs {
		if _, ok := r.excludedResourcesSet[w.resourceType]; ok {
			continue
		}
		_, ok := r.watchedResourcesSet[w.resourceType]
		if len(r.watchedResourcesSet) > 0 && !ok {
			continue
		}
		watchConfigs = append(watchConfigs, w)
	}
	return watchConfigs, nil
}

func (r *ResourceWatcher) doPoll(cacheListWatch *cache.ListWatch, store *store.Store) {
	obj, err := cacheListWatch.List(metav1.ListOptions{})
	if err != nil {
		logrus.Warningf("Error on listing resource: %v", err)
	}
	lst, err := apimeta.ExtractList(obj)
	if err != nil {
		logrus.Warningf("Error extracting list: %v", err)
	}
	store.AddResourceList(lst)
}

// FetchNamespaces gets the list of namespace from the cluster and fill
// the resource watcher with an initial list of namespaces
func (r *ResourceWatcher) FetchNamespaces(ctx context.Context) error {
	clientset, err := r.storeConfig.GetClientset()
	if err != nil {
		return err
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, namespace := range namespaces.Items {
		namespaceName := namespace.GetName()
		if util.IsStringMatching(namespaceName, r.excludedNamespaces) {
			logrus.Infof("namespace %s is in excluded namespaces, excluding", namespaceName)
			continue
		}
		if len(r.watchedNamespaces) > 0 && !util.IsStringMatching(namespaceName, r.watchedNamespaces) {
			logrus.Infof("namespace %s not in watched namespace, excluding", namespaceName)
			continue
		}
		r.namespaces = append(r.namespaces, namespaceName)
	}
	logrus.Infof("Fetched %d namespaces", len(r.namespaces))
	return nil
}

// DumpAPIResources dumps api resources file
func (r *ResourceWatcher) DumpAPIResources() error {
	destFile := r.storeConfig.GetResourceStorePath(resources.ResourceTypeApiResource)
	clientset, err := r.storeConfig.GetClientset()
	if err != nil {
		return err
	}
	resourceLists, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		return err
	}
	res := map[string]resources.K8sResource{}
	for _, resourceList := range resourceLists {
		a := resources.APIResourceList{}
		a.FromRuntime(resourceList, r.ctorConfig)
		res[resourceList.GroupVersion] = &a
	}
	err = util.EncodeToFile(res, destFile)
	return err
}

func (r *ResourceWatcher) getCacheListWatch(cfg WatchConfig, store *store.Store, namespace string) *cache.ListWatch {
	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Everything().String()
		options.ResourceVersion = "0"
	}
	cacheListWatch := cache.NewFilteredListWatchFromClient(cfg.getter,
		cfg.resourceType.String(), namespace, optionsModifier)
	return cacheListWatch
}

func (r *ResourceWatcher) pollResource(ctx context.Context,
	cfg WatchConfig, store *store.Store) {
	logrus.Infof("Start poller for %s", cfg.resourceType)
	cacheListWatch := r.getCacheListWatch(cfg, store, "")
	r.doPoll(cacheListWatch, store)
	ticker := time.NewTicker(cfg.pollingPeriod)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("Exiting poll of %s", cfg.resourceType)
			return
		case <-ticker.C:
			r.doPoll(cacheListWatch, store)
		}
	}
}

func (r *ResourceWatcher) startWatch(cfg WatchConfig,
	store *store.Store, namespace string, stop chan struct{}) {
	cacheListWatch := r.getCacheListWatch(cfg, store, namespace)
	resourceHandlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    store.AddResource,
		DeleteFunc: store.DeleteResource,
		UpdateFunc: store.UpdateResource,
	}
	controller := cache.NewSharedInformer(
		cacheListWatch,
		cfg.runtimeObject,
		// No resync
		time.Second*0,
	)
	controller.AddEventHandler(resourceHandlers)
	watchErrorHandler := func(r *cache.Reflector, err error) {
		if errors.IsForbidden(err) {
			logrus.Warnf("Resource %s is forbidden, stopping watcher. err: %s", cfg.resourceType, err)
			close(stop)
		}
	}
	controller.SetWatchErrorHandler(watchErrorHandler)
	controller.Run(stop)
}

func (r *ResourceWatcher) watchResource(ctx context.Context,
	cfg WatchConfig, store *store.Store, namespaces []string) {
	logrus.Infof("Start watch for %s on namespace %s", cfg.resourceType, namespaces)
	stop := make(chan struct{})
	for _, ns := range namespaces {
		go r.startWatch(cfg, store, ns, stop)
	}
	<-ctx.Done()
	logrus.Infof("Exiting watch of %s namespace %s", cfg.resourceType, namespaces)
	close(stop)
}
