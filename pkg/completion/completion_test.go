package completion

import (
	"context"
	"kubectlfzf/pkg/fetcher/fetchertest"
	"kubectlfzf/pkg/httpserver/httpservertest"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/parse"
	"path"
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
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"get", "pods", ""}},
		{"get", []string{"po", ""}},
		{"logs", []string{""}},
		{"exec", []string{"-ti", ""}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		require.Greater(t, len(comps), 0)
		require.Contains(t, comps[0], "minikube\tkube-system\tcoredns-6d4b75cb6d-m6m4q\t172.17.0.3\t192.168.49.2\tminikube\tRunning\tBurstable\tcoredns\tCriticalAddonsOnly:,node-role.kubernetes.io/master:NoSchedule,node-role.kubernetes.io/control-plane:NoSchedule\tNone")
	}
}

func TestProcessNamespace(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "-n"}},
		{"get", []string{"po", "-n="}},
		{"logs", []string{"--namespace", ""}},
		{"logs", []string{"--namespace="}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		require.Greater(t, len(comps), 0)
		require.Contains(t, comps[0], "minikube\tdefault\t30d\tkubernetes.io/metadata.name=default")
	}
}

func TestProcessLabelCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "-l="}},
		{"get", []string{"pods", "-l"}},
		{"get", []string{"pods", "-l", ""}},
		{"get", []string{"pods", "--selector", ""}},
		{"get", []string{"pods", "--selector="}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		require.Equal(t, "minikube\tkube-system\ttier=control-plane\t4", comps[0])
		require.Len(t, comps, 12)
	}
}

func TestProcessFieldSelectorCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "--field-selector", ""}},
		{"get", []string{"pods", "--field-selector="}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		assert.Equal(t, "minikube\tkube-system\tspec.nodeName=minikube\t7", comps[0])
	}
}

func TestUnmanagedCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"-t"}},
		{"get", []string{"-i"}},
		{"get", []string{"--field-selector"}},
		{"get", []string{"--selector"}},
		{"get", []string{"--all-namespaces"}},
	}
	for _, cmdArg := range cmdArgs {
		_, _, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.Error(t, err)
		require.IsType(t, parse.UnmanagedFlagError(""), err)
	}
}

func TestManagedCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "--selector", ""}},
		{"get", []string{"pods", "--selector="}},
		{"get", []string{"pods", "--field-selector", ""}},
		{"get", []string{"pods", "--field-selector="}},
		{"get", []string{"pods", "--all-namespaces", ""}},
		{"get", []string{"pods", "-t", ""}},
		{"get", []string{"pods", "-i", ""}},
		{"get", []string{"pods", "-ti", ""}},
		{"get", []string{"pods", "-it", ""}},
		{"get", []string{"-n"}},
		{"get", []string{"-n", ""}},
	}
	for _, cmdArg := range cmdArgs {
		_, comps, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		require.NoError(t, err)
		require.NotNil(t, comps)
	}
}

func TestPodCompletionFile(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, fetchConfig)
	require.NoError(t, err)
	t.Log(res)
	assert := assert.New(t)
	assert.Contains(res[0], "minikube\tkube-system\t")
	assert.Len(res, 7)
}

func TestNamespaceFilterFile(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)

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
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypeApiResource, nil, fetchConfig)
	require.NoError(t, err)
	assert := assert.New(t)
	sort.Strings(res)
	assert.Contains(res[0], "apiservices\tNone\tapiregistration.k8s.io/v1\tfalse\tAPIService")
}

func TestHttpServerApiCompletion(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypeApiResource, nil, f)
	require.NoError(t, err)
	sort.Strings(res)
	assert.Contains(t, res[0], "apiservices\tNone\tapiregistration.k8s.io/v1\tfalse\tAPIService")
	assert.Len(t, res, 56)

	expectedPath := path.Join(tempDir, "nothing", resources.ResourceTypeApiResource.String())
	assert.FileExists(t, expectedPath)
}

func TestHttpServerPodCompletion(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, f)
	require.NoError(t, err)
	assert.Contains(t, res[0], "minikube\tkube-system\t")
	assert.Len(t, res, 7)

	expectedPath := path.Join(tempDir, "nothing", resources.ResourceTypePod.String())
	assert.FileExists(t, expectedPath)
}

func TestHttpUnknownResourceCompletion(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	_, err := getResourceCompletion(context.Background(), resources.ResourceTypePersistentVolume, nil, f)
	require.Error(t, err)

	expectedPath := path.Join(tempDir, "nothing")
	assert.NoFileExists(t, expectedPath)
}

func TestHttpServerCachePod(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, f)
	require.NoError(t, err)
	assert.Len(t, res, 7)

	podCache := path.Join(tempDir, "nothing", resources.ResourceTypePod.String())
	assert.FileExists(t, podCache)
	require.Equal(t, fzfHttpServer.ResourceHit, 1)
	lastModified := path.Join(tempDir, "nothing", "lastModified")
	assert.FileExists(t, lastModified)

	res, err = getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, f)
	require.Equal(t, fzfHttpServer.ResourceHit, 1)
}
