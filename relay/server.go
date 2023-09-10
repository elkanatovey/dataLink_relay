package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

// ExportingServer exports services via relay
type ExportingServer struct {
	Connection    *http.Client
	RelayURL      string // address of relay
	ExporterID    string
	maxBufferSize int
}

// NewExportingServer creates a new ExportingServer
func NewExportingServer(url string, id string, opts ...func(c *ExportingServer)) *ExportingServer {
	s := &ExportingServer{
		RelayURL:      url,
		ExporterID:    id,
		Connection:    &http.Client{},
		maxBufferSize: 1 << 16,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *ExportingServer) AcceptConnection(ctx context.Context, cr *ConnectionRequest) (net.Conn, error) {

	//create request

	//run request
	// capture socket and pass back

	return nil, nil
}

// AdvertiseService maintains the persistent connection through which clients send connection requests
func (s *ExportingServer) AdvertiseService(ctx context.Context, handlingCH chan *ConnectionRequest) error {
	resp, err := s.listenRequest(ctx)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not connect to stream: %s", http.StatusText(resp.StatusCode))
	}
	reader := NewEventStreamReader(resp.Body, s.maxBufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil //should handlingCH be closed here?
		default:
			event, err := reader.ReadEvent()
			if err != nil {
				//send off to be handled
				return err
			}
			//sendoff to be handled
			handlingCH <- event
		}

	}

}

// listenRequest opens the connection to the relay
func (s *ExportingServer) listenRequest(ctx context.Context) (*http.Response, error) {

	req, err := s.createListenRequest(ctx)
	if err != nil {
		return nil, err
	}
	return s.Connection.Do(req)
}

func (s *ExportingServer) createListenRequest(ctx context.Context) (*http.Request, error) {
	reqBody := ExporterAnnouncement{ExporterID: s.ExporterID}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", s.RelayURL+Listen, bytes.NewReader(reqBodyBytes)) //@todo url here needs to aslddresss appropriate handler
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
