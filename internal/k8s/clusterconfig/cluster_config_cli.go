package clusterconfig

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ClusterConfigCli struct {
	ClusterName string
	CacheDir    string
	Kubeconfig  string
}

func SetClusterConfigCli(fs *pflag.FlagSet) {
	if home := os.Getenv("HOME"); home != "" {
		fs.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		fs.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	fs.String("cache-dir", "/tmp/kubectl_fzf_cache/", "Cache dir location.")
	fs.String("cluster-name", "", "The cluster name. Needed for cross-cluster completion.")
}

func GetClusterConfigCli() *ClusterConfigCli {
	c := ClusterConfigCli{}
	c.ClusterName = viper.GetString("cluster-name")
	c.Kubeconfig = viper.GetString("kubeconfig")
	c.CacheDir = viper.GetString("cache-dir")
	return &c
}
