/**********************************************************/
/* relay : This connects services behind firewalls that can only service outgoing connections.
/**********************************************************/
// Workflow of relay usage
// After relay starts up
//    1) Server wishing to export connects with a call to HandleServerLongTermConnection. This is a long term connection back to the exporter
//    2) Client wishing to import connects via  a call to HandleClientConnection
//    3) Relay initiates a callback via persistent connection to exporter requesting a callback
//    4) ListeningServer calls back, exporter and importer sockets are connected

// Package relay implements the core logic of the relay via which a tcp_endpoints.RelayDialer and tcp_endpoints.RelayListener connect
package relay

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elkanatovey/dataLink_relay/pkg/api"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// callbackTimeout bounds how long a client waits for a listening server to call back
const callbackTimeout = 30 * time.Second

// Relay contains Data for running relay code
type Relay struct {
	Data StateManager
	// Mux serves all three routes over a single plaintext listener. This is the simplest deployment
	// and the historical behaviour.
	Mux http.Handler
	// DataMux serves only the hijacked client dial and server callback routes. Pair it with
	// ControlMux to keep server registration off the plaintext listener entirely.
	DataMux http.Handler
	// ControlMux serves only server registration and requires a verified client certificate, so it
	// must be given to a http.Server configured with TLS and tls.RequireAndVerifyClientCert.
	ControlMux http.Handler
	logger     *logrus.Entry
}

// StateManager represents the relays internal state with respect to all listeningServers and connectingClients
// it is currently serving
type StateManager interface {
	AddListeningServer(expID string, exp *ListeningServer)
	RemoveListeningServer(expID string)
	NotifyListeningServer(expID string, msg *ClientData) error
	AddConnectingClient(impID string, imp *ConnectingClient)
	RemoveConnectingClient(impID string)
	NotifyConnectingClient(impID string, connection *ServerConn) error
}

// RelayData contains dbs of listeningServers advertising services and connectingClients waiting for callback connections
type RelayData struct {
	*listeningServerDB
	*connectingClientDB
	logger      *logrus.Entry
	routingKeys atomic.Pointer[[]*api.RelayKeyPair] // keyring used to open sealed routing metadata
}

func initRelayData() *RelayData {
	return &RelayData{
		listeningServerDB:  initListeningServerDB(),
		connectingClientDB: initConnectingClientDB(),
		logger:             logrus.WithField("component", "relaydata"),
	}
}

// SetRoutingKeys replaces the relay's routing keyring. Safe to call while the relay is serving.
func (d *RelayData) SetRoutingKeys(ring []*api.RelayKeyPair) {
	d.routingKeys.Store(&ring)
}

// maxRoutingBody caps how much of a routing message the relay will buffer. Routing messages are a
// small JSON object, sealed or not, so this is generous while keeping an unauthenticated request
// from exhausting relay memory.
const maxRoutingBody = 64 << 10

// readRoutingBody reads a routing message off the request, refusing bodies larger than
// maxRoutingBody.
func readRoutingBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	return io.ReadAll(http.MaxBytesReader(w, r.Body, maxRoutingBody))
}

// decodeRouting opens a sealed routing message using the relay's keyring, falling back to plaintext
// JSON when no keyring is configured or the body is not sealed.
func (d *RelayData) decodeRouting(body []byte, v any) error {
	if ring := d.routingKeys.Load(); ring != nil && len(*ring) > 0 {
		if err := api.OpenRouting(body, *ring, v); err == nil {
			return nil
		}
	}
	return json.Unmarshal(body, v)
}

// NewRelay returns a Relay with all initialised data structures and handler functions. To start the relay it's mux needs
// to be passed to a http.Server and then start the server
func NewRelay(routingKeys ...*api.RelayKeyPair) *Relay {
	data := initRelayData()
	if len(routingKeys) > 0 {
		data.SetRoutingKeys(routingKeys)
	}
	return &Relay{
		Data:       data,
		Mux:        registerHandlers(data),
		DataMux:    registerDataHandlers(data),
		ControlMux: registerControlHandlers(data),
		logger:     logrus.WithField("component", "relay"),
	}
}

// SetRoutingKeys updates the relay's routing keyring at runtime, without a restart.
func (r *Relay) SetRoutingKeys(ring []*api.RelayKeyPair) {
	if d, ok := r.Data.(*RelayData); ok {
		d.SetRoutingKeys(ring)
	}
}

func registerHandlers(relayState *RelayData) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(api.Listen, HandleServerLongTermConnection(relayState)) //listen
	mux.HandleFunc(api.Dial, HandleClientConnection(relayState))           //call
	mux.HandleFunc(api.Accept, HandleServerCallBackConnection(relayState)) //accept
	return mux
}

// registerDataHandlers serves only the routes that get hijacked into the data tunnel. They carry
// end-to-end encrypted traffic, so wrapping them in TLS would only nest encryption.
func registerDataHandlers(relayState *RelayData) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(api.Dial, HandleClientConnection(relayState))           //call
	mux.HandleFunc(api.Accept, HandleServerCallBackConnection(relayState)) //accept
	return mux
}

// registerControlHandlers serves only server registration, behind a client certificate check.
func registerControlHandlers(relayState *RelayData) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(api.Listen, requireClientCert(HandleServerLongTermConnection(relayState)))
	return mux
}

// requireClientCert rejects requests that did not arrive over TLS with a verified client
// certificate. It guards against a ControlMux accidentally being served on a plaintext listener.
func requireClientCert(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "client certificate required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// authorizeServerID checks that a control connection's client certificate authorises registering
// serverID. On a plaintext connection there is no certificate to check and the legacy unauthenticated
// behaviour is preserved.
func authorizeServerID(r *http.Request, serverID string) error {
	if r.TLS == nil {
		return nil
	}
	if len(r.TLS.PeerCertificates) == 0 {
		// Not reachable behind ControlMux, which already requires a certificate, but a relay serving
		// the combined Mux over TLS without RequireAndVerifyClientCert would land here.
		return errors.New("no client certificate presented")
	}
	if err := r.TLS.PeerCertificates[0].VerifyHostname(serverID); err != nil {
		return fmt.Errorf("certificate does not authorise server id %q: %w", serverID, err)
	}
	return nil
}

// HandleServerLongTermConnection maintains a persistent connection on behalf of an ListeningServer and passes connection
// requests back to the underlying ExportingServer when received
func HandleServerLongTermConnection(relayState *RelayData) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}
		relayState.logger.Infof("server connected to relay..")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		body, err := readRoutingBody(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}
		var req api.ListenRequest
		if err := relayState.decodeRouting(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}

		// Get the exporterID
		exporterID := req.ServerID
		if exporterID == "" {
			http.Error(w, "Please specify an exporter name!", http.StatusInternalServerError)
			relayState.logger.Errorln("exporter name not specified")
			return
		}

		if err := authorizeServerID(r, exporterID); err != nil {
			http.Error(w, "not authorised to register this server id", http.StatusForbidden)
			relayState.logger.Errorln(err)
			return
		}

		relayState.logger.Infof("listening exporter: %s", req.ServerID)
		//allow connectingClients to listenRequest this service
		connectionRequests := InitListeningServer(r.Context())
		relayState.AddListeningServer(exporterID, connectionRequests)

		go func() {
			<-r.Context().Done()
			relayState.RemoveListeningServer(exporterID)
			relayState.logger.Infof(" exporter %s stopped listening", req.ServerID)
			close(connectionRequests.serverNotificationCh)
			for connectionRequest := range connectionRequests.serverNotificationCh {
				connectionRequest.resultNotificationCh <- api.ForwardingSuccessNotification{Message: api.NoteServerConnLost}
			}

		}()

		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for importer := range connectionRequests.serverNotificationCh {

			event, err := api.MarshalToSSEEvent(&importer.msg)
			if err != nil {
				relayState.logger.Errorln(err)
				importer.resultNotificationCh <- api.ForwardingSuccessNotification{Message: api.NoteFail, Error: err}
			}

			_, err = fmt.Fprint(w, event)
			if err != nil {
				relayState.logger.Errorln(err)
				importer.resultNotificationCh <- api.ForwardingSuccessNotification{Message: api.NoteFail, Error: err}
			}

			flusher.Flush()
			importer.resultNotificationCh <- api.ForwardingSuccessNotification{Message: api.NotePassed}
		}

	}
}

// HandleClientConnection passes a ConnectionRequest to a waiting ListeningServer and waits for a socket to be received
// from a callback connection with which to connect an ConnectingClient connection
func HandleClientConnection(relayState *RelayData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := readRoutingBody(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}
		var cr api.ConnectionRequest
		if err := relayState.decodeRouting(body, &cr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}
		relayState.logger.Infof("connection request server: %s, client: %s", cr.ServerID, cr.ClientID)

		// register the waiting client before notifying the server so a fast callback cannot arrive first
		imp := InitConnectingClient(r.Context())
		relayState.AddConnectingClient(getWaitingClientId(cr), imp)
		defer relayState.removeAndDrainConnectingClient(getWaitingClientId(cr), imp)

		imd := InitClientData(cr)
		err = relayState.NotifyListeningServer(cr.ServerID, imd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			relayState.logger.Errorln(err, " notifyListeningServer failed")
			return
		}

		var res = <-imd.resultNotificationCh
		if res.Message != api.NotePassed {
			http.Error(w, res.Error.Error(), http.StatusBadRequest)
			return
		}

		var serverConn *ServerConn
		select {
		case serverConn = <-imp.sockPassCh:
		case <-r.Context().Done():
			relayState.logger.Infof("client %s stopped waiting for server %s", cr.ClientID, cr.ServerID)
			return
		case <-time.After(callbackTimeout):
			http.Error(w, "timed out waiting for server callback", http.StatusGatewayTimeout)
			relayState.logger.Errorf("timed out waiting for server %s callback for client %s", cr.ServerID, cr.ClientID)
			return
		}

		if serverConn.err != nil {
			http.Error(w, serverConn.err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(serverConn.err)
			return
		}

		relayState.logger.Infof("connecting server: %s, client: %s", cr.ServerID, cr.ClientID)
		//hijack connection
		clientConn := hijackConn(w)
		if clientConn == nil {
			relayState.logger.Errorln("server does not support hijacking")
			serverConn.conn.Close()
			return
		}

		err = uniteConnections(clientConn, serverConn.conn)
		if err != nil {
			relayState.logger.Errorln(err, "unite connections quit unexpectedly")
		}
		relayState.logger.Infof("connection stopped server: %s, client: %s", cr.ServerID, cr.ClientID)
		return
	}
}

// HandleServerCallBackConnection manages the server callback upon receiving a client request,
// and passes the connection to the waiting client handler for gluing
func HandleServerCallBackConnection(relayState *RelayData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := readRoutingBody(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}
		var ca api.ConnectionAccept
		if err := relayState.decodeRouting(body, &ca); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}

		if ca.ServerID == "" {
			http.Error(w, "Please specify a server name!", http.StatusInternalServerError)
			relayState.logger.Errorln("server name not specified")
			return
		}
		if ca.ClientID == "" {
			http.Error(w, "Please specify a client name!", http.StatusInternalServerError)
			relayState.logger.Errorln("client name not specified")
			return
		}

		//hijack connection
		conn := hijackConn(w)
		err = nil
		if conn == nil {
			relayState.logger.Errorln("relay does not support hijacking") //@todo should we notify the importer?
			err = errors.New("unsuccessful server connect")
		}

		cn := &ServerConn{conn, err}

		err = relayState.NotifyConnectingClient(getCallingServerId(ca), cn)
		if err != nil {
			relayState.logger.Errorln(err)
			// the waiting client is gone; close the hijacked socket so it is not leaked
			if conn != nil {
				conn.Close()
			}
		}

		return
	}
}
