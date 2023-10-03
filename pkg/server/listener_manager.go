package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/api"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/utils/httputils"
	"net"
	"net/http"
)

// listenerManager contains internal implementation details of a RelayListener's api
type listenerManager struct {
	Connection    *http.Client
	RelayIPPort   string // ip of relay + port for example: 127.0.0.1:39887
	ServerID      string
	maxBufferSize int
	logger        *logrus.Entry
}

// newListenerManager creates a new listenerManager
func newListenerManager(relayAddr string, id string) *listenerManager {
	s := &listenerManager{
		RelayIPPort:   relayAddr,
		ServerID:      id,
		Connection:    &http.Client{},
		maxBufferSize: 1 << 16,
		logger:        logrus.WithField("component", "exportingserver"),
	}

	return s
}

// listenInternal maintains the persistent connection through which clients send connection requests,
// errors are propagated through the passed in channel, canceling the context will close both passed in channels
func (s *listenerManager) listenInternal(ctx context.Context, handlingCH chan struct {
	*api.ConnectionRequest
	error
},
	errCH chan error) error {

	resp, err := s.listenRequest(ctx)
	if err != nil {
		s.logger.Errorln(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("could not connect to stream: %s", http.StatusText(resp.StatusCode))
		s.logger.Errorln(err)
		resp.Body.Close()
		return err
	}

	go func() {
		defer resp.Body.Close()
		defer close(handlingCH)
		defer close(errCH)
		reader := newEventStreamReader(resp.Body, s.maxBufferSize)

		for {
			select { //@todo this case may be able to be removed
			case <-ctx.Done():
				errCH <- context.Canceled //should handlingCH be closed here?
				return
			default:
				event, err := reader.readEvent()
				if err != nil { // should we check here to make sure the connection is with the correct exporter?

					// notify reason for closing is from our end
					if errors.Is(err, context.Canceled) {
						errCH <- context.Canceled

					}

					//send off to be handled
					handlingCH <- struct {
						*api.ConnectionRequest
						error
					}{nil, err}
					return
				}

				//sendoff to be handled
				s.logger.Infof("received connection request from: %s", event.ClientID)
				handlingCH <- struct {
					*api.ConnectionRequest
					error
				}{event, nil}
			}
		}
	}()

	return nil
}

// listenRequest opens the connection to the relay after building the connection request
func (s *listenerManager) listenRequest(ctx context.Context) (*http.Response, error) {

	req, err := s.createListenRequest(ctx)
	if err != nil {
		return nil, err
	}
	return s.Connection.Do(req)
}

// createListenRequest builds the request to open the listen connection for the server
func (s *listenerManager) createListenRequest(ctx context.Context) (*http.Request, error) {
	reqBody := api.ListenRequest{ServerID: s.ServerID}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", api.TCP+s.RelayIPPort+api.Listen, bytes.NewReader(reqBodyBytes)) //@todo should we cancel context in case of error?
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")
	return req, nil
}

// internalTCPCallbackReq calls back a client via the relay at the given ip
func (s *listenerManager) internalTCPCallbackReq(importerName string) (net.Conn, error) {
	s.logger.Infof("Starting TCP callback to importer id %v via relay ip %v", importerName, s.RelayIPPort)
	url := api.TCP + s.RelayIPPort + api.Accept

	jsonData, err := json.Marshal(api.ConnectionAccept{ClientID: importerName, ServerID: s.ServerID})
	if err != nil {
		s.logger.Errorln(err)
		return nil, err
	}

	conn, resp := httputils.Connect(s.RelayIPPort, url, string(jsonData))
	if resp == nil {
		s.logger.Infof("Successfully Connected")
		return conn, nil
	}

	s.logger.Errorf("callback Request Failed")
	return nil, fmt.Errorf("callback Request Failed")
}
