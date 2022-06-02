package resourcewatcher

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podResource(name string, ns string, labels map[string]string) corev1.Pod {
	meta := corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         ns,
			Labels:            labels,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec:   corev1.PodSpec{},
		Status: corev1.PodStatus{},
	}
	return meta
}

func getK8sStore(t *testing.T) (string, *K8sStore) {
	tempDir, err := ioutil.TempDir("/tmp/", "cacheTest")
	assert.Nil(t, err)

	watchConfig := WatchConfig{
		k8sresources.NewPodFromRuntime, k8sresources.ResourceTypePod, nil, &corev1.Pod{}, true, true, 0,
	}
	clusterCliConf := util.ClusterCliConf{"test", false, tempDir, ""}
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, 5*time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctorConfig := k8sresources.CtorConfig{}
	k8sStore := NewK8sStore(ctx, watchConfig, storeConfig, ctorConfig)
	assert.Nil(t, err)

	pods := []corev1.Pod{
		podResource("Test1", "ns1", map[string]string{"app": "app1"}),
		podResource("Test2", "ns2", map[string]string{"app": "app2"}),
		podResource("Test3", "ns2", map[string]string{"app": "app2"}),
		podResource("Test4", "aaa", map[string]string{"app": "app3"}),
	}

	for _, pod := range pods {
		k8sStore.AddResource(&pod)
	}
	return tempDir, k8sStore
}

func TestDumpFullState(t *testing.T) {
	tempDir, k := getK8sStore(t)
	defer os.RemoveAll(tempDir)

	err := k.DumpFullState()
	assert.NoError(t, err)
	podFilePath := path.Join(tempDir, "test", "pods")
	assert.FileExists(t, podFilePath)

	pods := map[string]k8sresources.K8sResource{}
	err = util.LoadFromFile(&pods, podFilePath)
	assert.NoError(t, err)

	assert.Equal(t, 4, len(pods))
	assert.Contains(t, pods, "ns1_Test1")
	assert.Contains(t, pods, "ns2_Test2")
	assert.Contains(t, pods, "ns2_Test3")
	assert.Contains(t, pods, "aaa_Test4")
}

func TestSortedPairList(t *testing.T) {
	tempDir, k := getK8sStore(t)
	defer os.RemoveAll(tempDir)

	labelStr, err := k.generateLabel()
	assert.NoError(t, err, "Generate label")
	split := strings.Split(labelStr, "\n")
	t.Log(labelStr)
	assert.Equal(t, len(split), 3)
	assert.Contains(t, split[0], "app2")
	assert.Contains(t, split[1], "app3")
	assert.Contains(t, split[2], "app1")
}
