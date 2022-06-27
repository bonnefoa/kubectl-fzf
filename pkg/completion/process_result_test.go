package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSimpleRsult(t *testing.T) {
	res, err := processResultWithNamespace("minikube kube-system kube-controller-manager-minikube", "kubectl get pods ")
	assert.NoError(t, err)
	assert.Equal(t, res, "kubectl get pods kube-controller-manager-minikube")
}
