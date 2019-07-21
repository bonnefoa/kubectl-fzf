package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/bonnefoa/kubectl-fzf/pkg/k8sresources"
	"github.com/bonnefoa/kubectl-fzf/pkg/resourcewatcher"
	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version                = "1.2"
	displayVersion         bool
	cpuProfile             bool
	kubeconfig             string
	excludedNamespaces     []string
	cacheDir               string
	roleBlacklist          []string
	roleBlacklistSet       map[string]bool
	timeBetweenFullDump    time.Duration
	nodePollingPeriod      time.Duration
	namespacePollingPeriod time.Duration
)

func init() {
	if home := os.Getenv("HOME"); home != "" {
		flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	defaultCacheDirEnv, assigned := os.LookupEnv("KUBECTL_FZF_CACHE")
	if assigned == false {
		defaultCacheDirEnv = "/tmp/kubectl_fzf_cache/"
	}

	flag.Bool("version", false, "Display version and exit")
	flag.Bool("cpu-profile", false, "Start with cpu profiling")
	flag.String("excluded-namespaces", "", "Namespaces to exclude, separated by comma")
	flag.String("cache-dir", defaultCacheDirEnv, "Cache dir location. Default to KUBECTL_FZF_CACHE env var")
	flag.String("role-blacklist", "", "List of roles to hide from node list, separated by commas")
	flag.Duration("time-between-fulldump", 60*time.Second, "Buffer changes and only do full dump every x secondes")
	flag.Duration("node-polling-period", 300*time.Second, "Polling period for nodes")
	flag.Duration("namespace-polling-period", 600*time.Second, "Polling period for namespaces")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName(".kubectl_fzf")
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		util.FatalIf(err)
	}

	displayVersion = viper.GetBool("version")
	cpuProfile = viper.GetBool("cpu-profile")
	kubeconfig = viper.GetString("kubeconfig")
	cacheDir = viper.GetString("cache-dir")
	roleBlacklist = viper.GetStringSlice("role-blacklist")
	excludedNamespaces = viper.GetStringSlice("excluded-namespaces")
	timeBetweenFullDump = viper.GetDuration("time-between-fulldump")
	nodePollingPeriod = viper.GetDuration("node-polling-period")
	namespacePollingPeriod = viper.GetDuration("namespace-polling-period")
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

func startWatchOnCluster(ctx context.Context, config *restclient.Config, cluster string) resourcewatcher.ResourceWatcher {
	storeConfig := resourcewatcher.StoreConfig{
		CacheDir:            cacheDir,
		Cluster:             cluster,
		TimeBetweenFullDump: timeBetweenFullDump,
	}
	watcher := resourcewatcher.NewResourceWatcher(config, storeConfig, excludedNamespaces)
	watcher.FetchNamespaces()
	watchConfigs := watcher.GetWatchConfigs(nodePollingPeriod, namespacePollingPeriod)
	ctorConfig := k8sresources.CtorConfig{
		RoleBlacklist: roleBlacklistSet,
	}

	glog.Infof("Start cache build on cluster %s", config.Host)
	for _, watchConfig := range watchConfigs {
		err := watcher.Start(ctx, watchConfig, ctorConfig)
		util.FatalIf(err)
	}
	err := watcher.DumpAPIResources()
	util.FatalIf(err)
	return watcher
}

func getClientConfigAndCluster() (*restclient.Config, string) {
	configInBytes, err := ioutil.ReadFile(kubeconfig)
	util.FatalIf(err)

	clientConfig, err := clientcmd.NewClientConfigFromBytes(configInBytes)
	util.FatalIf(err)

	rawConfig, err := clientConfig.RawConfig()
	util.FatalIf(err)
	cluster := rawConfig.CurrentContext

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	util.FatalIf(err)
	return restConfig, cluster
}

func processArgs() {
	glog.Infof("Building role blacklist from \"%s\"", roleBlacklist)
	roleBlacklistSet = util.StringSliceToSet(roleBlacklist)
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	processArgs()

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

	currentRestConfig, currentCluster := getClientConfigAndCluster()
	watcher := startWatchOnCluster(ctx, currentRestConfig, currentCluster)
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			restConfig, cluster := getClientConfigAndCluster()
			glog.V(7).Infof("Checking config %s %s ", restConfig.Host, currentRestConfig.Host)
			if restConfig.Host != currentRestConfig.Host {
				glog.Infof("Detected cluster change %s != %s", restConfig.Host, currentRestConfig.Host)
				watcher.Stop()
				watcher = startWatchOnCluster(ctx, restConfig, cluster)
				currentRestConfig = restConfig
				currentCluster = cluster
			}
		}
	}
}
