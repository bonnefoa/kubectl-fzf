package httpservertest

import (
	"context"
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

func TestHttpServerApiCompletion(t *testing.T) {
	f := StartTestHttpServer(t)
	ctx := context.Background()
	s, err := f.GetStats(ctx)
	require.NoError(t, err)
	assert.Len(t, s, 1)
}
