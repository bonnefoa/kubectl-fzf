package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatCompletion(t *testing.T) {
	res := FormatCompletion("header1\thead2", []string{"comp1\tc1", "c2\tc22"})
	expected := `header1 head2
comp1   c1
c2      c22
`
	require.Equal(t, expected, res)
}
