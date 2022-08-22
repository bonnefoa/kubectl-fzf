package fetchertest

import (
	"fmt"
	"kubectlfzf/pkg/fetcher"
	"kubectlfzf/pkg/k8s/clusterconfig"
	"testing"
)

func GetTestFetcher(t *testing.T, port int) *fetcher.Fetcher {
	fetchCli := &fetcher.FetcherCli{
		FetcherCachePath: t.TempDir(),
		ClusterConfigCli: clusterconfig.ClusterConfigCli{
			ClusterName: "minikube",
			CacheDir:    "testdata",
		},
		HttpEndpoint: fmt.Sprintf("localhost:%d", port),
	}
	f := fetcher.NewFetcher(fetchCli)
	return f
}

func GetTestFetcherWithDefaultPort(t *testing.T) *fetcher.Fetcher {
	return GetTestFetcher(t, 8080)
}
