package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResourceType(t *testing.T) {
	testDatas := []struct {
		resourceName string
		resourceType ResourceType
	}{
		{"", ResourceTypeUnknown},
		{"pods", ResourceTypePod},
		{"pod", ResourceTypePod},
		{"statefulsets.apps", ResourceTypeStatefulSet},
	}

	for _, v := range testDatas {
		r := ParseResourceType(v.resourceName)
		assert.Equal(t, v.resourceType, r)
	}
}

func TestGetResourceType(t *testing.T) {
	testDatas := []struct {
		args         []string
		resourceType ResourceType
	}{
		{[]string{""}, ResourceTypeApiResource},
		{[]string{"pods"}, ResourceTypeApiResource},
		{[]string{"pods", ""}, ResourceTypePod},
	}
	for _, testData := range testDatas {
		parsedType := GetResourceType("get", testData.args)
		assert.Equal(t, testData.resourceType, parsedType, "Args: %s, type %s, result: %s", testData.args, testData.resourceType, parsedType)
	}
}
