package fetcher

import (
	"context"
	"fmt"
	"path"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/portforward"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Fetcher) loadResourceFromHttpServer(endpoint string, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	resources, err := f.checkHttpCache(endpoint, r)
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
	logrus.Infof("Getting resources %s from port forward", r)
	stopChan, err := f.openPortForward(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("localhost:%d", f.portForwardLocalPort)
	resources, err := f.loadResourceFromHttpServer(endpoint, r)
	stopChan <- struct{}{}
	return resources, err
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
	ns := f.fetcherState.getFzfNamespace(f.GetContext())
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
	f.fetcherState.updateNamespace(f.GetContext(), pod.GetNamespace())
	return &pod, nil
}

func (f *Fetcher) getPortForwardRequest(ctx context.Context) (portForwardRequest portforward.PortForwardRequest, err error) {
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
