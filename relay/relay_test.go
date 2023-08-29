package relay

import (
	"context"
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

func TestExporterDB(t *testing.T) {
	a := InitExporterDB()
	exp := "aa"
	a.AddExporter(exp, InitExporter(context.TODO()))
	a.RemoveExporter(exp)

	actualSize := len(a.exporters)
	expectedSize := 0
	if actualSize != expectedSize {
		t.Errorf("Expected Size(%d) is not same as"+
			" actual Size (%d)", expectedSize, actualSize)
	}
}

func TestHandleServerLongTermConnection(t *testing.T) {
	mockDB := InitExporterDB()
	handler := HandleServerLongTermConnection(mockDB)

	reqBody := ExporterAnnouncement{ExporterID: "123"}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/long-term-connection", bytes.NewReader(reqBodyBytes))
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
	connReq := ConnectionRequest{
		Data:       "Some data",
		ImporterID: "123",
		ExporterID: "456",
	}
	err = mockDB.NotifyExporter("123", InitImporterData(connReq))
	if err != nil {
		return
	}
	// Wait for a short time to ensure the handler handles the sent message
	time.Sleep(100 * time.Millisecond)
	event, _ := MarshalToSSEEvent(connReq)

	// Check the response body contains the expected SSE event
	if rr.Body.String() != event {
		t.Errorf("response body does not match expected SSE event:\nExpected: %s\nActual: %s", event, rr.Body.String())
	}
}
