package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/resourcewatcher"
	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version                = "1.0"
	displayVersion         bool
	cpuProfile             bool
	kubeconfig             string
	namespace              string
	cacheDir               string
	timeBetweenFullDump    time.Duration
	nodePollingPeriod      time.Duration
	namespacePollingPeriod time.Duration
)

func init() {
	if home := os.Getenv("HOME"); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.BoolVar(&displayVersion, "version", false, "Display version and exit")
	flag.BoolVar(&cpuProfile, "cpu-profile", false, "Start with cpu profiling")
	flag.StringVar(&namespace, "namespace", "", "Namespace to watch, empty for all namespaces")
	flag.StringVar(&cacheDir, "dir", os.Getenv("KUBECTL_FZF_CACHE"), "Cache dir location. Default to KUBECTL_FZF_CACHE env var")
	flag.DurationVar(&timeBetweenFullDump, "time-between-fulldump", 60*time.Second, "Buffer changes and only do full dump every x secondes")
	flag.DurationVar(&nodePollingPeriod, "node-polling-period", 300*time.Second, "Polling period for nodes")
	flag.DurationVar(&namespacePollingPeriod, "namespace-polling-period", 600*time.Second, "Polling period for namespaces")
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

func startWatchOnCluster(ctx context.Context, config *restclient.Config) resourcewatcher.ResourceWatcher {
	watcher := resourcewatcher.NewResourceWatcher(namespace, config)
	watchConfigs := watcher.GetWatchConfigs(nodePollingPeriod, namespacePollingPeriod)
	cluster, err := util.ExtractClusterFromHost(config.Host)
	util.FatalIf(err)
	storeConfig := resourcewatcher.StoreConfig{
		CacheDir:            cacheDir,
		Cluster:             cluster,
		TimeBetweenFullDump: timeBetweenFullDump,
	}

	glog.Infof("Start cache build on cluster %s", config.Host)
	for _, watchConfig := range watchConfigs {
		err := watcher.Start(ctx, watchConfig, storeConfig)
		util.FatalIf(err)
	}
	return watcher
}

func main() {
	flag.Parse()

	if displayVersion {
		fmt.Printf("%s", version)
		os.Exit(0)
	}

	if cpuProfile {
		f, err := os.Create("cpu.pprof")
		util.FatalIf(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	currentConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	util.FatalIf(err)

	watcher := startWatchOnCluster(ctx, currentConfig)
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			glog.V(7).Infof("Checking config %s %s ", config.Host, currentConfig.Host)
			util.FatalIf(err)
			if config.Host != currentConfig.Host {
				glog.Infof("Detected cluster change %s != %s", config.Host, currentConfig.Host)
				watcher.Stop()
				watcher = startWatchOnCluster(ctx, config)
				currentConfig = config
			}
		}
	}
}
