package kubectlfzfserver

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/httpserver"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resourcewatcher"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func startWatchOnCluster(ctx context.Context,
	resourceWatcherCli resourcewatcher.ResourceWatcherCli,
	storeConfig *store.StoreConfig) (*resourcewatcher.ResourceWatcher, []*store.Store, error) {
	cluster := storeConfig.GetContext()
	watcher, err := resourcewatcher.NewResourceWatcher(cluster, resourceWatcherCli, storeConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating resource watcher")
	}
	err = watcher.FetchNamespaces(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error fetching namespaces")
	}
	watchConfigs, err := watcher.GetWatchConfigs()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting watchdog configs")
	}
	logrus.Infof("Start cache build on cluster %s", cluster)
	stores := make([]*store.Store, 0)
	for _, watchConfig := range watchConfigs {
		store := watcher.Start(ctx, watchConfig)
		stores = append(stores, store)
	}
	err = watcher.DumpAPIResources()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error when dumping api resources")
	}
	return watcher, stores, nil
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
	watcher, stores, err := startWatchOnCluster(ctx, resourceWatcherCli, storeConfig)
	util.FatalIf(err)
	ticker := time.NewTicker(time.Second * 5)

	currentRestConfig, err := storeConfig.GetClientConfig()
	if err != nil {
		logrus.Fatalf("Error getting client config: %s", err)
	}
	httpServerConfCli := httpserver.GetHttpServerConfigCli()
	_, err = httpserver.StartHttpServer(ctx, &httpServerConfCli, storeConfig, stores)
	if err != nil {
		logrus.Fatalf("Error starting http server: %s", err)
	}

	go func() {
		logrus.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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
				// TODO: Handle stores change
				watcher, _, err = startWatchOnCluster(ctx, resourceWatcherCli, storeConfig)
				util.FatalIf(err)
				currentRestConfig = restConfig
			}
		}
	}
}
