package tcp_endpoints

import (
	"context"
	"errors"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/api"
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
	address       RelayAddress
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
	return r.address
}

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

func NewRelayListener(relayURL string, listenerAddress string) RelayListener {
	ctx, cancel := context.WithCancelCause(context.Background())

	listener := RelayListener{
		newListenerManager(relayURL),
		make(chan struct {
			*api.ConnectionRequest
			error
		}, bufferSize),
		make(chan error, 1),
		ctx,
		cancel,
		RelayAddress{listenerAddress},
	}
	return listener
}

func ListenRelay(relayURL string, listenerID string) (net.Listener, error) {
	l := NewRelayListener(relayURL, listenerID)

	err := l.manager.listenInternal(l.ctx, l.reqHandlingCh, l.reqErrCh, listenerID)
	if err != nil {
		return nil, err
	}

	return l, err
}

// RelayAddress represents the address that a RelayListener is listening on. To connect to a RelayListener the
// ServerID field of an api.ConnectionRequest should be the same as the name field here
type RelayAddress struct {
	Name string
}

func (r RelayAddress) Network() string {
	return "tcp"
}

// String is the address that a RelayListener is listening on
func (r RelayAddress) String() string {
	return r.Name
}
