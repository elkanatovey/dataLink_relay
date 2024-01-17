// Package tcp_endpoints provides Dialer and Listener api's for connecting via a Relay.Relay
package tcp_endpoints

import (
	"context"
	"errors"
	"github.com/elkanatovey/dataLink_relay/pkg/api"
	"io"
	"net"
)

const bufferSize = 100

// RelayListener listens for incoming connections via a relay over a http sse connection.
type RelayListener struct {
	manager       *listenerManager //should this be promoted?
	reqHandlingCh chan struct {
		*api.ConnectionRequest
		error
	}
	reqErrCh      chan error
	ctx           context.Context
	closeListener context.CancelCauseFunc //calling this CancelFunc will close the persistent connection maintained by listen_internal()
}

// Accept return an error if the listener closed. The first error returned is the reason for closing,
// after which calls will fail with net.ErrClosed
func (r RelayListener) Accept() (net.Conn, error) {
	req := <-r.reqHandlingCh // a blocked Accept() call will be released when reqHandlingCh is closed i.e. RelayListener closed

	//handling closed server
	if req.error != nil {
		return nil, req.error
	}
	if req.ConnectionRequest == nil && req.error == nil {
		return nil, net.ErrClosed
	}
	return r.manager.internalTCPCallbackReq(req.ClientID, r.Addr().String())

}

// Close returns an error if it was called after the RelayListener already closed.
// An api.ConnectionRequest that has already buffered may complete even after close is called
func (r RelayListener) Close() error {
	r.closeListener(io.EOF)

	err := <-r.reqErrCh

	//successful close
	if errors.Is(err, context.Canceled) {
		return nil
	}
	//already closed
	if err == nil {
		return errors.New("already closed")
	}
	// anything else
	return err
}

func (r RelayListener) Addr() net.Addr { //what should go here?
	return r.manager.listeningAddress
}

// Listen will only know to listen on an address if the RelayListener was properly initialised with a relay URL
func (r RelayListener) Listen(network, address string) (net.Listener, error) {
	if network != "tcp" {
		panic("we only support tcp")
	}

	err := r.manager.listenInternal(r.ctx, r.reqHandlingCh, r.reqErrCh, address)
	if err != nil {
		return nil, err
	}

	return r, err
}

// NewRelayListener creates a RelayListener that implements the net.Listener api. To run the RelayListener call
// RelayListener.Listen. relayURL is the address of the relay via which we listen
func NewRelayListener(relayURL string) RelayListener {
	ctx, cancel := context.WithCancelCause(context.Background())

	listener := RelayListener{
		manager: newListenerManager(relayURL),
		reqHandlingCh: make(chan struct {
			*api.ConnectionRequest
			error
		}, bufferSize),
		reqErrCh:      make(chan error, 1),
		ctx:           ctx,
		closeListener: cancel,
	}
	return listener
}

// ListenRelay creates and starts a RelayListener without prior initialisation of the backing struct
// address is the address listened on, relayURL is the address of the relay via which we listen,
// currently only tcp is supported
// example usage `tcp_endpoints.listen_relay("tcp", "myserver.com:4444" ,"golang.org:8080")`
func ListenRelay(network, address string, relayURL string) (net.Listener, error) {
	l := NewRelayListener(relayURL)

	return l.Listen(network, address)
}

// ListenerAddress represents the address that a RelayListener is listening on. To connect to a RelayListener the
// ServerID field of an api.ConnectionRequest should be the same as the name field here
type ListenerAddress struct {
	Name string
}

func (r ListenerAddress) Network() string {
	return "tcp"
}

// String is the address that a RelayListener is listening on. It panics of the RelayListener hasn't started listening
func (r ListenerAddress) String() string {
	if r.Name != "" {
		return r.Name
	}
	panic("listener has not started listening")
}
