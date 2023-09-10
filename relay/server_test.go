package relay

import (
	"context"
	"net/http/httptest"
	"testing"
)

var relayServer *httptest.Server

func TestExportingServer_AdvertiseService(t *testing.T) {
	//start relay
	//open sub point
	//send messages
	r := NewRelay()
	relayServer = httptest.NewServer(r.mux)

	exportingServer := NewExportingServer(relayServer.URL, "foobar")

	handlingChennel := make(chan *ConnectionRequest)
	ctx, _ := context.WithCancel(context.Background()) // need to add  sse events to server to send + spin up gouroutine for export logic

	err := exportingServer.AdvertiseService(ctx, handlingChennel)
	if err != nil {
		t.Errorf("should be nil!")
	}
}
