package server

import (
	"context"
	"errors"
	"io"
	"mbg-relay/pkg/api"
	"mbg-relay/pkg/relay"
	"net/http/httptest"
	"testing"
	"time"
)

var relayServer *httptest.Server

var exporterName = "foobar"
var connReq1 = api.ConnectionRequest{
	Data:     "Some data",
	ClientID: "imp1",
	ServerID: exporterName,
}

var connReq2 = api.ConnectionRequest{
	Data:     "Some data",
	ClientID: "imp2",
	ServerID: exporterName,
}

func TestExportingServer_AdvertiseService(t *testing.T) {
	//start relay
	//open sub point
	//send messages

	//start relay
	r := relay.NewRelay()
	relayServer = httptest.NewServer(r.Mux)

	exportingServer := NewExportingServer(relayServer.Listener.Addr().String(), exporterName)

	// channel to receive connrequests
	handlingChennel := make(chan struct {
		*api.ConnectionRequest
		error
	},
		100)
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background()) // need to add  sse events to server to send + spin up gouroutine for export logic

	//advertise
	err := exportingServer.AdvertiseService(ctx, handlingChennel, errChan)
	if err != nil {
		t.Errorf(err.Error())
		t.Errorf("connreq1 fail")
	}
	time.Sleep(1000 * time.Millisecond)

	// notify exporter on relay end
	err = r.Data.NotifyListeningServer(connReq1.ServerID, relay.InitClientData(connReq1))
	if err != nil {
		t.Errorf(err.Error())
		t.Errorf("connreq1 fail")
	}
	err = r.Data.NotifyListeningServer(connReq2.ServerID, relay.InitClientData(connReq2))
	if err != nil {
		t.Errorf(err.Error())
		t.Errorf("connreq2 fail")
	}

	a := <-handlingChennel
	if *a.ConnectionRequest != connReq1 {
		t.Errorf("response body does not match expected SSE event:\nExpected: %s\nActual: %s", connReq1, a)
	}
	b := <-handlingChennel
	if *b.ConnectionRequest != connReq2 {
		t.Errorf("response body does not match expected SSE event:\nExpected: %s\nActual: %s", connReq2, a)
	}

	cancel()

	c := <-handlingChennel
	if !errors.Is(c.error, context.Canceled) {
		t.Errorf(c.Error())
		t.Errorf("should be %s!", io.EOF)
	}
	//var ee error
	err = <-errChan
	if !errors.Is(err, context.Canceled) {
		t.Errorf(err.Error())
		t.Errorf("should be %s!", io.EOF)
	}
}
