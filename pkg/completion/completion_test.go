package completion

import (
	"context"
	"kubectlfzf/pkg/httpserver/httpservertest"
	"kubectlfzf/pkg/k8s/resources"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cmdArg struct {
	verb string
	args []string
}

func TestProcessResourceName(t *testing.T) {
	fetchConfig := httpservertest.GetTestFetchConfig(t)
	cmdArgs := []cmdArg{
		{"get", []string{"get", "pods", ""}},
		{"get", []string{"po", ""}},
		{"logs", []string{""}},
		{"exec", []string{"-ti", ""}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		assert.Contains(t, comps[0], "minikube\tkube-system\tcoredns-6d4b75cb6d-m6m4q\t172.17.0.3\t192.168.49.2\tminikube\tRunning\tBurstable\tcoredns\tCriticalAddonsOnly:,node-role.kubernetes.io/master:NoSchedule,node-role.kubernetes.io/control-plane:NoSchedule\tNone")
	}
}

func TestProcessLabelCompletion(t *testing.T) {
	fetchConfig := httpservertest.GetTestFetchConfig(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "-l="}},
		{"get", []string{"pods", "-l"}},
		{"get", []string{"pods", "-l", ""}},
		{"get", []string{"pods", "--selector", ""}},
		{"get", []string{"pods", "--selector"}},
		{"get", []string{"pods", "--selector="}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		assert.Equal(t, "minikube\tkube-system\ttier=control-plane\t4", comps[0])
		assert.Len(t, comps, 12)
	}
}

func TestProcessFieldSelectorCompletion(t *testing.T) {
	fetchConfig := httpservertest.GetTestFetchConfig(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "--field-selector", ""}},
		{"get", []string{"pods", "--field-selector"}},
		{"get", []string{"pods", "--field-selector="}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		assert.Equal(t, "minikube\tkube-system\tspec.nodeName=minikube\t7", comps[0])
	}
}

func TestPodCompletionFile(t *testing.T) {
	fetchConfig := httpservertest.GetTestFetchConfig(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, fetchConfig)
	require.NoError(t, err)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "minikube\tkube-system\t")
	assert.Len(res, 7)
}

func TestNamespaceFilterFile(t *testing.T) {
	fetchConfig := httpservertest.GetTestFetchConfig(t)

	// everything is filtered
	namespace := "test"
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, &namespace, fetchConfig)
	require.NoError(t, err)
	t.Log(res)
	assert := assert.New(t)
	assert.Len(res, 0)

	// all results match
	namespace = "kube-system"
	res, err = getResourceCompletion(context.Background(), resources.ResourceTypePod, &namespace, fetchConfig)
	assert.Len(res, 7)
	require.NoError(t, err)
}

func TestApiResourcesFile(t *testing.T) {
	fetchConfig := httpservertest.GetTestFetchConfig(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypeApiResource, nil, fetchConfig)
	require.NoError(t, err)
	assert := assert.New(t)
	sort.Strings(res)
	assert.Contains(res[0], "apiservices\tNone\tapiregistration.k8s.io/v1\tfalse\tAPIService")
}

func TestHttpServerApiCompletion(t *testing.T) {
	f := httpservertest.StartTestHttpServer(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypeApiResource, nil, f)
	require.NoError(t, err)
	sort.Strings(res)
	assert.Contains(t, res[0], "apiservices\tNone\tapiregistration.k8s.io/v1\tfalse\tAPIService")
	assert.Len(t, res, 56)
}

func TestHttpServerPodCompletion(t *testing.T) {
	f := httpservertest.StartTestHttpServer(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, f)
	require.NoError(t, err)
	assert.Contains(t, res[0], "minikube\tkube-system\t")
	assert.Len(t, res, 7)
}

func TestHttpUnknownResourceCompletion(t *testing.T) {
	f := httpservertest.StartTestHttpServer(t)
	_, err := getResourceCompletion(context.Background(), resources.ResourceTypePersistentVolume, nil, f)
	require.Error(t, err)
}
