package api

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

	event := fmt.Sprintf("event: connection\nData: %s\n\n", data)
	return event, nil
}

func UnmarshalFromSSEEvent(sseEvent string) (*ConnectionRequest, error) {
	var connReq ConnectionRequest

	// Find the start of the JSON Data in the SSE event
	dataStart := strings.Index(sseEvent, "\nData:")
	if dataStart == -1 {
		return nil, fmt.Errorf("no Data field found in SSE event")
	}

	// Extract the JSON Data
	jsonData := sseEvent[dataStart+len("\nData:"):]
	jsonData = strings.TrimSpace(jsonData)
	// Unmarshal the JSON Data into the struct
	err := json.Unmarshal([]byte(jsonData), &connReq)
	if err != nil {
		return nil, err
	}

	return &connReq, nil
}
