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
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"net/http"
)

// Relay contains data for running relay code
type Relay struct {
	data   *RelayData
	mux    http.Handler
	logger *logrus.Entry
}

type RelayData struct {
	activeExporters  *ExporterDB
	waitingImporters *ImporterDB
	logger           *logrus.Entry
}

func initRelayData() *RelayData {
	return &RelayData{
		InitExporterDB(),
		InitImporterDB(),
		logrus.WithField("component", "relaydata"),
	}
}

func NewRelay() *Relay {
	data := initRelayData()
	mux := registerHandlers(data)
	return &Relay{
		data:   data,
		mux:    mux,
		logger: logrus.WithField("component", "relay"),
	}
}

const ServerPort = 3333

// StartRelay starts the main relay function.
// Responsibilities: start listener for servers, start listeners for clients
func StartRelay() {
	data := initRelayData()
	mux := registerHandlers(data)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", ServerPort),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
	}
}

func registerHandlers(relayState *RelayData) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(Listen, HandleServerLongTermConnection(relayState)) //listen
	mux.HandleFunc(Dial, HandleClientConnection(relayState))           //call
	mux.HandleFunc(Accept, HandleServerCallBackConnection)             //accept
	return mux
}

// under construction
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
			return
		}

		// Get the exporterID
		exporterID := req.ExporterID
		if exporterID == "" {
			http.Error(w, "Please specify an exporter name!", http.StatusInternalServerError)
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
				fmt.Println(err)
				importer.resultNotificationCh <- ForwardingSuccessNotification{NoteFail, err}
			}

			_, err = fmt.Fprint(w, event)
			if err != nil {
				fmt.Println(err)
				importer.resultNotificationCh <- ForwardingSuccessNotification{NoteFail, err}
			}

			flusher.Flush()
			importer.resultNotificationCh <- ForwardingSuccessNotification{NotePassed, nil}
		}

		fmt.Printf("server: %s /\n", r.Method)

	}
}

// getWaitingID calculates the id of a waiting request based on relevant importer/exporter ids
func getWaitingId(cr ConnectionRequest) string {
	return cr.ImporterID + cr.ExporterID
}

func HandleClientConnection(relayState *RelayData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cr ConnectionRequest
		err := json.NewDecoder(r.Body).Decode(&cr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		imd := InitImporterData(cr)

		err = relayState.activeExporters.NotifyExporter(cr.ExporterID, imd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var res = <-imd.resultNotificationCh
		if res.Message != NotePassed {
			http.Error(w, res.Error.Error(), http.StatusBadRequest)
			return
		}
		imp := InitImporter(r.Context())

		relayState.waitingImporters.AddImporter(getWaitingId(cr), imp)
		go func() { //@todo the relay in principle allows other server attempts befpre return, should handle?
			<-r.Context().Done()
			relayState.waitingImporters.RemoveImporter(getWaitingId(cr))
		}()

		serverConn := <-imp.sockPassCh //@todo need to deal with timeout here

		if serverConn.err != nil {
			http.Error(w, serverConn.err.Error(), http.StatusBadRequest)
			return
		}

		//hijack connection
		clientConn := hijackConn(w)
		if clientConn == nil {
			return
		}

		err = uniteConnections(clientConn, serverConn.conn)
		if err != nil {
			relayState.logger.Errorln(err, "unite connections quit unexpectedly")
		}
		return
		//wait for callback
		//create data that can be connected to
	}
}

func uniteConnections(importerConn net.Conn, exporterConn net.Conn) error {

	var eg errgroup.Group

	eg.Go(func() error {
		defer importerConn.Close()
		defer exporterConn.Close()

		_, err := io.Copy(exporterConn, importerConn)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("exporter->importer: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer importerConn.Close()
		defer exporterConn.Close()

		_, err := io.Copy(importerConn, exporterConn)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("importer->exporter: %w", err)
		}

		return nil
	})

	return eg.Wait()
}

// Hijack the HTTP connection and use the TCP session
func hijackConn(w http.ResponseWriter) net.Conn {
	// Check if we can hijack connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server doesn't support hijacking", http.StatusInternalServerError)
		return nil
	}
	w.WriteHeader(http.StatusOK) //should this be here?
	// Hijack the connection
	conn, _, _ := hj.Hijack()
	return conn
}

func HandleServerCallBackConnection(w http.ResponseWriter, r *http.Request) {
	//get client id
	//get socket from db according to id
	//connect sockets

	fmt.Printf("servercallback: %s /\n", r.Method)
}

// proxy via which server and client connect

func MaintainConnection() int { return 1 }

// func ServerConnect() int { return 1 }

// func ClientConnect() int { return 1 }
