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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDumpFullState(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp/", "cacheTest")
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir)

	cfg := WatchConfig{
		resourceName: "pods",
	}
	storeConfig := StoreConfig{
		CacheDir: tempDir,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctorConfig := k8sresources.CtorConfig{}
	k, err := NewK8sStore(ctx, cfg, storeConfig, ctorConfig)
	assert.Nil(t, err)

	meta1 := metav1.ObjectMeta{Name: "Test",
		CreationTimestamp: metav1.Time{Time: time.Now()}}
	meta2 := metav1.ObjectMeta{Name: "Test2",
		CreationTimestamp: metav1.Time{Time: time.Now()}}

	pod1 := &k8sresources.Pod{}
	pod1.FromObjectMeta(meta1)
	pod2 := &k8sresources.Pod{}
	pod2.FromObjectMeta(meta2)
	k.data["test"] = pod1
	k.data["test2"] = pod2

	err = k.DumpFullState()
	assert.Nil(t, err)

	f, err := os.OpenFile(path.Join(tempDir, "pods_resource"), os.O_RDONLY, 0644)
	assert.Nil(t, err)

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	assert.Equal(t, len(lines), 2)
	assert.Contains(t, strings.Split(lines[0], " "), "Test")
	assert.Contains(t, strings.Split(lines[1], " "), "Test2")
	assert.NoError(t, scanner.Err(), "Scanner")
}
