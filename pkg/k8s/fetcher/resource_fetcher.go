package fetcher

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/util"
	"path"

	"github.com/sirupsen/logrus"
)

func loadResourceFromFile(filePath string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, filePath)
	return resources, err
}

func loadResourceFromHttpServer(url string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromHttpServer(&resources, url)
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
	resources, err := loadResourceFromHttpServer(httpPath)
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
		return loadResourceFromHttpServer(httpPath)
	}
	return f.getResourcesFromPortForward(ctx, r)
}
