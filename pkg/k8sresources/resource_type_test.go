package k8sresources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResourceType(t *testing.T) {
	r := ParseResourceType("pods")
	assert.Equal(t, r, ResourceTypePod)
}
