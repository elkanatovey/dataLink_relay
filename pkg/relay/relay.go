/**********************************************************/
/* relay : This connects services behind firewalls that can only service outgoing connections.
/**********************************************************/
// Workflow of relay usage
// After relay starts up
//    1) Server wishing to export connects with a call to HandleServerLongTermConnection. This is a long term connection back to the exporter
//    2) Client wishing to import connects via  a call to HandleClientConnection
//    3) Relay initiates a callback via persistent connection to exporter requesting a callback
//    4) ListeningServer calls back, exporter and importer sockets are connected
//@todo support mtls

package relay

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"mbg-relay/pkg/api"
	"net/http"
)

// Relay contains Data for running relay code
type Relay struct {
	Data   StateManager
	Mux    http.Handler
	logger *logrus.Entry
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
	logger *logrus.Entry
}

func initRelayData() *RelayData {
	return &RelayData{
		initListeningServerDB(),
		initConnectingClientDB(),
		logrus.WithField("component", "relaydata"),
	}
}

// NewRelay returns a Relay with all initialised data structures and handler functions. To start the relay it's mux needs
// to be passed to a http.Server and then start the server
func NewRelay() *Relay {
	data := initRelayData()
	mux := registerHandlers(data)
	return &Relay{
		Data:   data,
		Mux:    mux,
		logger: logrus.WithField("component", "relay"),
	}
}

func registerHandlers(relayState *RelayData) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(api.Listen, HandleServerLongTermConnection(relayState)) //listen
	mux.HandleFunc(api.Dial, HandleClientConnection(relayState))           //call
	mux.HandleFunc(api.Accept, HandleServerCallBackConnection(relayState)) //accept
	return mux
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

		var req api.ListenRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
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
				connectionRequest.resultNotificationCh <- api.ForwardingSuccessNotification{api.NoteServerConnLost, nil}
			}

		}()

		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for importer := range connectionRequests.serverNotificationCh {

			event, err := api.MarshalToSSEEvent(&importer.msg)
			if err != nil {
				relayState.logger.Errorln(err)
				importer.resultNotificationCh <- api.ForwardingSuccessNotification{api.NoteFail, err}
			}

			_, err = fmt.Fprint(w, event)
			if err != nil {
				relayState.logger.Errorln(err)
				importer.resultNotificationCh <- api.ForwardingSuccessNotification{api.NoteFail, err}
			}

			flusher.Flush()
			importer.resultNotificationCh <- api.ForwardingSuccessNotification{api.NotePassed, nil}
		}

	}
}

// HandleClientConnection passes a ConnectionRequest to a waiting ListeningServer and waits for a socket to be received
// from a callback connection with which to connect an ConnectingClient connection
func HandleClientConnection(relayState *RelayData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cr api.ConnectionRequest
		err := json.NewDecoder(r.Body).Decode(&cr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			//relayState.logger.Caller("location", "HandleClientConn")
			relayState.logger.Errorln(err)
			return
		}
		relayState.logger.Infof("connection request server: %s, client: %s", cr.ServerID, cr.ClientID)

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
		imp := InitConnectingClient(r.Context())

		relayState.AddConnectingClient(getWaitingClientId(cr), imp)
		go func() { //@todo the relay in principle allows other server attempts befpre return, should handle?
			<-r.Context().Done()
			relayState.RemoveConnectingClient(getWaitingClientId(cr))
		}()

		serverConn := <-imp.sockPassCh //@todo need to deal with timeout here

		if serverConn.err != nil {
			http.Error(w, serverConn.err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}

		relayState.logger.Infof("connecting server: %s, client: %s", cr.ServerID, cr.ClientID)
		//hijack connection
		clientConn := hijackConn(w)
		if clientConn == nil {
			relayState.logger.Errorln("server does not support hijacking")
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
		var ca api.ConnectionAccept
		err := json.NewDecoder(r.Body).Decode(&ca)
		if err != nil {
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
		}

		return
	}
}

// proxy via which server and client connect

func MaintainConnection() int { return 1 }
