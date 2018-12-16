package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var (
	kubeconfig             string
	namespace              string
	cacheDir               string
	nodePollingPeriod      time.Duration
	namespacePollingPeriod time.Duration
)

func init() {
	if home := os.Getenv("HOME"); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.StringVar(&namespace, "namespace", "", "Namespace to watch, empty for all namespaces")
	flag.StringVar(&cacheDir, "dir", os.Getenv("KUBECTL_FZF_CACHE"), "Cache dir location. Default to KUBECTL_FZF_CACHE env var")
	flag.DurationVar(&nodePollingPeriod, "node-polling-period", 300*time.Second, "Polling period for nodes")
	flag.DurationVar(&namespacePollingPeriod, "namespace-polling-period", 300*time.Second, "Polling period for namespaces")
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

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	resourceWatcher := NewResourceWatcher(namespace)
	coreGetter := resourceWatcher.clientset.Core().RESTClient()
	appsGetter := resourceWatcher.clientset.Apps().RESTClient()

	storeConfigs := []watchConfig{
		watchConfig{NewPodFromRuntime, PodHeader, string(corev1.ResourcePods), coreGetter, &corev1.Pod{}, true, 0},
		watchConfig{NewServiceFromRuntime, ServiceHeader, string(corev1.ResourceServices), coreGetter, &corev1.Service{}, true, 0},
		watchConfig{NewReplicaSetFromRuntime, ReplicaSetHeader, "replicasets", appsGetter, &appsv1.ReplicaSet{}, true, 0},
		watchConfig{NewConfigMapFromRuntime, ConfigMapHeader, "configmaps", coreGetter, &corev1.ConfigMap{}, true, 0},
		watchConfig{NewStatefulSetFromRuntime, StatefulSetHeader, "statefulsets", appsGetter, &appsv1.StatefulSet{}, true, 0},
		watchConfig{NewDeploymentFromRuntime, DeploymentHeader, "deployments", appsGetter, &appsv1.Deployment{}, true, 0},
		watchConfig{NewEndpointsFromRuntime, EndpointsHeader, "endpoints", coreGetter, &corev1.Endpoints{}, true, 0},
		watchConfig{NewPersistentVolumeFromRuntime, PersistentVolumeHeader, "persistentvolumes", coreGetter, &corev1.PersistentVolume{}, false, 0},
		watchConfig{NewPersistentVolumeClaimFromRuntime, PersistentVolumeClaimHeader, string(corev1.ResourcePersistentVolumeClaims), coreGetter, &corev1.PersistentVolumeClaim{}, true, 0},
		watchConfig{NewNodeFromRuntime, NodeHeader, "nodes", coreGetter, &corev1.Node{}, false, nodePollingPeriod},
		watchConfig{NewNamespaceFromRuntime, NamespaceHeader, "namespaces", coreGetter, &corev1.Namespace{}, false, namespacePollingPeriod},
	}

	for _, watchConfig := range storeConfigs {
		store, err := NewK8sStore(watchConfig.resourceCtor, watchConfig.resourceName, watchConfig.header, cacheDir)
		util.FatalIf(err)
		if watchConfig.pollingPeriod > 0 {
			go resourceWatcher.pollResource(ctx, watchConfig, store)
		} else {
			go resourceWatcher.watchResource(ctx, watchConfig, store)
		}
	}

	<-ctx.Done()
}
