package server

import (
	"context"
	"io"
	"mbg-relay/relayconn/api"
	"mbg-relay/relayconn/relay"
	"net/http/httptest"
	"testing"
	"time"
)

var relayServer *httptest.Server

var exporterName = "foobar"
var connReq1 = api.ConnectionRequest{
	Data:       "Some data",
	ImporterID: "imp1",
	ExporterID: exporterName,
}

var connReq2 = api.ConnectionRequest{
	Data:       "Some data",
	ImporterID: "imp2",
	ExporterID: exporterName,
}

func TestExportingServer_AdvertiseService(t *testing.T) {
	//start relay
	//open sub point
	//send messages

	//start relay
	r := relay.NewRelay()
	relayServer = httptest.NewServer(r.Mux)

	exportingServer := NewExportingServer(relayServer.URL, exporterName)

	// channel to receive connrequests
	handlingChennel := make(chan *api.ConnectionRequest, 100)
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
	err = r.Data.NotifyExporter(connReq1.ExporterID, relay.InitImporterData(connReq1))
	if err != nil {
		t.Errorf(err.Error())
		t.Errorf("connreq1 fail")
	}
	err = r.Data.NotifyExporter(connReq2.ExporterID, relay.InitImporterData(connReq2))
	if err != nil {
		t.Errorf(err.Error())
		t.Errorf("connreq2 fail")
	}

	a := <-handlingChennel
	if *a != connReq1 {
		t.Errorf("response body does not match expected SSE event:\nExpected: %s\nActual: %s", connReq1, a)
	}
	b := <-handlingChennel
	if *b != connReq2 {
		t.Errorf("response body does not match expected SSE event:\nExpected: %s\nActual: %s", connReq2, a)
	}

	cancel()
	//var ee error
	err = <-errChan
	if err != io.EOF {
		t.Errorf(err.Error())
		t.Errorf("should be %s!", io.EOF)
	}
}
