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
	res, err := GetResourceCompletion(k8sresources.ResourceTypePod, nil, storeConfig)
	assert.NoError(t, err)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "Cluster")
	assert.Contains(res[1], "minikube kube-system ")
	assert.Len(res, 8)
}

func TestNamespaceFilter(t *testing.T) {
	storeConfig := getStoreConfig()
	namespace := "test"
	res, err := GetResourceCompletion(k8sresources.ResourceTypePod, &namespace, storeConfig)
	assert.NoError(t, err)
	t.Log(res)
	assert := assert.New(t)
	assert.Len(res, 1)

	namespace = "kube-system"
	res, err = GetResourceCompletion(k8sresources.ResourceTypePod, &namespace, storeConfig)
	assert.Len(res, 8)
}

func TestApiResources(t *testing.T) {
	storeConfig := getStoreConfig()
	res, err := GetResourceCompletion(k8sresources.ResourceTypeApiResource, nil, storeConfig)
	assert.NoError(t, err)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "Fullname")
}
