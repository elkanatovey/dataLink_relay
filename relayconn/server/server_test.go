package server

import (
	"context"
	"mbg-relay/relayconn/api"
	"mbg-relay/relayconn/relay"
	"net/http/httptest"
	"testing"
	"time"
)

var relayServer *httptest.Server

func TestExportingServer_AdvertiseService(t *testing.T) {
	//start relay
	//open sub point
	//send messages
	r := relay.NewRelay()
	relayServer = httptest.NewServer(r.Mux)

	exportingServer := NewExportingServer(relayServer.URL, "foobar")

	handlingChennel := make(chan *api.ConnectionRequest)
	ctx, cancel := context.WithCancel(context.Background()) // need to add  sse events to server to send + spin up gouroutine for export logic

	errChan := exportingServer.AdvertiseService(ctx, handlingChennel)
	time.Sleep(1000 * time.Millisecond)
	cancel()
	//var ee error
	err := <-errChan
	if err != nil {
		t.Errorf(err.Error())
		t.Errorf("should be nil!")
	}
}
