package completion

import (
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getStoreConfig() *k8sresources.StoreConfig {
	clusterCliConf := util.ClusterCliConf{ClusterName: "minikube",
		InCluster: false, CacheDir: "./testdata", Kubeconfig: ""}
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, time.Second)
	return storeConfig
}

func TestGetPods(t *testing.T) {
	storeConfig := getStoreConfig()
	res := GetResourceCompletion(k8sresources.ResourceTypePod, storeConfig)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "Cluster")
}

func TestApiResources(t *testing.T) {
	storeConfig := getStoreConfig()
	res := GetResourceCompletion(k8sresources.ResourceTypeApiResource, storeConfig)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "Fullname")
}
