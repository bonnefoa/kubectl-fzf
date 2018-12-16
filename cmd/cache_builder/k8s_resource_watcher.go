package main

import (
	"context"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/k8sresources"
	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/tools/cache"
)

type resourceWatcher struct {
	clientset *kubernetes.Clientset
	namespace string
	cluster   string
}

type watchConfig struct {
	resourceCtor  func(obj interface{}) k8sresources.K8sResource
	header        string
	resourceName  string
	getter        cache.Getter
	runtimeObject runtime.Object
	hasNamespace  bool
	pollingPeriod time.Duration
}

func NewResourceWatcher(namespace string) resourceWatcher {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	util.FatalIf(err)

	resourceWatcher := resourceWatcher{}
	resourceWatcher.namespace = namespace
	resourceWatcher.clientset, err = kubernetes.NewForConfig(config)

	util.FatalIf(err)
	return resourceWatcher
}

func (r *resourceWatcher) pollResource(ctx context.Context,
	cfg watchConfig, k8sStore K8sStore) {
	glog.V(4).Infof("Start poller for %s on namespace %s", k8sStore.resourceName, r.namespace)
	namespace := ""
	if cfg.hasNamespace {
		namespace = r.namespace
	}
	watchlist := cache.NewListWatchFromClient(cfg.getter,
		k8sStore.resourceName, namespace, fields.Everything())
	for {
		obj, err := watchlist.List(metav1.ListOptions{})
		if err != nil {
			glog.Warningf("Error on listing %s: %v", k8sStore.resourceName, err)
		}
		lst, err := apimeta.ExtractList(obj)
		if err != nil {
			glog.Warningf("Error extracting list %s: %v", k8sStore.resourceName, err)
		}
		k8sStore.AddResourceList(lst)
		time.Sleep(cfg.pollingPeriod)
	}
}

func (r *resourceWatcher) watchResource(ctx context.Context,
	cfg watchConfig, k8sStore K8sStore) {
	glog.V(4).Infof("Start watch for %s on namespace %s", k8sStore.resourceName, r.namespace)
	namespace := ""
	if cfg.hasNamespace {
		namespace = r.namespace
	}
	watchlist := cache.NewListWatchFromClient(cfg.getter,
		k8sStore.resourceName, namespace, fields.Everything())

	_, controller := cache.NewInformer(
		watchlist, cfg.runtimeObject, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    k8sStore.AddResource,
			DeleteFunc: k8sStore.DeleteResource,
			UpdateFunc: k8sStore.UpdateResource,
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)
	<-ctx.Done()
	close(stop)
}
