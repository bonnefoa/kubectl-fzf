package store

import (
	"kubectlfzf/pkg/k8s/clusterconfig"
	"kubectlfzf/pkg/k8s/resources"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileStoreExists(t *testing.T) {
	c := &StoreConfigCli{
		ClusterConfigCli: clusterconfig.ClusterConfigCli{
			ClusterName: "minikube",
			CacheDir:    "./testdata", Kubeconfig: "",
		}, TimeBetweenFullDump: 1 * time.Second}
	s := NewStoreConfig(c)
	assert.True(t, s.FileStoreExists(resources.ResourceTypePod))
	assert.False(t, s.FileStoreExists(resources.ResourceTypeApiResource))
}
