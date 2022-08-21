package httpservertest

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/httpserver"
	"kubectlfzf/pkg/k8s/clusterconfig"
	"kubectlfzf/pkg/k8s/fetcher"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/k8s/store/storetest"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetTestClusterConfigCli() clusterconfig.ClusterConfigCli {
	return clusterconfig.ClusterConfigCli{ClusterName: "minikube", CacheDir: "./testdata", Kubeconfig: ""}
}

func GetTestStoreConfigCli() *store.StoreConfigCli {
	return &store.StoreConfigCli{ClusterConfigCli: GetTestClusterConfigCli()}
}

func GetTestFetchConfig(t *testing.T) *fetcher.Fetcher {
	f := fetcher.FetcherCli{ClusterConfigCli: GetTestClusterConfigCli(), HttpEndpoint: "localhost:0"}
	fetchConfig := fetcher.NewFetcher(&f)
	return fetchConfig
}

func StartTestHttpServer(t *testing.T) *fetcher.Fetcher {
	ctx := context.Background()
	storeConfigCli := GetTestStoreConfigCli()
	storeConfig := store.NewStoreConfig(storeConfigCli)
	_, podStore := storetest.GetTestPodStore(t)
	h := &httpserver.HttpServerConfigCli{ListenAddress: "localhost:0", Debug: false}
	port, err := httpserver.StartHttpServer(ctx, h, storeConfig, []*store.Store{podStore})
	require.NoError(t, err)
	fetchConfigCli := &fetcher.FetcherCli{
		FetcherCachePath: t.TempDir(),
		ClusterConfigCli: clusterconfig.ClusterConfigCli{
			ClusterName: "testcluster",
			CacheDir:    "doesntexist",
		},
		HttpEndpoint: fmt.Sprintf("localhost:%d", port),
	}
	f := fetcher.NewFetcher(fetchConfigCli)
	require.NoError(t, err)
	return f
}
