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
