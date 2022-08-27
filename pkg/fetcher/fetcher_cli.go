package fetcher

import (
	"kubectlfzf/pkg/k8s/clusterconfig"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type FetcherCli struct {
	clusterconfig.ClusterConfigCli
	HttpEndpoint         string
	FzfNamespace         string
	FetcherCachePath     string
	MinimumCache         time.Duration
	PortForwardLocalPort int
}

func SetFetchConfigFlags(fs *pflag.FlagSet) {
	clusterconfig.SetClusterConfigCli(fs)
	fs.String("http-endpoint", "", "Force completion to fetch data from a specific http endpoint.")
	fs.String("fetcher-cache-path", "/tmp/kubectl_fzf_cache/fetcher_cache", "Location of cached resources fetched from a remote kubectl-fzf instance.")
	fs.String("fzf-namespace", "", "The namespace to look for a kubectl-fzf pod.")
	fs.Int("port-forward-local-port", 8080, "The local port to use for port-forward.")
	fs.Duration("minimum-cache", time.Minute, "The minimum duration after which the http endpoint will be queried to check for resource modification.")
}

func GetFetchConfigCli() FetcherCli {
	return FetcherCli{
		ClusterConfigCli:     clusterconfig.GetClusterConfigCli(),
		FetcherCachePath:     viper.GetString("fetcher-cache-path"),
		HttpEndpoint:         viper.GetString("http-endpoint"),
		FzfNamespace:         viper.GetString("fzf-namespace"),
		MinimumCache:         viper.GetDuration("minimum-cache"),
		PortForwardLocalPort: viper.GetInt("port-forward-local-port"),
	}
}
