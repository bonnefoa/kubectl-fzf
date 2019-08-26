package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"

	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/resourcewatcher"
	"kubectlfzf/pkg/util"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version                = "1.2"
	displayVersion         bool
	cpuProfile             bool
	inCluster              bool
	kubeconfig             string
	excludedNamespaces     []string
	cacheDir               string
	roleBlacklist          []string
	roleBlacklistSet       map[string]bool
	timeBetweenFullDump    time.Duration
	nodePollingPeriod      time.Duration
	namespacePollingPeriod time.Duration

	daemonCmd         string
	daemonName        string
	daemonPidFilePath string
	daemonLogFilePath string
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
	flag.Bool("in-cluster", false, "Use in-cluster configuration")
	flag.String("excluded-namespaces", "", "Namespaces to exclude, separated by comma")
	flag.String("cache-dir", defaultCacheDirEnv, "Cache dir location. Default to KUBECTL_FZF_CACHE env var")
	flag.String("role-blacklist", "", "List of roles to hide from node list, separated by commas")
	flag.Duration("time-between-fulldump", 60*time.Second, "Buffer changes and only do full dump every x secondes")
	flag.Duration("node-polling-period", 300*time.Second, "Polling period for nodes")
	flag.Duration("namespace-polling-period", 600*time.Second, "Polling period for namespaces")

	flag.String("daemon", "", `Send signal to the daemon:
  start - run as a daemon
  stop — fast shutdown`)
	defaultName := path.Base(os.Args[0])
	flag.String("daemon-name", defaultName, "Daemon name")
	defaultPidPath := path.Join("/tmp/", defaultName+".pid")
	defaultLogPath := path.Join("/tmp/", defaultName+".log")
	flag.String("daemon-pid-file", defaultPidPath, "Daemon's PID file path")
	flag.String("daemon-log-file", defaultLogPath, "Daemon's log file path")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName(".kubectl_fzf")
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		util.FatalIf(err)
	}

	displayVersion = viper.GetBool("version")
	cpuProfile = viper.GetBool("cpu-profile")
	inCluster = viper.GetBool("in-cluster")
	kubeconfig = viper.GetString("kubeconfig")
	cacheDir = viper.GetString("cache-dir")
	roleBlacklist = viper.GetStringSlice("role-blacklist")
	excludedNamespaces = viper.GetStringSlice("excluded-namespaces")
	timeBetweenFullDump = viper.GetDuration("time-between-fulldump")
	nodePollingPeriod = viper.GetDuration("node-polling-period")
	namespacePollingPeriod = viper.GetDuration("namespace-polling-period")

	daemonCmd = viper.GetString("daemon")
	daemonName = viper.GetString("daemon-name")
	daemonPidFilePath = viper.GetString("daemon-pid-file")
	daemonLogFilePath = viper.GetString("daemon-log-file")
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

func termHandler(sig os.Signal) error {
	glog.Infoln("Terminating daemon...")
	return daemon.ErrStop
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

func getClientConfigAndCluster() (*rest.Config, string) {
	if inCluster {
		restConfig, err := rest.InClusterConfig()
		util.FatalIf(err)
		return restConfig, "incluster"
	}

	configInBytes, err := ioutil.ReadFile(kubeconfig)
	util.FatalIf(err)
	clientConfig, err := clientcmd.NewClientConfigFromBytes(configInBytes)
	util.FatalIf(err)

	rawConfig, err := clientConfig.RawConfig()
	util.FatalIf(err)
	cluster := rawConfig.CurrentContext

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	util.FatalIf(err)
	return cfg, cluster
}

func processArgs() {
	glog.Infof("Building role blacklist from \"%s\"", roleBlacklist)
	roleBlacklistSet = util.StringSliceToSet(roleBlacklist)
}

func start() {
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

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	processArgs()

	if daemonCmd == "" && !daemon.WasReborn() {
		start()
		return
	}

	daemon.AddCommand(daemon.StringFlag(&daemonCmd, "stop"), syscall.SIGTERM, termHandler)

	cntxt := &daemon.Context{
		PidFileName: daemonPidFilePath,
		PidFilePerm: 0644,
		LogFileName: daemonLogFilePath,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{daemonName},
	}

	if len(daemon.ActiveFlags()) > 0 {
		glog.Infof("Stopping daemon...")
		d, err := cntxt.Search()
		if err != nil {
			glog.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}
	glog.Infof("Starting daemon...")

	d, err := cntxt.Reborn()
	if err != nil {
		glog.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	glog.Infoln("- - - - - - - - - - - - - - -")
	glog.Infoln("daemon started")

	go start()

	err = daemon.ServeSignals()
	if err != nil {
		glog.Infof("Error: %s", err.Error())
	}

	glog.Infoln("daemon terminated")

}
