package httpservertest

import (
	"context"
	"os"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher/fetchertest"
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
	fzfHttpServer := StartTestHttpServer(t)
	f, _ := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	ctx := context.Background()
	s, err := f.GetStats(ctx)
	require.NoError(t, err)
	assert.Len(t, s, 1)
}
