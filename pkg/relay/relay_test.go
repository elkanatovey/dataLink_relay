package relay

import (
	"mbg-relay/pkg/api"

	//"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// test function
func TestMaintainConnection(t *testing.T) {
	actualInt := MaintainConnection()
	expectedInt := 1
	if actualInt != expectedInt {
		t.Errorf("Expected Int(%d) is not same as"+
			" actual string (%d)", expectedInt, actualInt)
	}
}

func TestHandleServerLongTermConnection(t *testing.T) {
	mockDB := initRelayData()
	handler := HandleServerLongTermConnection(mockDB)

	reqBody := api.ListenRequest{ServerID: "123"}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", api.Listen, bytes.NewReader(reqBodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	go handler(rr, req)
	// Wait for a short time for the handler to start up
	time.Sleep(1000 * time.Millisecond)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/event-stream")
	}

	if cacheControl := rr.Header().Get("Cache-Control"); cacheControl != "no-cache" {
		t.Errorf("handler returned wrong cache control header: got %v want %v", cacheControl, "no-cache")
	}

	if connectionControl := rr.Header().Get("Connection"); connectionControl != "keep-alive" {
		t.Errorf("handler returned wrong connection header: got %v want %v", connectionControl, "keep-alive")
	}

	// Simulate SSE messages
	connReq := api.ConnectionRequest{
		Data:     "Some data",
		ClientID: "123",
		ServerID: "456",
	}
	err = mockDB.NotifyListeningServer("123", InitClientData(connReq))
	if err != nil {
		return
	}
	// Wait for a short time to ensure the handler handles the message sent
	time.Sleep(100 * time.Millisecond)
	event, _ := api.MarshalToSSEEvent(&connReq)

	// Check the response body contains the expected SSE event
	if rr.Body.String() != event {
		t.Errorf("response body does not match expected SSE event:\nExpected: %s\nActual: %s", event, rr.Body.String())
	}
}
