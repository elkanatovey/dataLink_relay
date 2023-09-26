package relayconn

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

// AdvertiseService maintains the persistent connection through which clients send connection requests,
// errors are propagated through the returned channel
func (s *ExportingServer) AdvertiseService(ctx context.Context, handlingCH chan *ConnectionRequest) <-chan error {
	res := make(chan error)

	go func() {
		resp, err := s.listenRequest(ctx)
		if err != nil {
			res <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			res <- fmt.Errorf("could not connect to stream: %s", http.StatusText(resp.StatusCode))
			return
		}
		reader := NewEventStreamReader(resp.Body, s.maxBufferSize)
		for {
			select {
			case <-ctx.Done():
				res <- nil //should handlingCH be closed here?
				return
			default:
				event, err := reader.ReadEvent()
				if err != nil {
					//send off to be handled
					res <- err
					return
				}
				//sendoff to be handled
				handlingCH <- event
			}

		}
	}()

	return res
}

// listenRequest opens the connection to the relay after building the connection request
func (s *ExportingServer) listenRequest(ctx context.Context) (*http.Response, error) {

	req, err := s.createListenRequest(ctx)
	if err != nil {
		return nil, err
	}
	return s.Connection.Do(req)
}

// createListenRequest builds the request to open the listen connection for the server
func (s *ExportingServer) createListenRequest(ctx context.Context) (*http.Request, error) {
	reqBody := ExporterAnnouncement{ExporterID: s.ExporterID}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", s.RelayURL+Listen, bytes.NewReader(reqBodyBytes)) //@todo should we cancel context in case of error?
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
