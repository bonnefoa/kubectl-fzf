package store

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestDumpAPIResources(t *testing.T) {
	resource := map[string]resources.K8sResource{}

	list := resources.APIResourceList{}
	list.GroupVersion = "v1"

	a := resources.APIResource{}
	a.Shortnames = []string{"short"}
	a.Name = "name"
	list.ApiResources = append(list.ApiResources, a)

	resource["v1"] = &list
	tempDir, err := ioutil.TempDir("/tmp/", "cacheTest")
	require.NoError(t, err)

	apiResourcesFilePath := path.Join(tempDir, "apiresources")
	err = util.EncodeToFile(resource, apiResourcesFilePath)
	require.NoError(t, err)

	loadResource := map[string]resources.K8sResource{}
	err = util.LoadGobFromFile(&loadResource, apiResourcesFilePath)
	require.NoError(t, err)
}
