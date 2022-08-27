package fetcher

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/k8s/clusterconfig"
	"kubectlfzf/pkg/k8s/portforward"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Fetcher defines configuration to fetch completion datas
type Fetcher struct {
	clusterconfig.ClusterConfig
	fetcherCachePath     string
	httpEndpoint         string
	fzfNamespace         string
	minimumCache         time.Duration
	portForwardLocalPort int // Local port to use for port-forward
}

func NewFetcher(fetchConfigCli *FetcherCli) *Fetcher {
	f := Fetcher{}
	f.ClusterConfig = clusterconfig.NewClusterConfig(&fetchConfigCli.ClusterConfigCli)
	f.httpEndpoint = fetchConfigCli.HttpEndpoint
	f.fzfNamespace = fetchConfigCli.FzfNamespace
	f.fetcherCachePath = fetchConfigCli.FetcherCachePath
	f.minimumCache = fetchConfigCli.MinimumCache
	f.portForwardLocalPort = fetchConfigCli.PortForwardLocalPort
	return &f
}

func (f *Fetcher) getKubectlFzfPod(ctx context.Context) (*corev1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: "app=kubectl-fzf",
		FieldSelector: "status.phase=Running",
	}
	clientset, err := f.GetClientset()
	if err != nil {
		return nil, err
	}
	ns, needNamespaceWrite := f.getCachedNamespace()
	logrus.Infof("Looking for fzf pod in namespace '%s'", ns)
	podList, err := clientset.CoreV1().Pods(ns).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		err = fmt.Errorf("no kubectl-fzf pods found, bailing out")
		return nil, err
	}
	pod := podList.Items[0]
	if len(pod.Spec.Containers) != 1 {
		err = fmt.Errorf("kubectl-fzf pod should have only one container, got %d", len(pod.Spec.Containers))
		return nil, err
	}
	if needNamespaceWrite {
		err = f.writeCachedNamespace(pod.Namespace)
		if err != nil {
			return nil, err
		}
	}
	return &pod, nil
}

func (f *Fetcher) getPortForwardRequest(ctx context.Context) (portForwardRequest portforward.PortForwardRequest, err error) {
	logrus.Debugf("Falling back to port forwarding")
	pod, err := f.getKubectlFzfPod(ctx)
	if err != nil {
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

func (f *Fetcher) openPortForward(ctx context.Context) (chan (struct{}), error) {
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})
	errChan := make(chan error)
	portForwardRequest, err := f.getPortForwardRequest(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create port forward")
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
	return stopChan, nil
}
