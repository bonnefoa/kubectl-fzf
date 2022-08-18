package completion

import (
	"context"
	"kubectlfzf/pkg/k8s/resources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagLabel(t *testing.T) {
	fetchConfig := getTestFetchConfig(t)
	labelMap, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeLabel)
	assert.NoError(t, err)
	t.Log(labelMap)

	assert.Contains(t, labelMap, TagResourceKey{"minikube", "kube-system", "k8s-app=kube-dns"})
	assert.Contains(t, labelMap, TagResourceKey{"minikube", "kube-system", "tier=control-plane"})
	assert.Equal(t, 4, labelMap[TagResourceKey{"minikube", "kube-system", "tier=control-plane"}])
}

func TestLabelNamespaceFiltering(t *testing.T) {
	fetchConfig := getTestFetchConfig(t)
	namespace := "default"
	labelMap, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, &namespace, fetchConfig, TagTypeLabel)
	assert.NoError(t, err)
	assert.Len(t, labelMap, 0)
}

func TestLabelCompletion(t *testing.T) {
	fetchConfig := getTestFetchConfig(t)
	labelComps, err := GetTagResourceCompletion(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeLabel)
	assert.NoError(t, err)
	assert.Len(t, labelComps, 12)

	t.Log(labelComps)
	assert.Equal(t, "minikube\tkube-system\ttier=control-plane\t4", labelComps[0])
	assert.Equal(t, "minikube\tkube-system\taddonmanager.kubernetes.io/mode=Reconcile\t1", labelComps[1])
}

func TestGetFieldSelector(t *testing.T) {
	fetchConfig := getTestFetchConfig(t)
	fieldSelectorOccurrences, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeFieldSelector)
	require.NoError(t, err)

	assert.Contains(t, fieldSelectorOccurrences, TagResourceKey{"minikube", "kube-system", "spec.nodeName=minikube"})
	assert.Equal(t, 7, fieldSelectorOccurrences[TagResourceKey{"minikube", "kube-system", "spec.nodeName=minikube"}])
}
