package clusterconfig

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ClusterConfigCli struct {
	ClusterName string // Only for testing purpose
	CacheDir    string
}

func SetClusterConfigCli(fs *pflag.FlagSet) {
	fs.String("cache-dir", "/tmp/kubectl_fzf_cache/", "Cache dir location.")
}

func GetClusterConfigCli() *ClusterConfigCli {
	c := ClusterConfigCli{}
	c.CacheDir = viper.GetString("cache-dir")
	return &c
}
