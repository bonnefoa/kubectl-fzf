package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	"github.com/sirupsen/logrus"

	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/resourcewatcher"
	"kubectlfzf/pkg/util"

	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	Version   string
	BuildDate string
	GitCommit string
	GitBranch string
	GoVersion string

	displayVersion         bool
	cpuProfile             bool
	excludedResources      []string
	excludedNamespaces     []string
	roleBlacklist          []string
	roleBlacklistSet       map[string]bool
	nodePollingPeriod      time.Duration
	namespacePollingPeriod time.Duration
	timeBetweenFullDump    time.Duration

	daemonCmd         string
	daemonName        string
	daemonPidFilePath string
	daemonLogFilePath string
)

func init() {
	util.SetClusterConfFlags()
	util.SetLogConfFlags()

	flag.Bool("version", false, "Display version and exit")
	flag.Bool("cpu-profile", false, "Start with cpu profiling")
	flag.String("excluded-namespaces", "", "Namespaces to exclude, separated by space")
	flag.String("excluded-resources", "", "Resources to exclude, separated by space. To exclude everything: pods configmaps services serviceaccounts replicasets daemonsets secrets statefulsets deployments endpoints ingresses cronjobs jobs horizontalpodautoscalers persistentvolumes persistentvolumeclaims nodes namespaces")
	flag.String("role-blacklist", "", "List of roles to hide from node list, separated by commas")
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
	flag.Duration("time-between-fulldump", 60*time.Second, "Buffer changes and only do full dump every x secondes")

	util.ParseFlags()

	displayVersion = viper.GetBool("version")
	cpuProfile = viper.GetBool("cpu-profile")
	roleBlacklist = viper.GetStringSlice("role-blacklist")
	excludedNamespaces = viper.GetStringSlice("excluded-namespaces")
	excludedResources = viper.GetStringSlice("excluded-resources")
	nodePollingPeriod = viper.GetDuration("node-polling-period")
	namespacePollingPeriod = viper.GetDuration("namespace-polling-period")
	timeBetweenFullDump = viper.GetDuration("time-between-fulldump")

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
			logrus.Errorf("Caught signal '%s' (%d); terminating.", sig, sig)
			cancel()
		}
	}
}

func termHandler(sig os.Signal) error {
	logrus.Infoln("Terminating daemon...")
	return daemon.ErrStop
}

func startWatchOnCluster(ctx context.Context, clusterCliConf *util.ClusterCliConf) resourcewatcher.ResourceWatcher {
	config, cluster := clusterCliConf.GetClientConfigAndCluster()
	storeConfig := k8sresources.NewStoreConfig(clusterCliConf, timeBetweenFullDump)
	watcher := resourcewatcher.NewResourceWatcher(config, storeConfig, excludedNamespaces)
	watcher.FetchNamespaces(ctx)
	watchConfigs := watcher.GetWatchConfigs(nodePollingPeriod, namespacePollingPeriod, excludedResources)
	ctorConfig := k8sresources.CtorConfig{
		RoleBlacklist: roleBlacklistSet,
		Cluster:       cluster,
	}

	logrus.Infof("Start cache build on cluster %s", cluster)
	for _, watchConfig := range watchConfigs {
		err := watcher.Start(ctx, watchConfig, ctorConfig)
		util.FatalIf(err)
	}
	err := watcher.DumpAPIResources()
	util.FatalIf(err)
	return watcher
}

func processArgs() {
	logrus.Infof("Building role blacklist from \"%s\"", roleBlacklist)
	roleBlacklistSet = util.StringSliceToSet(roleBlacklist)
}

func handleDisplayVersion() {
	if displayVersion {
		fmt.Printf("Version: %s\n", Version)
		if GitCommit != "" {
			fmt.Printf("Git hash: %s\n", GitCommit)
		}
		if GitBranch != "" {
			fmt.Printf("Git branch: %s\n", GitBranch)
		}
		if BuildDate != "" {
			fmt.Printf("Build date: %s\n", BuildDate)
		}
		if GoVersion != "" {
			fmt.Printf("Go Version: %s\n", GoVersion)
		}
		os.Exit(0)
	}
}

func handleCpuProfile() {
	if cpuProfile {
		f, err := os.Create("cpu.pprof")
		util.FatalIf(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
}

func start() {
	util.ConfigureLog()
	handleDisplayVersion()

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	clusterCliConf := util.GetClusterCliConf()
	watcher := startWatchOnCluster(ctx, &clusterCliConf)
	ticker := time.NewTicker(time.Second * 5)

	currentRestConfig, _ := clusterCliConf.GetClientConfigAndCluster()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			restConfig, _ := clusterCliConf.GetClientConfigAndCluster()
			logrus.Debugf("Checking config %s %s ", restConfig.Host, currentRestConfig.Host)
			if restConfig.Host != currentRestConfig.Host {
				logrus.Infof("Detected cluster change %s != %s", restConfig.Host, currentRestConfig.Host)
				watcher.Stop()
				watcher = startWatchOnCluster(ctx, &clusterCliConf)
				currentRestConfig = restConfig
			}
		}
	}
}

func startDaemon() {
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
		logrus.Infof("Stopping daemon...")
		d, err := cntxt.Search()
		if err != nil {
			logrus.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}
	logrus.Infof("Starting daemon...")

	d, err := cntxt.Reborn()
	if err != nil {
		logrus.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	logrus.Infoln("- - - - - - - - - - - - - - -")
	logrus.Infoln("daemon started")

	go start()

	err = daemon.ServeSignals()
	if err != nil {
		logrus.Infof("Error: %s", err.Error())
	}
	logrus.Infoln("daemon terminated")
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	processArgs()

	if daemonCmd == "" && !daemon.WasReborn() {
		start()
		return
	}
	startDaemon()
}
