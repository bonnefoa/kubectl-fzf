package httpservertest

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/httpserver"
	"kubectlfzf/pkg/k8s/clusterconfig"
	"kubectlfzf/pkg/k8s/fetcher"
	"kubectlfzf/pkg/k8s/store"
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
	h := &httpserver.HttpServerConfigCli{ListenAddress: "localhost:0", Debug: false}
	port, err := httpserver.StartHttpServer(ctx, h, storeConfig, nil)
	require.NoError(t, err)
	fetchConfigCli := &fetcher.FetcherCli{
		ClusterConfigCli: clusterconfig.ClusterConfigCli{
			CacheDir: "doenstexist",
		},
		HttpEndpoint: fmt.Sprintf("localhost:%d", port),
	}
	f := fetcher.NewFetcher(fetchConfigCli)
	require.NoError(t, err)
	return f
}
