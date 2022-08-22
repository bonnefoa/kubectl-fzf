package fetcher

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/util"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func loadResourceFromFile(filePath string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, filePath)
	return resources, err
}

func (f *Fetcher) writeResourceToCache(b []byte, r resources.ResourceType) error {
	destDir := path.Join(f.fetcherCachePath, f.GetContext())
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return errors.Wrap(err, "error mkdirall")
	}
	cachePath := path.Join(destDir, r.String())
	logrus.Debugf("Caching resource in %s", cachePath)
	err = os.WriteFile(cachePath, b, 0644)
	if err != nil {
		return errors.Wrap(err, "error writing cache file")
	}
	return nil
}

func (f *Fetcher) loadResourceFromHttpServer(url string, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	logrus.Debugf("Loading from %s", url)
	b, err := util.GetBodyFromHttpServer(url)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body content")
	}
	err = f.writeResourceToCache(b, r)
	if err != nil {
		return nil, errors.Wrap(err, "error writing fetcher cache")
	}

	resources := map[string]resources.K8sResource{}
	util.DecodeGob(&resources, b)
	return resources, err
}

func (f *Fetcher) getResourceHttpPath(host string, r resources.ResourceType) string {
	fullPath := path.Join("k8s", "resources", r.String())
	return fmt.Sprintf("http://%s/%s", host, fullPath)
}

func (f *Fetcher) getResourcesFromPortForward(ctx context.Context, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	stopChan, err := f.openPortForward(ctx)
	if err != nil {
		return nil, err
	}
	httpPath := f.getResourceHttpPath(fmt.Sprintf("localhost:%d", f.portForwardLocalPort), r)
	resources, err := f.loadResourceFromHttpServer(httpPath, r)
	stopChan <- struct{}{}
	return resources, err
}

func (f *Fetcher) GetResources(ctx context.Context, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	if f.FileStoreExists(r) {
		filePath := f.GetFilePath(r)
		logrus.Debugf("%s found, using resources from file", filePath)
		resources, err := loadResourceFromFile(filePath)
		return resources, err
	}
	if f.httpEndpoint != "" && f.httpAddressReachable() {
		httpPath := f.getResourceHttpPath(f.httpEndpoint, r)
		logrus.Debugf("Using %s for completion", httpPath)
		return f.loadResourceFromHttpServer(httpPath, r)
	}
	return f.getResourcesFromPortForward(ctx, r)
}
