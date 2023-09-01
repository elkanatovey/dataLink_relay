package relay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// server exports services via relay
type ExportingServer struct {
	Connection    *http.Client
	URL           string // address of relay
	ExporterID    string
	maxBufferSize int
}

func (s *ExportingServer) AdvertiseService(ctx context.Context) error {

	resp, err := s.request(ctx)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not connect to stream: %s", http.StatusText(resp.StatusCode))
	}
	reader := NewEventStreamReader(resp.Body, s.maxBufferSize)
	for {
		event, err := reader.ReadEvent()
		unmarshaled, err := UnmarshalFromSSEEvent(string(event[:]))
		//sendoff to be handled
	}

	return nil
}

// request opens the connection to the relay
func (s *ExportingServer) request(ctx context.Context) (*http.Response, error) {

	req, err := s.createRequest(ctx)
	if err != nil {
		return nil, err
	}
	return s.Connection.Do(req)
}

func (s *ExportingServer) createRequest(ctx context.Context) (*http.Request, error) {
	reqBody := ExporterAnnouncement{ExporterID: s.ExporterID}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", s.URL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")
	return req, nil
}

func (s *ExportingServer) runServer() error {

	return nil
}
