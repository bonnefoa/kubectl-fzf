package clusterconfig

import (
	"fmt"
	"os"
	"path"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClusterConfig struct {
	clusterName string
	destDir     string
	cacheDir    string

	apiConfig *clientcmdapi.Config
}

func NewClusterConfig(clusterConfigCli *ClusterConfigCli) ClusterConfig {
	c := ClusterConfig{}
	c.clusterName = clusterConfigCli.ClusterName
	c.cacheDir = clusterConfigCli.CacheDir
	c.destDir = path.Join(c.cacheDir, c.clusterName)
	return c
}

func (c *ClusterConfig) LoadClusterConfig() (err error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	c.apiConfig, err = loadingRules.Load()
	if err != nil {
		return errors.Wrap(err, "error reading kubeconfig file")
	}
	c.clusterName = c.apiConfig.CurrentContext
	if c.clusterName == "" {
		logrus.Infof("Couldn't read kubeconfig file, assuming incluster")
		c.clusterName = "incluster"
	}
	c.destDir = path.Join(c.cacheDir, c.clusterName)
	logrus.Infof("Cluster config set to target '%s'", c.destDir)
	return nil
}

func (c *ClusterConfig) CreateDestDir() error {
	if c.clusterName == "" {
		return errors.New("clustername is empty, call LoadClusterConfig before")
	}
	logrus.Infof("Creating destination dir '%s'", c.destDir)
	err := os.MkdirAll(c.destDir, os.ModePerm)
	return err
}

func (c *ClusterConfig) GetResourceStorePath(r resources.ResourceType) string {
	return path.Join(c.destDir, r.String())
}

func (c *ClusterConfig) FileStoreExists(r resources.ResourceType) bool {
	p := c.GetResourceStorePath(r)
	return util.FileExists(p)
}

func (c *ClusterConfig) GetClientset() (*kubernetes.Clientset, error) {
	restConfig, err := c.GetClientConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	return clientset, err
}

func (c *ClusterConfig) GetNamespace() (string, error) {
	contextStruct, ok := c.apiConfig.Contexts[c.apiConfig.CurrentContext]
	if !ok {
		return "", fmt.Errorf("context %s not found in config", c.apiConfig.CurrentContext)
	}
	return contextStruct.Namespace, nil
}

func (c *ClusterConfig) GetContext() string {
	return c.clusterName
}

func (c *ClusterConfig) GetClientConfig() (*rest.Config, error) {
	restConfig, err := rest.InClusterConfig()
	if err == nil {
		return restConfig, nil
	}
	cmdConfig := clientcmd.NewDefaultClientConfig(*c.apiConfig, nil)
	restConfig, err = cmdConfig.ClientConfig()
	return restConfig, err
}
