package server

import (
	"context"
	"io"
	"mbg-relay/pkg/api"
	"net"
)

const bufferSize = 100

type RelayListener struct {
	manager       *ExportingServer //should this be promoted?
	reqHandlingCh chan *api.ConnectionRequest
	reqErrCh      chan error
	closeListener context.CancelFunc //calling this CancelFunc will close the persistent connection maintained by AdvertiseService()
}

func (r RelayListener) Accept() (net.Conn, error) {
	req := <-r.reqHandlingCh // a blocked Accept() call will be released when reqHandlingCh is closed i.e. RelayListener closed

	//handling closed server
	if req == nil {
		return nil, io.EOF
	}
	return r.manager.TCPCallbackReq(req.ClientID)

}

func (r RelayListener) Close() error {
	r.closeListener()

	err := <-r.reqErrCh
	return err
}

func (r RelayListener) Addr() net.Addr { //what should go here?
	a := net.Dialer{}
	return a.LocalAddr
}

func Listen(relayURL string, listenerID string) (*RelayListener, error) {
	ctx, cancel := context.WithCancel(context.Background())

	listener := &RelayListener{
		NewExportingServer(relayURL, listenerID),
		make(chan *api.ConnectionRequest, bufferSize),
		make(chan error, 1),
		cancel,
	}
	err := listener.manager.AdvertiseService(ctx, listener.reqHandlingCh, listener.reqErrCh)
	if err != nil {
		return nil, err
	}

	return listener, err
}
