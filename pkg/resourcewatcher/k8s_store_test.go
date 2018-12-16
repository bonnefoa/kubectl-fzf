package resourcewatcher

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDumpFullState(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp/", "cacheTest")
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir)

	k, err := NewK8sStore(func() K8sResource { return &Pod{} }, "pods", tempDir)
	assert.Nil(t, err)

	k.data["test"] = &Pod{
		ResourceMeta: ResourceMeta{name: "Test", creationTime: time.Now()},
	}
	k.data["test2"] = &Pod{
		ResourceMeta: ResourceMeta{name: "Test2", creationTime: time.Now()},
	}

	k.DumpFullState()

	k.currentFile.Seek(0, 0)
	scanner := bufio.NewScanner(k.currentFile)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	assert.NoError(t, scanner.Err(), "Scanner")
}
