package relayconn

// Marshaling functions related to ConnectionRequests

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (cr *ConnectionRequest) ToJSON() ([]byte, error) {
	return json.Marshal(cr)
}

func (cr *ConnectionRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, cr)
}

func MarshalToSSEEvent(connReq *ConnectionRequest) (string, error) {
	data, err := json.Marshal(connReq)
	if err != nil {
		return "json marshalling unsuccessful", err
	}

	event := fmt.Sprintf("event: connection\ndata: %s\n\n", data)
	return event, nil
}

func UnmarshalFromSSEEvent(sseEvent string) (*ConnectionRequest, error) {
	var connReq ConnectionRequest

	// Find the start of the JSON data in the SSE event
	dataStart := strings.Index(sseEvent, "\ndata:")
	if dataStart == -1 {
		return nil, fmt.Errorf("no data field found in SSE event")
	}

	// Extract the JSON data
	jsonData := sseEvent[dataStart+len("\ndata:"):]
	jsonData = strings.TrimSpace(jsonData)
	// Unmarshal the JSON data into the struct
	err := json.Unmarshal([]byte(jsonData), &connReq)
	if err != nil {
		return nil, err
	}

	return &connReq, nil
}
