package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"mbg-relay/relayconn/api"
	"mbg-relay/relayconn/utils/httputils"
	"net"
	"net/http"
)

// ExportingServer exports services via relay
type ExportingServer struct {
	Connection    *http.Client
	RelayURL      string // address of relay + port
	ExporterID    string
	maxBufferSize int
	logger        *logrus.Entry
}

// NewExportingServer creates a new ExportingServer
func NewExportingServer(url string, id string, opts ...func(c *ExportingServer)) *ExportingServer {
	s := &ExportingServer{
		RelayURL:      url,
		ExporterID:    id,
		Connection:    &http.Client{},
		maxBufferSize: 1 << 16,
		logger:        logrus.WithField("component", "exportingserver"),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *ExportingServer) AcceptConnection(ctx context.Context, cr *api.ConnectionRequest) (net.Conn, error) {

	//create request

	//run request
	// capture socket and pass back

	return nil, nil
}

// AdvertiseService maintains the persistent connection through which clients send connection requests,
// errors are propagated through the returned channel
func (s *ExportingServer) AdvertiseService(ctx context.Context, handlingCH chan *api.ConnectionRequest) <-chan error {
	res := make(chan error, 1)

	go func() {
		resp, err := s.listenRequest(ctx)
		if err != nil {
			s.logger.Errorln(err)
			res <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("could not connect to stream: %s", http.StatusText(resp.StatusCode))
			s.logger.Errorln(err)
			res <- err
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
				if err != nil { // should we check here to make sure the connection is with the correct exporter?

					// make sure that it wasn't closed on our end before logging
					if err != io.EOF {
						s.logger.Errorln(err)
					}
					//send off to be handled
					res <- err
					return
				}
				//sendoff to be handled
				s.logger.Infof("received connection request from: %s", event.ImporterID)
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
	reqBody := api.ExporterAnnouncement{ExporterID: s.ExporterID}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", s.RelayURL+api.Listen, bytes.NewReader(reqBodyBytes)) //@todo should we cancel context in case of error?
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")
	return req, nil
}

// TCPCallbackReq calls back an importer via the relay at the given ip
func (s *ExportingServer) TCPCallbackReq(importerName string) (net.Conn, error) {
	s.logger.Infof("Starting TCP callback to importer id %v via relay ip %v", importerName, s.RelayURL)
	url := api.TCP + s.RelayURL + api.Accept

	jsonData, err := json.Marshal(api.ConnectionAccept{ImporterID: importerName, ExporterID: s.ExporterID})
	if err != nil {
		s.logger.Errorln(err)
		return nil, err
	}

	conn, resp := httputils.Connect(api.TCP+s.RelayURL, url, string(jsonData))
	if resp == nil {
		s.logger.Infof("Successfully Connected")
		return conn, nil
	}

	s.logger.Errorf("callback Request Failed")
	return nil, fmt.Errorf("callback Request Failed")
}

func (s *ExportingServer) runServer() error {

	return nil
}
