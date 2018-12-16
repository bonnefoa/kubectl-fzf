package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/resourcewatcher"
	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
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

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	watcher := resourcewatcher.NewResourceWatcher(namespace, kubeconfig)
	watchConfigs := watcher.GetWatchConfigs(nodePollingPeriod, namespacePollingPeriod)

	for _, watchConfig := range watchConfigs {
		err := watcher.Start(ctx, watchConfig, cacheDir)
		util.FatalIf(err)
	}

	<-ctx.Done()
}
