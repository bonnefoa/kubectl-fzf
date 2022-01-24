package util

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
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

	defaultCacheDirEnv, assigned := os.LookupEnv("KUBECTL_FZF_CACHE")
	if assigned == false {
		defaultCacheDirEnv = "/tmp/kubectl_fzf_cache/"
	}
	flag.String("cache-dir", defaultCacheDirEnv, "Cache dir location. Default to KUBECTL_FZF_CACHE env var")
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

func ParseFlags() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName(".kubectl_fzf")
	viper.AddConfigPath("/etc/kubectl_fzf/")
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		FatalIf(err)
	}
}
