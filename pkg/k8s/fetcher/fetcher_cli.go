package fetcher

import (
	"kubectlfzf/pkg/k8s/clusterconfig"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type FetcherCli struct {
	clusterconfig.ClusterConfigCli
	HttpEndpoint         string
	FetcherCachePath     string
	PortForwardLocalPort int
}

func SetFetchConfigFlags(fs *pflag.FlagSet) {
	clusterconfig.SetClusterConfigCli(fs)
	fs.String("http-endpoint", "", "Force completion to fetch data from a specific http endpoint.")
	fs.String("fetcher-cache-path", "/tmp/kubectl_fzf/fetcher_cache", "Location of cached resources fetched from a remote kubectl-fzf instance.")
	fs.Int("port-forward-local-port", 8080, "The local port to use for port-forward.")
}

func GetFetchConfigCli() FetcherCli {
	f := FetcherCli{
		ClusterConfigCli: clusterconfig.GetClusterConfigCli(),
	}
	f.HttpEndpoint = viper.GetString("http-endpoint")
	f.PortForwardLocalPort = viper.GetInt("port-forward-local-port")
	return f
}
