package store

import (
	"testing"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
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
