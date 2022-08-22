package httpservertest

import (
	"context"
	"kubectlfzf/pkg/httpserver"
	"kubectlfzf/pkg/k8s/clusterconfig"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/k8s/store/storetest"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetTestClusterConfigCli() clusterconfig.ClusterConfigCli {
	return clusterconfig.ClusterConfigCli{ClusterName: "minikube", CacheDir: "./testdata", Kubeconfig: ""}
}

func GetTestStoreConfigCli() *store.StoreConfigCli {
	return &store.StoreConfigCli{ClusterConfigCli: GetTestClusterConfigCli()}
}

func StartTestHttpServer(t *testing.T) int {
	ctx := context.Background()
	storeConfigCli := GetTestStoreConfigCli()
	storeConfig := store.NewStoreConfig(storeConfigCli)
	_, podStore := storetest.GetTestPodStore(t)
	h := &httpserver.HttpServerConfigCli{ListenAddress: "localhost:0", Debug: false}
	port, err := httpserver.StartHttpServer(ctx, h, storeConfig, []*store.Store{podStore})
	require.NoError(t, err)
	return port
}
