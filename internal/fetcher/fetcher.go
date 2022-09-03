package fetcher

import (
	"context"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
)

// Fetcher defines configuration to fetch completion datas
type Fetcher struct {
	clusterconfig.ClusterConfig
	fetcherCachePath     string
	httpEndpoint         string
	fzfNamespace         string
	minimumCache         time.Duration
	portForwardLocalPort int // Local port to use for port-forward
	fetcherState         FetcherState
}

func NewFetcher(fetchConfigCli *FetcherCli) *Fetcher {
	f := Fetcher{
		ClusterConfig:        clusterconfig.NewClusterConfig(fetchConfigCli.ClusterConfigCli),
		httpEndpoint:         fetchConfigCli.HttpEndpoint,
		fzfNamespace:         fetchConfigCli.FzfNamespace,
		fetcherCachePath:     fetchConfigCli.FetcherCachePath,
		minimumCache:         fetchConfigCli.MinimumCache,
		portForwardLocalPort: fetchConfigCli.PortForwardLocalPort,
		fetcherState:         *newFetcherState(fetchConfigCli.FetcherCachePath),
	}
	return &f
}

func (f *Fetcher) LoadFetcherState() error {
	err := f.LoadClusterConfig()
	if err != nil {
		return err
	}
	return f.fetcherState.loadStateFromDisk()
}

func (f *Fetcher) SaveFetcherState() error {
	return f.fetcherState.writeToDisk()
}

func loadResourceFromFile(filePath string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, filePath)
	return resources, err
}

func (f *Fetcher) GetResources(ctx context.Context, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	resources, err := f.checkLocalFiles(r)
	if resources != nil || err != nil {
		return resources, err
	}

	// Check for recent cache
	resources, err = f.checkRecentCache(r)
	if resources != nil || err != nil {
		return resources, err
	}

	// Fetch remote
	if util.IsAddressReachable(f.httpEndpoint) {
		return f.loadResourceFromHttpServer(f.httpEndpoint, r)
	}
	return f.getResourcesFromPortForward(ctx, r)
}
