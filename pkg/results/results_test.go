package results

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestParseNamespaceFlag(t *testing.T) {
	r, err := parseNamespaceFlag([]string{"get", "pods", "-ntest"})
	require.NoError(t, err)
	assert.Equal(t, "test", *r)

	r, err = parseNamespaceFlag([]string{"get", "pods", "--namespace", "kube-system"})
	require.NoError(t, err)
	assert.Equal(t, "kube-system", *r)

	r, err = parseNamespaceFlag([]string{"get", "pods", "--context", "minikube", "--namespace", "kube-system"})
	require.NoError(t, err)
	assert.Equal(t, "kube-system", *r)
}

func TestResult(t *testing.T) {
	testDatas := []struct {
		fzfResult        string
		cmdUse           string
		cmdArgs          []string
		currentNamespace string
		expectedResult   string
	}{
		{"minikube kube-system kube-controller-manager-minikube", "get", []string{"pods", " "}, "kube-system", "kube-controller-manager-minikube"},
		{"minikube kube-system coredns-64897985d-nrblm", "get", []string{"pods", "--context", "minikube", "--namespace", "kube-system", ""}, "default", "coredns-64897985d-nrblm"},
		{"minikube kube-system kube-controller-manager-minikube", "get", []string{"pods", " "}, "default", "kube-controller-manager-minikube -n kube-system"},
		{"minikube kube-system kube-controller-manager-minikube", "get", []string{"pods", "-nkube-system", " "}, "default", "kube-controller-manager-minikube"},
		// Namespace
		{"incluster default 30d kubernetes.io/metadata.name=default", "get", []string{"pods", "-n="}, "default", "-n=default"},
		{"incluster default 30d kubernetes.io/metadata.name=default", "get", []string{"pods", "-n"}, "default", "-ndefault"},
		{"incluster default 30d kubernetes.io/metadata.name=default", "get", []string{"pods", "-n", " "}, "default", "default"},
		// Label
		{"minikube kube-system tier=control-plane", "get", []string{"pods", "-l="}, "default", "-l=tier=control-plane -n kube-system"},
		{"minikube kube-system tier=control-plane", "get", []string{"pods", "-l", " "}, "default", "tier=control-plane -n kube-system"},
		{"minikube kube-system tier=control-plane", "get", []string{"pods", "-l"}, "default", "-ltier=control-plane -n kube-system"},
		// Field selector
		{"minikube kube-system spec.nodeName=minikube", "get", []string{"pods", "--field-selector="}, "default", "--field-selector=spec.nodeName=minikube -n kube-system"},
		{"minikube kube-system spec.nodeName=minikube", "get", []string{"pods", "--field-selector", " "}, "default", "spec.nodeName=minikube -n kube-system"},
		{"minikube kube-system coredns-64897985d-nrblm", "get", []string{"pods", "c"}, "default", "coredns-64897985d-nrblm -n kube-system"},
		{"apiservices.apiregistration.k8s.io None apiregistration.k8s.io/v1", "get", []string{" "}, "default", "apiservices.apiregistration.k8s.io"},
	}
	for _, testData := range testDatas {
		res, err := processResultWithNamespace(testData.cmdUse, testData.cmdArgs, testData.fzfResult, testData.currentNamespace)
		require.NoError(t, err)
		require.Equal(t, testData.expectedResult, res,
			"Fzf result %s, cmdUse %s, cmdArgs %s, current namespace %s, res: %s", testData.fzfResult, testData.cmdUse,
			testData.cmdArgs, testData.currentNamespace, res)
	}
}
