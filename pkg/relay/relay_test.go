package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elkanatovey/dataLink_relay/pkg/api"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleServerLongTermConnection(t *testing.T) {
	mockDB := initRelayData()
	server := httptest.NewServer(HandleServerLongTermConnection(mockDB))
	defer server.Close()

	reqBody := api.ListenRequest{ServerID: "123"}
	reqBodyBytes, _ := json.Marshal(reqBody)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", server.URL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// the server is registered before the handler flushes its 200, so headers are readable here
	if resp.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}
	if contentType := resp.Header.Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/event-stream")
	}
	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "no-cache" {
		t.Errorf("handler returned wrong cache control header: got %v want %v", cacheControl, "no-cache")
	}

	connReq := api.ConnectionRequest{
		Data:     "Some data",
		ClientID: "123",
		ServerID: "456",
	}
	if err := mockDB.NotifyListeningServer("123", InitClientData(connReq)); err != nil {
		t.Fatal(err)
	}

	event, _ := api.MarshalToSSEEvent(&connReq)
	buf := make([]byte, len(event))
	if _, err := io.ReadFull(resp.Body, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != event {
		t.Errorf("response body does not match expected SSE event:\nExpected: %s\nActual: %s", event, string(buf))
	}
}
