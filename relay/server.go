package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// server exports services via relay
type ExportingServer struct {
	Connection *http.Client
	URL        string
	ExporterID string
}

func (s *ExportingServer) AdvertiseService() error {
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
