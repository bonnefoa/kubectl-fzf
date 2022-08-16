package portforward

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardRequest struct {
	podName      string
	podNamespace string
	localPort    int
	podPort      int
}

func (p *PortForwardRequest) getPort() []string {
	return []string{fmt.Sprintf("%d:%d", p.localPort, p.podPort)}
}

func (p *PortForwardRequest) getPath() string {
	return fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", p.podNamespace, p.podName)
}

func NewPortForwardRequest(podName, podNamespace string, localPort, podPort int) PortForwardRequest {
	return PortForwardRequest{podName, podNamespace, localPort, podPort}
}

func OpenPortForward(config *restclient.Config, p PortForwardRequest, readyChan, stopChan chan struct{}) error {
	address := []string{"localhost"}
	ports := p.getPort()
	path := p.getPath()

	hostIP := strings.TrimPrefix(config.Host, "https://")
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost,
		&url.URL{Scheme: "https", Path: path, Host: hostIP})
	var berr, bout bytes.Buffer
	buffErr := bufio.NewWriter(&berr)
	buffOut := bufio.NewWriter(&bout)
	logrus.Debugf("New port forward on %s on resource %s using ports %s", hostIP, path, ports)
	portForwarder, err := portforward.NewOnAddresses(dialer, address, ports, stopChan, readyChan, buffOut, buffErr)
	if err != nil {
		return err
	}
	logrus.Debug("Forwarding port")
	err = portForwarder.ForwardPorts()
	if err != nil {
		return err
	}
	return nil
}
