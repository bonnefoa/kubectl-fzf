package util

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoding(t *testing.T) {
	data := "test"

	f, err := ioutil.TempFile("", "encoding")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	err = EncodeToFile(data, f.Name())
	require.NoError(t, err)

	var res string
	err = LoadGobFromFile(&res, f.Name())
	require.NoError(t, err)
	assert.Equal(t, "test", res)
}
