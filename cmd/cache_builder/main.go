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

	watcher := resourcewatcher.NewResourceWatcher(namespace, kubeconfig)
	watchConfigs := watcher.GetWatchConfigs(nodePollingPeriod, namespacePollingPeriod)

	for _, watchConfig := range watchConfigs {
		err := watcher.Start(ctx, watchConfig, cacheDir, timeBetweenFullDump)
		util.FatalIf(err)
	}

	<-ctx.Done()
}
