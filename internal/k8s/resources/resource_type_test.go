package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResourceType(t *testing.T) {
	testDatas := []struct {
		resourceName string
		resourceType ResourceType
	}{
		{"", ResourceTypeUnknown},
		{"pods", ResourceTypePod},
		{"pod", ResourceTypePod},
		{"statefulsets", ResourceTypeStatefulSet},
		{"sts", ResourceTypeStatefulSet},
	}

	for _, v := range testDatas {
		r := ParseResourceType(v.resourceName)
		require.Equal(t, v.resourceType, r, "Expected %s, got %s", v.resourceType, r)
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

func TestGetResourceSetFromSliceWithErrors(t *testing.T) {
	testDatas := [][]string{
		{"po", "t", "secrets"},
		{"saa", "pod"},
	}
	for _, testData := range testDatas {
		_, err := GetResourceSetFromSlice(testData)
		require.Error(t, err)
	}
}

func TestGetResourceSetFromSlice(t *testing.T) {
	testDatas := [][]string{
		{"pods", "secrets"},
		{"sa", "pod"},
	}
	for _, testData := range testDatas {
		_, err := GetResourceSetFromSlice(testData)
		require.NoError(t, err)
	}
}
