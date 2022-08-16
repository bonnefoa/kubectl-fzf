package completion

import (
	"kubectlfzf/pkg/k8s/resources"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResourceType(t *testing.T) {
	testDatas := []struct {
		args         []string
		resourceType resources.ResourceType
	}{
		{[]string{""}, resources.ResourceTypeApiResource},
		{[]string{"pods"}, resources.ResourceTypeApiResource},
		{[]string{"pods", ""}, resources.ResourceTypePod},
	}
	for _, testData := range testDatas {
		parsedType := getResourceType("get", testData.args)
		assert.Equal(t, testData.resourceType, parsedType, "Args: %s, type %s, result: %s", testData.args, testData.resourceType, parsedType)
	}
}
