package cmd

import (
	"net/http"
	"net/url"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// PortForwardOptions contains all the options for running the port-forward cli command.
type PortForwardOptions struct {
	Config       *restclient.Config
	Address      []string
	Ports        []string
	StopChannel  chan struct{}
	ReadyChannel chan struct{}
}

type defaultPortForwarder struct {
	genericclioptions.IOStreams
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts PortForwardOptions) (PortForward, error) {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return PortForward{}, err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforward.NewOnAddresses(dialer, opts.Address, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return PortForward{}, err
	}
	errCn := make(chan error)
	go func() {
		errCn <- fw.ForwardPorts()
	}()
	select {
	case <-errCn:
		return PortForward{}, err
	case <-opts.ReadyChannel: // Required for GetPorts
	}
	ports, err := fw.GetPorts()
	if err != nil {
		return PortForward{}, err
	}
	return PortForward{
		LocalPort: ports[0].Local,
		Status:    errCn,
	}, nil
}

type PortForward struct {
	LocalPort uint16
	Status    chan error
}
