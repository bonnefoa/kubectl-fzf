package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/tools/cache"
)

var (
	kubeconfig string
	namespace  string
	cacheDir   string
)

func init() {
	if home := os.Getenv("HOME"); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&namespace, "namespace", "",
		"Namespace to watch, empty for all namespaces")
	flag.StringVar(&cacheDir, "dir", os.Getenv("KUBECTL_FZF_CACHE"),
		"Cache dir location. Default to KUBECTL_FZF_CACHE env var")
}

type resourceWatcher struct {
	clientset *kubernetes.Clientset
	namespace string
	cluster   string
}

func fatalIf(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err)
		os.Exit(-1)
	}
}

func (r *resourceWatcher) watchResource(ctx context.Context, getter cache.Getter,
	k8sStore K8sStore, runtimeObject runtime.Object) {
	glog.V(4).Infof("Start watch for %s on namespace %s", k8sStore.resourceName,
		r.namespace)
	watchlist := cache.NewListWatchFromClient(getter,
		k8sStore.resourceName, r.namespace, fields.Everything())

	_, controller := cache.NewInformer(
		watchlist, runtimeObject, time.Second*0,
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

func handleSignals(cancel context.CancelFunc) {
	sigIn := make(chan os.Signal, 100)
	signal.Notify(sigIn)
	for sig := range sigIn {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			glog.Errorf("Caught signal '%s' (%d); terminating.", sig, sig)
			cancel()
		}
	}
}

type storeConfig struct {
	resourceCtor  func() K8sResource
	resourceName  string
	getter        cache.Getter
	runtimeObject runtime.Object
}

func main() {
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	fatalIf(err)

	ctx, cancel := context.WithCancel(context.Background())

	go handleSignals(cancel)

	resourceWatcher := resourceWatcher{}
	resourceWatcher.namespace = namespace
	resourceWatcher.clientset, err = kubernetes.NewForConfig(config)
	fatalIf(err)

	coreGetter := resourceWatcher.clientset.Core().RESTClient()
	appsGetter := resourceWatcher.clientset.Apps().RESTClient()

	storeConfigs := []storeConfig{
		storeConfig{func() K8sResource { return &Pod{} }, string(corev1.ResourcePods), coreGetter, &corev1.Pod{}},
		storeConfig{func() K8sResource { return &Service{} }, string(corev1.ResourceServices), coreGetter, &corev1.Service{}},
		storeConfig{func() K8sResource { return &ReplicaSet{} }, "replicasets", appsGetter, &appsv1.ReplicaSet{}},
		storeConfig{func() K8sResource { return &ConfigMap{} }, "configmaps", coreGetter, &corev1.ConfigMap{}},
		storeConfig{func() K8sResource { return &StatefulSet{} }, "statefulsets", appsGetter, &appsv1.StatefulSet{}},
		storeConfig{func() K8sResource { return &Deployment{} }, "deployments", appsGetter, &appsv1.Deployment{}},
		storeConfig{func() K8sResource { return &Endpoints{} }, "endpoints", coreGetter, &corev1.Endpoints{}},
	}

	for _, storeConfig := range storeConfigs {
		store, err := NewK8sStore(storeConfig.resourceCtor, storeConfig.resourceName, cacheDir)
		fatalIf(err)
		go resourceWatcher.watchResource(ctx, storeConfig.getter, store, storeConfig.runtimeObject)
	}

	<-ctx.Done()
}
