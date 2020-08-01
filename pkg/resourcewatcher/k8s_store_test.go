package resourcewatcher

import (
	"bufio"
	"context"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"kubectlfzf/pkg/k8sresources"

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
	defer os.RemoveAll(tempDir)

	cfg := WatchConfig{
		k8sresources.NewPodFromRuntime, k8sresources.PodHeader, string(corev1.ResourcePods), nil, &corev1.Pod{}, true, true, 0,
	}
	storeConfig := StoreConfig{
		CacheDir: tempDir,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctorConfig := k8sresources.CtorConfig{}
	k, err := NewK8sStore(ctx, cfg, storeConfig, ctorConfig)
	assert.Nil(t, err)

	pods := []corev1.Pod{
		podResource("Test1", "ns1", map[string]string{"app": "app1"}),
		podResource("Test2", "ns2", map[string]string{"app": "app2"}),
		podResource("Test3", "ns2", map[string]string{"app": "app2"}),
		podResource("Test4", "aaa", map[string]string{"app": "app3"}),
	}

	for _, pod := range pods {
		k.AddResource(&pod)
	}
	return tempDir, &k
}

func TestDumpFullState(t *testing.T) {
	tempDir, k := getK8sStore(t)
	err := k.DumpFullState()
	assert.Nil(t, err)

	f, err := os.OpenFile(path.Join(tempDir, "pods_resource"), os.O_RDONLY, 0644)
	assert.Nil(t, err)

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	assert.Equal(t, len(lines), 4)
	assert.Contains(t, strings.Split(lines[0], " "), "Test4")
	assert.Contains(t, strings.Split(lines[1], " "), "Test1")
	assert.NoError(t, scanner.Err(), "Scanner")
}

func TestSortedPairList(t *testing.T) {
	_, k := getK8sStore(t)
	labelStr, err := k.generateLabel()
	assert.NoError(t, err, "Generate label")
	split := strings.Split(labelStr, "\n")
	t.Log(labelStr)
	assert.Equal(t, len(split), 3)
	assert.Contains(t, split[0], "app2")
	assert.Contains(t, split[1], "app3")
	assert.Contains(t, split[2], "app1")
}
