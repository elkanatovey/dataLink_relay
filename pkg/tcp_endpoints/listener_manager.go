package tcp_endpoints

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/elkanatovey/dataLink_relay/pkg/api"
	"github.com/elkanatovey/dataLink_relay/pkg/utils/httputils"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

// listenerManager contains internal implementation details of a RelayListener's api
type listenerManager struct {
	Connection       *http.Client
	RelayIPPort      string // ip of relay + port for example: 127.0.0.1:39887
	ServerID         string
	maxBufferSize    int
	listeningAddress ListenerAddress
	relayPub         *[32]byte
	controlAddr      string // relay mTLS control endpoint, empty when registration is plaintext
	controlTLS       *tls.Config
	logger           *logrus.Entry
}

// newListenerManager creates a new listenerManager
func newListenerManager(relayAddr string, o options) *listenerManager {
	s := &listenerManager{
		RelayIPPort:   relayAddr,
		Connection:    &http.Client{},
		maxBufferSize: 1 << 16,
		relayPub:      o.relayPub,
		logger:        logrus.WithField("component", "exportingserver"),
	}

	if o.controlTLS != nil {
		s.controlAddr = o.controlAddr
		s.controlTLS = o.controlTLS
		s.Connection = &http.Client{Transport: &http.Transport{
			TLSClientConfig: o.controlTLS,
			// The registration connection is a long lived event stream that sets hop-by-hop
			// headers, which the HTTP/2 transport rejects. Setting TLSClientConfig already makes
			// net/http skip HTTP/2, but pin it explicitly so enabling ForceAttemptHTTP2 later
			// cannot silently break registration.
			TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
		}}
	}

	return s
}

// controlEndpoint returns the scheme and address the registration connection is made to. It is the
// relay's mTLS control endpoint when one is configured, and the plaintext relay otherwise.
func (s *listenerManager) controlEndpoint() (string, string) {
	if s.controlTLS != nil {
		return api.MTLS, s.controlAddr
	}
	return api.TCP, s.RelayIPPort
}

// listenInternal maintains the persistent connection through which clients send connection requests,
// the listening address of the server is only set once it calls listenInternal
// errors are propagated through the passed in channel, canceling the context will close both passed in channels
func (s *listenerManager) listenInternal(ctx context.Context, handlingCH chan connRequestResult,
	errCH chan error, address string) error {
	s.listeningAddress = ListenerAddress{address}

	resp, err := s.listenRequest(ctx, address)
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
					handlingCH <- connRequestResult{nil, err}
					return
				}

				//sendoff to be handled
				s.logger.Infof("received connection request from: %s", event.ClientID)
				handlingCH <- connRequestResult{event, nil}
			}
		}
	}()

	return nil
}

// listenRequest opens the connection to the relay after building the connection request
func (s *listenerManager) listenRequest(ctx context.Context, address string) (*http.Response, error) {

	req, err := s.createListenRequest(ctx, address)
	if err != nil {
		return nil, err
	}
	return s.Connection.Do(req)
}

// createListenRequest builds the request to open the listen connection for the server
func (s *listenerManager) createListenRequest(ctx context.Context, address string) (*http.Request, error) {
	reqBodyBytes, err := api.EncodeRouting(api.ListenRequest{ServerID: address}, s.relayPub)
	if err != nil {
		return nil, err
	}

	scheme, addr := s.controlEndpoint()

	req, err := http.NewRequest("POST", scheme+addr+api.Listen, bytes.NewReader(reqBodyBytes)) //@todo should we cancel context in case of error?
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
func (s *listenerManager) internalTCPCallbackReq(importerName string, address string) (net.Conn, error) {
	s.logger.Infof("Starting TCP callback to importer id %v via relay ip %v", importerName, s.RelayIPPort)
	url := api.TCP + s.RelayIPPort + api.Accept

	jsonData, err := api.EncodeRouting(api.ConnectionAccept{ClientID: importerName, ServerID: address}, s.relayPub)
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
