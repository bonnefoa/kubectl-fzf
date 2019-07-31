package resourcewatcher

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
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

	cfg := watchConfig{
		resourceName: "pods",
	}
	storeConfig := StoreConfig{
		CacheDir: tempDir,
	}
	k, err := NewK8sStore(cfg, storeConfig)
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

	k.DumpFullState()

	k.currentFile.Seek(0, 0)
	scanner := bufio.NewScanner(k.currentFile)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	assert.NoError(t, scanner.Err(), "Scanner")
}
