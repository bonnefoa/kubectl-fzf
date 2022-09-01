package httpservertest

import (
	"context"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/httpserver"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store/storetest"
	"github.com/stretchr/testify/require"
)

func GetTestClusterConfigCli() *clusterconfig.ClusterConfigCli {
	return &clusterconfig.ClusterConfigCli{ClusterName: "minikube", CacheDir: "./testdata", Kubeconfig: ""}
}

func GetTestStoreConfigCli() *store.StoreConfigCli {
	return &store.StoreConfigCli{ClusterConfigCli: GetTestClusterConfigCli()}
}

func StartTestHttpServer(t *testing.T) *httpserver.FzfHttpServer {
	ctx := context.Background()
	storeConfigCli := GetTestStoreConfigCli()
	storeConfig := store.NewStoreConfig(storeConfigCli)
	_, podStore := storetest.GetTestPodStore(t)
	h := &httpserver.HttpServerConfigCli{ListenAddress: "localhost:0", Debug: false}
	fzfHttpServer, err := httpserver.StartHttpServer(ctx, h, storeConfig, []*store.Store{podStore})
	require.NoError(t, err)
	return fzfHttpServer
}
