package clusterconfig

import (
	"fmt"
	"io/ioutil"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/util"
	"os"
	"path"

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
	kubeconfig  string
}

func NewClusterConfig(clusterConfigCli *ClusterConfigCli) ClusterConfig {
	c := ClusterConfig{}
	c.kubeconfig = clusterConfigCli.Kubeconfig
	c.clusterName = clusterConfigCli.ClusterName
	c.cacheDir = clusterConfigCli.CacheDir
	c.destDir = path.Join(c.cacheDir, c.clusterName)
	logrus.Infof("Cluster config set to target '%s'", c.destDir)
	return c
}

func (c *ClusterConfig) SetClusterNameFromCurrentContext() error {
	if util.FileExists(c.kubeconfig) {
		rawConfig, err := c.getRawConfig()
		if err != nil {
			return err
		}
		c.clusterName = rawConfig.CurrentContext
		c.destDir = path.Join(c.cacheDir, c.clusterName)
		return nil
	}
	logrus.Infof("kubeconfig file %s doesn't exists, assuming incluster", c.kubeconfig)
	c.clusterName = "incluster"
	c.destDir = path.Join(c.cacheDir, c.clusterName)
	return nil
}

func (c *ClusterConfig) CreateDestDir() error {
	if c.clusterName == "" {
		return errors.New("clustername is empty, call SetClusterNameFromCurrentContext before")
	}
	logrus.Infof("Creating destination dir '%s'", c.destDir)
	err := os.MkdirAll(c.destDir, os.ModePerm)
	return err
}

func (c *ClusterConfig) GetFilePath(r resources.ResourceType) string {
	return path.Join(c.destDir, r.String())
}

func (c *ClusterConfig) FileStoreExists(r resources.ResourceType) bool {
	p := c.GetFilePath(r)
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

func (c *ClusterConfig) getRawConfig() (*clientcmdapi.Config, error) {
	configInBytes, err := ioutil.ReadFile(c.kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "error reading kubeconfig file")
	}
	clientConfig, err := clientcmd.NewClientConfigFromBytes(configInBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error creating clientConfig from kubeconfig file")
	}
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, errors.Wrap(err, "error getting rawconfig from clientConfig")
	}
	return &rawConfig, nil
}

func (c *ClusterConfig) GetNamespace() (string, error) {
	rawConfig, err := c.getRawConfig()
	if err != nil {
		return "", err
	}
	contextStruct, ok := rawConfig.Contexts[rawConfig.CurrentContext]
	if !ok {
		return "", fmt.Errorf("context %s not found in config", rawConfig.CurrentContext)
	}
	return contextStruct.Namespace, nil
}

func (c *ClusterConfig) GetContext() string {
	return c.clusterName
}

func (c *ClusterConfig) GetClientConfig() (*rest.Config, error) {
	if util.FileExists(c.kubeconfig) {
		logrus.Tracef("kubeconfig file '%s' exists, reading config", c.kubeconfig)
		return clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	}
	logrus.Tracef("%s doesn't exists, assuming incluster setup", c.kubeconfig)
	return rest.InClusterConfig()
}
