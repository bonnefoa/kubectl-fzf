package fetcher

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/util"
	"path"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func loadResourceFromFile(filePath string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, filePath)
	return resources, err
}

func (f *Fetcher) loadResourceFromHttpServer(endpoint string, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	resources, err := f.checkLocalCache(endpoint, r)
	if err != nil {
		logrus.Infof("Error getting resources from cache: %s", err)
	}
	if resources != nil {
		logrus.Infof("Returning %s resources from cache", r.String())
		return resources, nil
	}
	logrus.Debugf("Loading from %s", endpoint)
	resourcePath := f.getResourceHttpPath(endpoint, r)
	headers, body, err := util.GetFromHttpServer(resourcePath)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body content")
	}
	err = f.writeResourceToCache(headers, body, r)
	if err != nil {
		return nil, errors.Wrap(err, "error writing fetcher cache")
	}
	util.DecodeGob(&resources, body)
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
	endpoint := fmt.Sprintf("localhost:%d", f.portForwardLocalPort)
	resources, err := f.loadResourceFromHttpServer(endpoint, r)
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
	if util.IsAddressReachable(f.httpEndpoint) {
		return f.loadResourceFromHttpServer(f.httpEndpoint, r)
	}
	return f.getResourcesFromPortForward(ctx, r)
}
