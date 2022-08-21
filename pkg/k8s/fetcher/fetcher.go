package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"kubectlfzf/pkg/k8s/clusterconfig"
	"kubectlfzf/pkg/k8s/portforward"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/util"
	"path"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Fetcher defines configuration to fetch completion datas
type Fetcher struct {
	clusterconfig.ClusterConfig
	httpEndpoint         string
	portForwardLocalPort int // Local port to use for port-forward
}

func NewFetcher(fetchConfigCli *FetcherCli) *Fetcher {
	f := Fetcher{}
	f.ClusterConfig = clusterconfig.NewClusterConfig(&fetchConfigCli.ClusterConfigCli)
	f.httpEndpoint = fetchConfigCli.HttpEndpoint
	f.portForwardLocalPort = fetchConfigCli.PortForwardLocalPort
	return &f
}

func (f *Fetcher) getResourceHttpPath(host string, r resources.ResourceType) string {
	fullPath := path.Join("k8s", "resources", r.String())
	return fmt.Sprintf("http://%s/%s", host, fullPath)
}

func (f *Fetcher) httpAddressReachable() bool {
	logrus.Debugf("Checking if %s is reachable", f.httpEndpoint)
	return util.IsAddressReachable(f.httpEndpoint)
}

func (f *Fetcher) getPortForwardRequest(ctx context.Context, r resources.ResourceType) (portForwardRequest portforward.PortForwardRequest, err error) {
	logrus.Debugf("Falling back to port forwarding")
	ns := "default"
	listOptions := metav1.ListOptions{
		LabelSelector: "app=kubectl-fzf",
		FieldSelector: "status.phase=Running",
	}
	clientset, err := f.GetClientset()
	if err != nil {
		return
	}
	podList, err := clientset.CoreV1().Pods(ns).List(ctx, listOptions)
	if err != nil {
		return
	}
	if len(podList.Items) == 0 {
		err = fmt.Errorf("no kubectl-fzf pods found, bailing out")
		return
	}
	pod := podList.Items[0]
	if len(pod.Spec.Containers) != 1 {
		err = fmt.Errorf("kubectl-fzf pod should have only one container, got %d", len(pod.Spec.Containers))
		return
	}
	containerPorts := pod.Spec.Containers[0].Ports
	if len(containerPorts) != 1 {
		err = fmt.Errorf("kubectl-fzf container should have only one port, got %d", len(containerPorts))
		return
	}
	podPort := int(containerPorts[0].ContainerPort)
	if podPort <= 0 {
		err = fmt.Errorf("container port invalid, should be > 0, got %d", podPort)
		return
	}
	portForwardRequest = portforward.NewPortForwardRequest(pod.Name, pod.Namespace, f.portForwardLocalPort, podPort)
	logrus.Infof("Found a kubectl-fzf pod found, trying port-forward to %s", pod.Name)
	return
}

func loadFromFile(filePath string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, filePath)
	return resources, err
}

func loadResourceFromHttpServer(url string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromHttpServer(&resources, url)
	return resources, err
}

func (f *Fetcher) getResourcesFromPortForward(ctx context.Context, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})
	errChan := make(chan error)
	portForwardRequest, err := f.getPortForwardRequest(ctx, r)
	if err != nil {
		return nil, err
	}
	go func() {
		restConfig, err := f.GetClientConfig()
		if err != nil {
			errChan <- err
		}
		err = portforward.OpenPortForward(restConfig, portForwardRequest, readyChan, stopChan)
		if err != nil {
			errChan <- err
		}
	}()
	select {
	case err := <-errChan:
		return nil, errors.Wrap(err, "error opening port forward")
	case <-readyChan:
	}
	close(errChan)
	logrus.Debug("Port forward ready")
	httpPath := f.getResourceHttpPath("localhost:8080", r)
	resources, err := loadResourceFromHttpServer(httpPath)
	stopChan <- struct{}{}
	return resources, err
}

func (f *Fetcher) GetStatsFromHttpServer(ctx context.Context) ([]*store.Stats, error) {
	url := fmt.Sprintf("http://%s/%s", f.httpEndpoint, "stats")
	b, err := util.GetBodyFromHttpServer(url)
	if err != nil {
		return nil, errors.Wrap(err, "error on http get")
	}
	stats := make([]*store.Stats, 0)
	logrus.Debugf("Received stats: %s", b)
	err = json.Unmarshal(b, &stats)
	return stats, err
}

func (f *Fetcher) GetStats(ctx context.Context) ([]*store.Stats, error) {
	logrus.Debugf("GetStats httpendpoint: %s", f.httpEndpoint)
	// TODO Handle local file
	if f.httpEndpoint != "" && f.httpAddressReachable() {
		logrus.Debugf("Fetching stats from %s", f.httpEndpoint)
		return f.GetStatsFromHttpServer(ctx)
	}
	// TODO Handle port forward
	return nil, nil
}

func (f *Fetcher) GetResources(ctx context.Context, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	if f.FileStoreExists(r) {
		filePath := f.GetFilePath(r)
		logrus.Debugf("%s found, using resources from file", filePath)
		resources, err := loadFromFile(filePath)
		return resources, err
	}
	if f.httpEndpoint != "" && f.httpAddressReachable() {
		httpPath := f.getResourceHttpPath(f.httpEndpoint, r)
		logrus.Debugf("Using %s for completion", httpPath)
		return loadResourceFromHttpServer(httpPath)
	}
	return f.getResourcesFromPortForward(ctx, r)
}
