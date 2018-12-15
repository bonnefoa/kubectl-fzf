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

func (r *resourceWatcher) watchResource(ctx context.Context, k8sStore K8sStore, runtimeObject runtime.Object) {
	glog.V(4).Infof("Start watch for %s on namespace %s", k8sStore.resourceName,
		r.namespace)
	watchlist := cache.NewListWatchFromClient(r.clientset.Core().RESTClient(),
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

	serviceStore, err := NewK8sStore(func() K8sResource { return &Service{} },
		"services", cacheDir)
	fatalIf(err)
	podStore, err := NewK8sStore(func() K8sResource { return &Pod{} },
		"pods", cacheDir)
	fatalIf(err)
	go resourceWatcher.watchResource(ctx, podStore, &corev1.Pod{})
	go resourceWatcher.watchResource(ctx, serviceStore, &corev1.Service{})

	<-ctx.Done()
}
