package completion

import (
	"context"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher/fetchertest"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagLabel(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	labelMap, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeLabel)
	assert.NoError(t, err)
	t.Log(labelMap)

	assert.Contains(t, labelMap, TagResourceKey{"kube-system", "k8s-app=kube-dns"})
	assert.Contains(t, labelMap, TagResourceKey{"kube-system", "tier=control-plane"})
	assert.Equal(t, 4, labelMap[TagResourceKey{"kube-system", "tier=control-plane"}])
}

func TestLabelNamespaceFiltering(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	namespace := "default"
	labelMap, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, &namespace, fetchConfig, TagTypeLabel)
	assert.NoError(t, err)
	assert.Len(t, labelMap, 0)
}

func TestLabelCompletionPod(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	labelHeader, labelComps, err := GetTagResourceCompletion(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeLabel)
	assert.NoError(t, err)
	assert.Len(t, labelComps, 12)

	t.Log(labelComps)
	assert.Equal(t, "Namespace\tLabel\tOccurrences", labelHeader)
	assert.Equal(t, "kube-system\ttier=control-plane\t4", labelComps[0])
	assert.Equal(t, "kube-system\taddonmanager.kubernetes.io/mode=Reconcile\t1", labelComps[1])
}

func TestLabelCompletionNode(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	labelHeader, labelComps, err := GetTagResourceCompletion(context.Background(), resources.ResourceTypeNode, nil, fetchConfig, TagTypeLabel)
	assert.NoError(t, err)
	assert.Len(t, labelComps, 12)

	t.Log(labelComps)
	assert.Equal(t, "Label\tOccurrences", labelHeader)
	assert.Equal(t, "beta.kubernetes.io/arch=amd64\t1", labelComps[0])
	assert.Equal(t, "beta.kubernetes.io/os=linux\t1", labelComps[1])
}

func TestGetFieldSelector(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	fieldSelectorOccurrences, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeFieldSelector)
	require.NoError(t, err)

	assert.Contains(t, fieldSelectorOccurrences, TagResourceKey{"kube-system", "spec.nodeName=minikube"})
	assert.Equal(t, 7, fieldSelectorOccurrences[TagResourceKey{"kube-system", "spec.nodeName=minikube"}])
}
