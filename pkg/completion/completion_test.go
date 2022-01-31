package completion

import (
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHeaderPresent(t *testing.T) {
	clusterCliConf := util.ClusterCliConf{ClusterName: "minikube",
		InCluster: false, CacheDir: "./testdata", Kubeconfig: ""}
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, time.Second)
	res := CompGetApiResources(storeConfig)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "Fullname")
}

func TestGetPods(t *testing.T) {
	clusterCliConf := util.ClusterCliConf{ClusterName: "minikube",
		InCluster: false, CacheDir: "./testdata", Kubeconfig: ""}
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, time.Second)
	res := CompGetResource(k8sresources.ResourceTypePod, storeConfig)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "Cluster")
}
