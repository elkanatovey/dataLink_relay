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
	Connection *http.Client
	URL        string // address of relay
	ExporterID string
}

func (s *ExportingServer) AdvertiseService(ctx context.Context) error {

	req, err := s.request(ctx)
	if err != nil {
		return err
	}
	defer req.Body.Close()

	switch req.StatusCode {
	case http.StatusOK:
		// we do not support BOM in sse streams, or \r line separators.
		r := bufio.NewReader(req.Body)
		for {
			event, err := s.parseEvent(r)
			if err != nil {
				return err
			}

			if err := handle(event); err != nil {
				return err
			}
		}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		connectionhandel
	default:
		// trigger a reconnect and output an error.
		return fmt.Errorf("bad response status code %d", req.StatusCode)
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
