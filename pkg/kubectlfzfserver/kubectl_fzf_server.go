package kubectlfzfserver

import (
	"context"
	"kubectlfzf/pkg/httpserver"
	"kubectlfzf/pkg/k8s/resourcewatcher"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/util"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func startWatchOnCluster(ctx context.Context, resourceWatcherCli resourcewatcher.ResourceWatcherCli,
	storeConfig *store.StoreConfig) (*resourcewatcher.ResourceWatcher, error) {
	cluster := storeConfig.GetContext()
	watcher, err := resourcewatcher.NewResourceWatcher(cluster, resourceWatcherCli, storeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error creating resource watcher")
	}
	err = watcher.FetchNamespaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching namespaces")
	}
	watchConfigs, err := watcher.GetWatchConfigs()
	if err != nil {
		return nil, errors.Wrap(err, "error getting watchdog configs")
	}

	logrus.Infof("Start cache build on cluster %s", cluster)
	for _, watchConfig := range watchConfigs {
		err := watcher.Start(ctx, watchConfig)
		if err != nil {
			return nil, errors.Wrap(err, "error starting watcher")
		}
	}
	err = watcher.DumpAPIResources()
	if err != nil {
		return nil, errors.Wrap(err, "error when dumping api resources")
	}
	return watcher, nil
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

func StartKubectlFzfServer() {
	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	storeConfigCli := store.GetStoreConfigCli()
	storeConfig := store.NewStoreConfig(&storeConfigCli)
	err := storeConfig.SetClusterNameFromCurrentContext()
	if err != nil {
		logrus.Fatal("Couldn't get current context: ", err)
	}
	err = storeConfig.CreateDestDir()
	if err != nil {
		logrus.Fatalf("error creating destination dir: %s", err)
	}

	resourceWatcherCli := resourcewatcher.GetResourceWatcherCli()
	watcher, err := startWatchOnCluster(ctx, resourceWatcherCli, storeConfig)
	util.FatalIf(err)
	ticker := time.NewTicker(time.Second * 5)

	currentRestConfig, err := storeConfig.GetClientConfig()
	if err != nil {
		logrus.Fatalf("Error getting client config: %s", err)
	}
	httpServerConfCli := httpserver.GetHttpServerConfigCli()
	_, err = httpserver.StartHttpServer(ctx, &httpServerConfCli, storeConfig)
	if err != nil {
		logrus.Fatalf("Error starting http server: %s", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			restConfig, err := storeConfig.GetClientConfig()
			util.FatalIf(err)
			logrus.Tracef("Checking config %s %s ", restConfig.Host, currentRestConfig.Host)
			if restConfig.Host != currentRestConfig.Host {
				logrus.Infof("Detected cluster change %s != %s", restConfig.Host, currentRestConfig.Host)
				watcher.Stop()
				storeConfig.SetClusterNameFromCurrentContext()
				watcher, err = startWatchOnCluster(ctx, resourceWatcherCli, storeConfig)
				util.FatalIf(err)
				currentRestConfig = restConfig
			}
		}
	}
}
