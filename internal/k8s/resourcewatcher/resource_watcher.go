package resourcewatcher

import (
	"context"
	"regexp"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"

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

	watchResourcesSet      map[resources.ResourceType]bool
	excludeResourcesSet    map[resources.ResourceType]bool
	excludeNamespaces      []*regexp.Regexp
	watchNamespaces        []*regexp.Regexp
	namespacePollingPeriod time.Duration
	nodePollingPeriod      time.Duration
	ctorConfig             resources.CtorConfig
	exitOnUnauthorized     bool
}

// WatchConfig provides the configuration to watch a specific kubernetes resource
type WatchConfig struct {
	resourceType  resources.ResourceType
	getter        cache.Getter
	runtimeObject runtime.Object
	hasNamespace  bool
	pollingPeriod time.Duration
}

// NewResourceWatcher creates a new resource watcher on a given cluster
func NewResourceWatcher(cluster string, resourceWatcherCli ResourceWatcherCli, storeConfig *store.StoreConfig) (*ResourceWatcher, error) {
	excludedNamespaces, err := util.StringSliceToRegexps(resourceWatcherCli.excludNamespaces)
	if err != nil {
		return nil, err
	}
	watchedNamespaces, err := util.StringSliceToRegexps(resourceWatcherCli.watchNamespaces)
	if err != nil {
		return nil, err
	}
	ignoredNodeRoles := util.StringSliceToSet(resourceWatcherCli.ignoreNodeRoles)
	excludedResources, err := resources.GetResourceSetFromSlice(resourceWatcherCli.excludResources)
	if err != nil {
		return nil, err
	}
	watchedResources, err := resources.GetResourceSetFromSlice(resourceWatcherCli.watchResources)
	if err != nil {
		return nil, err
	}
	resourceWatcher := ResourceWatcher{
		storeConfig:            storeConfig,
		excludeResourcesSet:    excludedResources,
		watchResourcesSet:      watchedResources,
		excludeNamespaces:      excludedNamespaces,
		watchNamespaces:        watchedNamespaces,
		nodePollingPeriod:      resourceWatcherCli.nodePollingPeriod,
		namespacePollingPeriod: resourceWatcherCli.namespacePollingPeriod,
		ctorConfig: resources.CtorConfig{
			IgnoredNodeRoles: ignoredNodeRoles,
		},
		exitOnUnauthorized: resourceWatcherCli.exitOnUnauthorized,
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
	} else {
		go r.watchResource(ctx, cfg, store, r.namespaces)
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
		{resources.ResourceTypePod, coreGetter, &corev1.Pod{}, true, 0},
		{resources.ResourceTypeConfigMap, coreGetter, &corev1.ConfigMap{}, true, 0},
		{resources.ResourceTypeService, coreGetter, &corev1.Service{}, true, 0},
		{resources.ResourceTypeServiceAccount, coreGetter, &corev1.ServiceAccount{}, true, 0},
		{resources.ResourceTypeReplicaSet, appsGetter, &appsv1.ReplicaSet{}, true, 0},
		{resources.ResourceTypeDaemonSet, appsGetter, &appsv1.DaemonSet{}, true, 0},
		{resources.ResourceTypeSecret, coreGetter, &corev1.Secret{}, true, 0},
		{resources.ResourceTypeStatefulSet, appsGetter, &appsv1.StatefulSet{}, true, 0},
		{resources.ResourceTypeDeployment, appsGetter, &appsv1.Deployment{}, true, 0},
		{resources.ResourceTypeEndpoints, coreGetter, &corev1.Endpoints{}, true, 0},
		{resources.ResourceTypeIngress, networkingGetter, &networkingv1.Ingress{}, true, 0},
		{resources.ResourceTypeCronJob, batchGetter, &batchv1.CronJob{}, true, 0},
		{resources.ResourceTypeJob, batchGetter, &batchv1.Job{}, true, 0},
		{resources.ResourceTypeHorizontalPodAutoscaler, autoscalingGetter, &autoscalingv1.HorizontalPodAutoscaler{}, true, 0},
		{resources.ResourceTypePersistentVolume, coreGetter, &corev1.PersistentVolume{}, false, 0},
		{resources.ResourceTypePersistentVolumeClaim, coreGetter, &corev1.PersistentVolumeClaim{}, true, 0},
		{resources.ResourceTypeNode, coreGetter, &corev1.Node{}, false, r.nodePollingPeriod},
		{resources.ResourceTypeNamespace, coreGetter, &corev1.Namespace{}, false, r.namespacePollingPeriod},
	}
	watchConfigs := []WatchConfig{}
	for _, w := range allWatchConfigs {
		if _, ok := r.excludeResourcesSet[w.resourceType]; ok {
			continue
		}
		_, ok := r.watchResourcesSet[w.resourceType]
		if len(r.watchResourcesSet) > 0 && !ok {
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
// This is only useful when we need to filter namespaces
func (r *ResourceWatcher) FetchNamespaces(ctx context.Context) error {
	if len(r.watchNamespaces) == 0 || len(r.excludeNamespaces) == 0 {
		// No need for namespace filtering
		return nil
	}

	clientset, err := r.storeConfig.GetClientset()
	if err != nil {
		return err
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		for _, watchNamespace := range r.watchNamespaces {
			r.namespaces = append(r.namespaces, watchNamespace.String())
		}
		logrus.Warnf("Failed to get the list of namespaces, will fallback to %s", r.namespaces)
		return nil
	}
	for _, namespace := range namespaces.Items {
		namespaceName := namespace.GetName()
		if util.IsStringMatching(namespaceName, r.excludeNamespaces) {
			logrus.Infof("namespace %s is in excluded namespaces, excluding", namespaceName)
			continue
		}
		if len(r.watchNamespaces) > 0 && !util.IsStringMatching(namespaceName, r.watchNamespaces) {
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
	watchErrorHandler := func(reflector *cache.Reflector, err error) {
		if errors.IsUnauthorized(err) && r.exitOnUnauthorized {
			logrus.Warnf("Resource %s is unauthorized, stopping watcher", cfg.resourceType)
			r.Stop()
		}
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
	stop := make(chan struct{})
	resourceType := cfg.resourceType
	isNamespaced := resourceType.IsNamespaced()
	if !isNamespaced {
		logrus.Infof("Resource %s is not Namespaced, will ignore namespace filters", resourceType)
	}
	if isNamespaced && len(namespaces) > 0 {
		logrus.Infof("Start watch for %s on namespace %s", resourceType, namespaces)
		for _, ns := range namespaces {
			go r.startWatch(cfg, store, ns, stop)
		}
	} else {
		logrus.Infof("Start watch for %s on all namespaces", resourceType)
		go r.startWatch(cfg, store, "", stop)
	}
	<-ctx.Done()
	logrus.Infof("Exiting watch of %s namespace %s", resourceType, namespaces)
	close(stop)
}
