/**********************************************************/
/* relay : This connects services behind firewalls that can only service outgoing connections.
/**********************************************************/
// Workflow of relay usage
// After relay starts up
//    1) Server wishing to export connects with a call to HandleServerLongTermConnection. This is a long term connection back to the exporter
//    2) Client wishing to import connects via  a call to HandleClientConnection
//    3) Relay initiates a callback via persistent connection to exporter requesting a callback
//    4) Exporter calls back, exporter and importer sockets are connected
//@todo support mtls

package relayconn

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Relay contains Data for running relay code
type Relay struct {
	Data   *RelayData
	Mux    http.Handler
	logger *logrus.Entry
}

// RelayData contains dbs of exporters advertising services and importers waiting for callback connections
type RelayData struct {
	activeExporters  *ExporterDB
	waitingImporters *ImporterDB
	logger           *logrus.Entry
}

func initRelayData() *RelayData {
	setLogStyle()
	return &RelayData{
		InitExporterDB(),
		InitImporterDB(),
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
	mux.HandleFunc(Listen, HandleServerLongTermConnection(relayState)) //listen
	mux.HandleFunc(Dial, HandleClientConnection(relayState))           //call
	mux.HandleFunc(Accept, HandleServerCallBackConnection(relayState)) //accept
	return mux
}

// HandleServerLongTermConnection maintains a persistent connection on behalf of an Exporter and passes connection
// requests back to the underlying ExportingServer when received
func HandleServerLongTermConnection(relayState *RelayData) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		fmt.Println("server connected to relay...")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		var req ExporterAnnouncement
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}

		// Get the exporterID
		exporterID := req.ExporterID
		if exporterID == "" {
			http.Error(w, "Please specify an exporter name!", http.StatusInternalServerError)
			relayState.logger.Errorln("exporter name not specified")
			return
		}

		//allow importers to listenRequest this service
		connectionRequests := InitExporter(r.Context())
		relayState.activeExporters.AddExporter(exporterID, connectionRequests)

		go func() {
			<-r.Context().Done()
			relayState.activeExporters.RemoveExporter(exporterID)
			for connectionRequest := range connectionRequests.exporterNotificationCh {
				connectionRequest.resultNotificationCh <- ForwardingSuccessNotification{NoteServerConnLost, nil}
			}

		}()

		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for importer := range connectionRequests.exporterNotificationCh {
			event, err := MarshalToSSEEvent(&importer.msg)
			if err != nil {
				relayState.logger.Errorln(err)
				importer.resultNotificationCh <- ForwardingSuccessNotification{NoteFail, err}
			}

			_, err = fmt.Fprint(w, event)
			if err != nil {
				relayState.logger.Errorln(err)
				importer.resultNotificationCh <- ForwardingSuccessNotification{NoteFail, err}
			}

			flusher.Flush()
			importer.resultNotificationCh <- ForwardingSuccessNotification{NotePassed, nil}
		}

		fmt.Printf("server: %s /\n", r.Method)

	}
}

// HandleClientConnection passes a ConnectionRequest to a waiting Exporter and waits for a socket to be received
// from a callback connection with which to connect an Importer connection
func HandleClientConnection(relayState *RelayData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cr ConnectionRequest
		err := json.NewDecoder(r.Body).Decode(&cr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			//relayState.logger.Caller("location", "HandleClientConn")
			relayState.logger.Errorln(err)
			return
		}

		imd := InitImporterData(cr)

		err = relayState.activeExporters.NotifyExporter(cr.ExporterID, imd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			relayState.logger.Errorln(err, " notifyexporter failed")
			return
		}

		var res = <-imd.resultNotificationCh
		if res.Message != NotePassed {
			http.Error(w, res.Error.Error(), http.StatusBadRequest)
			return
		}
		imp := InitImporter(r.Context())

		relayState.waitingImporters.AddImporter(getWaitingImporterId(cr), imp)
		go func() { //@todo the relay in principle allows other server attempts befpre return, should handle?
			<-r.Context().Done()
			relayState.waitingImporters.RemoveImporter(getWaitingImporterId(cr))
		}()

		serverConn := <-imp.sockPassCh //@todo need to deal with timeout here

		if serverConn.err != nil {
			http.Error(w, serverConn.err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}

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
		return
	}
}

// HandleServerCallBackConnection manages the server callback upon receiving a client request,
// and passes the connection to the waiting client handler for gluing
func HandleServerCallBackConnection(relayState *RelayData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ca ConnectionAccept
		err := json.NewDecoder(r.Body).Decode(&ca)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			relayState.logger.Errorln(err)
			return
		}

		if ca.ExporterID == "" {
			http.Error(w, "Please specify an exporter name!", http.StatusInternalServerError)
			relayState.logger.Errorln("exporter name not specified")
			return
		}
		if ca.ImporterID == "" {
			http.Error(w, "Please specify an importer name!", http.StatusInternalServerError)
			relayState.logger.Errorln("importer name not specified")
			return
		}

		//hijack connection
		conn := hijackConn(w)
		err = nil
		if conn == nil {
			relayState.logger.Errorln("server does not support hijacking") //@todo should we notify the importer?
			err = errors.New("unsuccesful server connect")
		}

		cn := &ServerConn{conn, err}

		err = relayState.waitingImporters.NotifyImporter(getCallingExporterId(ca), cn)
		if err != nil {
			relayState.logger.Errorln(err)
		}

		return
	}
	//get client id
	//get socket from db according to id
	//connect sockets

	//fmt.Printf("servercallback: %s /\n", r.Method)
}

// proxy via which server and client connect

func MaintainConnection() int { return 1 }

// func ServerConnect() int { return 1 }

// func ClientConnect() int { return 1 }
