package clusterconfig

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ClusterConfigCli struct {
	ClusterName string
	CacheDir    string
}

func SetClusterConfigCli(fs *pflag.FlagSet) {
	fs.String("cache-dir", "/tmp/kubectl_fzf_cache/", "Cache dir location.")
	fs.String("cluster-name", "", "The cluster name. Needed for cross-cluster completion.")
}

func GetClusterConfigCli() *ClusterConfigCli {
	c := ClusterConfigCli{}
	c.ClusterName = viper.GetString("cluster-name")
	c.CacheDir = viper.GetString("cache-dir")
	return &c
}
