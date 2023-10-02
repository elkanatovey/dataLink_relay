package api

import (
	"testing"
)

func TestMarshalAndUnmarshal(t *testing.T) {
	connReq := ConnectionRequest{
		Data:     "Some Data",
		ClientID: "123",
		ServerID: "456",
	}

	// Test ToJSON
	marshaledData, err := connReq.ToJSON()
	if err != nil {
		t.Errorf("Error marshaling: %v", err)
	}

	// Test FromJSON
	var unmarshaledConnReq ConnectionRequest
	err = unmarshaledConnReq.FromJSON(marshaledData)
	if err != nil {
		t.Errorf("Error unmarshaling: %v", err)
	}

	// Compare original and unmarshaled structs
	if connReq.Data != unmarshaledConnReq.Data ||
		connReq.ClientID != unmarshaledConnReq.ClientID ||
		connReq.ServerID != unmarshaledConnReq.ServerID {
		t.Errorf("Original and unmarshaled structs are not equal")
	}
}

func TestMarshalToSSEEvent(t *testing.T) {
	connReq := ConnectionRequest{
		Data:     "Some Data",
		ClientID: "123",
		ServerID: "456",
	}

	// Test MarshalToSSEEvent
	sseEvent, err := MarshalToSSEEvent(&connReq)
	if err != nil {
		t.Errorf("Error marshaling to SSE event: %v", err)
	}

	expectedSSEEvent := "event: connection\nData: {\"Data\":\"Some Data\",\"ClientID\":\"123\",\"ServerID\":\"456\"}\n\n"
	if sseEvent != expectedSSEEvent {
		t.Errorf("Unexpected SSE event string:\nExpected: %s\nActual:   %s", expectedSSEEvent, sseEvent)
	}
}

func TestUnmarshalFromSSEEvent(t *testing.T) {
	sseEvent := "event: connection\nData: {\"Data\":\"Some Data\",\"ClientID\":\"123\",\"ServerID\":\"456\"}\n\n"

	// Test UnmarshalFromSSEEvent
	connReq, err := UnmarshalFromSSEEvent(sseEvent)
	if err != nil {
		t.Errorf("Error unmarshaling from SSE event: %v", err)
	}

	expectedConnReq := &ConnectionRequest{
		Data:     "Some Data",
		ClientID: "123",
		ServerID: "456",
	}
	if *connReq != *expectedConnReq {
		t.Errorf("Unexpected ConnectionRequest:\nExpected: %+v\nActual:   %+v", expectedConnReq, connReq)
	}
}

func TestUnmarshalFromSSEEvent_Error(t *testing.T) {
	// Test invalid SSE event format
	sseEvent := "event: connection\nid: 123\n"

	_, err := UnmarshalFromSSEEvent(sseEvent)
	if err == nil {
		t.Error("Expected error for invalid SSE event format, but got none")
	}
}
