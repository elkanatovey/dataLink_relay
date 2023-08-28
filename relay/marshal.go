package relay

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (cr *ConnectionRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Data       string `json:"Data"`
		ImporterID string `json:"ImporterID"`
		ExporterID string `json:"ExporterID"`
	}{
		Data:       cr.Data,
		ImporterID: cr.ImporterID,
		ExporterID: cr.ExporterID,
	})
}

func (cr *ConnectionRequest) UnmarshalJSON(data []byte) error {
	var aux = &struct {
		Data       string `json:"Data"`
		ImporterID string `json:"ImporterID"`
		ExporterID string `json:"ExporterID"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	cr.Data = aux.Data
	cr.ImporterID = aux.ImporterID
	cr.ExporterID = aux.ExporterID

	return nil
}

func MarshalToSSEEvent(connReq ConnectionRequest) (string, error) {
	data, err := json.Marshal(connReq)
	if err != nil {
		return "json marshalling unsuccessful", err
	}

	event := fmt.Sprintf("event: connection\nid: %s\ndata: %s\n\n", connReq.ImporterID, data)
	return event, nil
}

func UnmarshalFromSSEEvent(sseEvent string) (ConnectionRequest, error) {
	var connReq ConnectionRequest

	// Find the start of the JSON data in the SSE event
	dataStart := strings.Index(sseEvent, "\ndata:")
	if dataStart == -1 {
		return connReq, fmt.Errorf("no data field found in SSE event")
	}

	// Extract the JSON data
	jsonData := sseEvent[dataStart+len("\ndata:"):]
	jsonData = strings.TrimSpace(jsonData)

	// Unmarshal the JSON data into the struct
	err := json.Unmarshal([]byte(jsonData), &connReq)
	if err != nil {
		return connReq, err
	}

	return connReq, nil
}
