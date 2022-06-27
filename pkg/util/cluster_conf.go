package util

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterCliConf struct {
	ClusterName string
	InCluster   bool
	CacheDir    string
	Kubeconfig  string
}

func SetClusterConfFlags() {
	if home := os.Getenv("HOME"); home != "" {
		flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.String("cache-dir", "/tmp/kubectl_fzf_cache/", "Cache dir location.")
	flag.String("cluster-name", "incluster", "The cluster name. Needed for cross-cluster completion.")
	flag.Bool("in-cluster", false, "Use in-cluster configuration")
}

func GetClusterCliConf() ClusterCliConf {
	c := ClusterCliConf{}
	c.InCluster = viper.GetBool("in-cluster")
	c.ClusterName = viper.GetString("cluster-name")
	c.Kubeconfig = viper.GetString("kubeconfig")
	c.CacheDir = viper.GetString("cache-dir")
	return c
}

func (c *ClusterCliConf) GetClusterName() string {
	clusterDir := c.ClusterName
	if c.InCluster {
		clusterDir = "incluster"
	}
	return clusterDir
}

func (c *ClusterCliConf) GetClientConfigAndCluster() (*rest.Config, string) {
	if c.InCluster {
		restConfig, err := rest.InClusterConfig()
		FatalIf(err)
		return restConfig, c.ClusterName
	}

	configInBytes, err := ioutil.ReadFile(c.Kubeconfig)
	FatalIf(err)
	clientConfig, err := clientcmd.NewClientConfigFromBytes(configInBytes)
	FatalIf(err)

	rawConfig, err := clientConfig.RawConfig()
	FatalIf(err)
	cluster := rawConfig.CurrentContext

	cfg, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	FatalIf(err)
	return cfg, cluster
}
