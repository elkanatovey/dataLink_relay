package server

import (
	"context"
	"errors"
	"io"
	"mbg-relay/pkg/api"
	"net"
)

const bufferSize = 100

type RelayListener struct {
	manager       *ExportingServer //should this be promoted?
	reqHandlingCh chan struct {
		*api.ConnectionRequest
		error
	}
	reqErrCh      chan error
	closeListener context.CancelCauseFunc //calling this CancelFunc will close the persistent connection maintained by AdvertiseService()
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
	return r.manager.TCPCallbackReq(req.ClientID)

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
	a := net.Dialer{}
	return a.LocalAddr
}

func Listen(relayURL string, listenerID string) (*RelayListener, error) {
	ctx, cancel := context.WithCancelCause(context.Background())

	listener := &RelayListener{
		NewExportingServer(relayURL, listenerID),
		make(chan struct {
			*api.ConnectionRequest
			error
		}, bufferSize),
		make(chan error, 1),
		cancel,
	}
	err := listener.manager.AdvertiseService(ctx, listener.reqHandlingCh, listener.reqErrCh)
	if err != nil {
		return nil, err
	}

	return listener, err
}
