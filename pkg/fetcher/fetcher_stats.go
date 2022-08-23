package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/util"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (f *Fetcher) getStatsFromHttpServer(ctx context.Context, url string) ([]*store.Stats, error) {
	logrus.Debugf("Fetching stats from %s", url)
	_, body, err := util.GetFromHttpServer(url)
	if err != nil {
		return nil, errors.Wrap(err, "error on http get")
	}
	stats := make([]*store.Stats, 0)
	err = json.Unmarshal(body, &stats)
	return stats, err
}

func (f *Fetcher) getStatsFromPortForward(ctx context.Context) ([]*store.Stats, error) {
	stopChan, err := f.openPortForward(ctx)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/%s", fmt.Sprintf("localhost:%d", f.portForwardLocalPort), "stats")
	stats, err := f.getStatsFromHttpServer(ctx, url)
	stopChan <- struct{}{}
	return stats, err
}

func (f *Fetcher) GetStats(ctx context.Context) ([]*store.Stats, error) {
	// TODO Handle local file
	if util.IsAddressReachable(f.httpEndpoint) {
		url := fmt.Sprintf("http://%s/%s", f.httpEndpoint, "stats")
		return f.getStatsFromHttpServer(ctx, url)
	}
	return f.getStatsFromPortForward(ctx)
}
