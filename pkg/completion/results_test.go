package completion

import (
	"os"
	"strings"
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
		sourceCmd        string
		currentNamespace string
		expectedResult   string
	}{
		{"minikube kube-system kube-controller-manager-minikube", "get pods ", "kube-system", "kube-controller-manager-minikube"},
		{"minikube kube-system coredns-64897985d-nrblm", "get pods --context minikube --namespace kube-system ", "default", "coredns-64897985d-nrblm"},
		{"minikube kube-system kube-controller-manager-minikube", "get pods ", "default", "kube-controller-manager-minikube -n kube-system"},
		{"minikube kube-system kube-controller-manager-minikube", "get pods -nkube-system ", "default", "kube-controller-manager-minikube"},
		// Label
		{"minikube kube-system tier=control-plane", "get pods -l=", "default", "-l=tier=control-plane -n kube-system"},
		{"minikube kube-system tier=control-plane", "get pods -l ", "default", "tier=control-plane -n kube-system"},
		{"minikube kube-system tier=control-plane", "get pods -l", "default", "-ltier=control-plane -n kube-system"},
		// Field selector
		{"minikube kube-system spec.nodeName=minikube", "get pods --field-selector=", "default", "--field-selector=spec.nodeName=minikube -n kube-system"},
		{"minikube kube-system spec.nodeName=minikube", "get pods --field-selector ", "default", "spec.nodeName=minikube -n kube-system"},
		{"minikube kube-system coredns-64897985d-nrblm", "get pods c", "default", "coredns-64897985d-nrblm -n kube-system"},
		{"apiservices.apiregistration.k8s.io None apiregistration.k8s.io/v1", "get ", "default", "apiservices.apiregistration.k8s.io"},
	}
	for _, testData := range testDatas {
		cmdArgs := strings.Split(testData.sourceCmd, " ")
		res, err := processResultWithNamespace(testData.fzfResult, cmdArgs, testData.currentNamespace)
		assert.NoError(t, err)
		assert.Equal(t, testData.expectedResult, res, "Fzf result %s, source cmd %s, current namespace %s, res: %s", testData.fzfResult, testData.sourceCmd, testData.currentNamespace, res)
	}
}
