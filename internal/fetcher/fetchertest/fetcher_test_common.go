package fetchertest

import (
	"fmt"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
)

func GetTestFetcher(t *testing.T, clusterName string, port int) (*fetcher.Fetcher, string) {
	tempDir := t.TempDir()
	fetchCli := &fetcher.FetcherCli{
		FetcherCachePath: tempDir,
		ClusterConfigCli: &clusterconfig.ClusterConfigCli{
			ClusterName: clusterName,
			CacheDir:    "testdata",
		},
		HttpEndpoint: fmt.Sprintf("localhost:%d", port),
	}
	f := fetcher.NewFetcher(fetchCli)
	return f, tempDir
}

func GetTestFetcherWithDefaults(t *testing.T) *fetcher.Fetcher {
	f, _ := GetTestFetcher(t, "minikube", 8080)
	return f
}
