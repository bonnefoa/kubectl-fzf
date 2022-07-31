package util

import (
	"flag"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterCliConf struct {
	ClusterName string
	InCluster   bool
	CacheDir    string
	Kubeconfig  string
	IsExplicit  bool
}

func SetClusterConfFlags() {
	flag.String("kubeconfig", clientcmd.RecommendedHomeFile, "(optional) absolute path to the kubeconfig file, KUBECONFIG variable has higher priority)")
	flag.Bool("explicit-config", false, "Configure config loader to use file specified by 'kubeconfig' flag explicitly.")
	flag.String("cache-dir", "/tmp/kubectl_fzf_cache/", "Cache dir location.")
	flag.String("cluster-name", "incluster", "The cluster name. Needed for cross-cluster completion.")
	flag.Bool("in-cluster", false, "Use in-cluster configuration")
}

func GetClusterCliConf() ClusterCliConf {
	c := ClusterCliConf{}
	c.InCluster = viper.GetBool("in-cluster")
	c.ClusterName = viper.GetString("cluster-name")
	c.Kubeconfig = viper.GetString("kubeconfig")
	c.IsExplicit = viper.GetBool("explicit-config")
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

	configPath := &clientcmd.PathOptions{
		GlobalFile:   c.Kubeconfig,
		EnvVar:       clientcmd.RecommendedConfigPathEnvVar,
		LoadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
	}

	if c.IsExplicit {
		configPath.LoadingRules.ExplicitPath = c.Kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configPath.LoadingRules, &clientcmd.ConfigOverrides{})
	rawConfig, err := clientConfig.RawConfig()
	c.ClusterName = rawConfig.CurrentContext
	FatalIf(err)

	cfg, err := clientcmd.BuildConfigFromKubeconfigGetter("", configPath.GetStartingConfig)
	FatalIf(err)

	//fmt.Printf("%+v\n", cfg)
	//os.Exit(1)

	return cfg, c.ClusterName
}
