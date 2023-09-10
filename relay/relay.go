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

package relay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// Relay contains data for running relay code
type Relay struct {
	activeExporters *ExporterDB
	mux             http.Handler
}

func NewRelay() *Relay {
	exp := InitExporterDB()
	mux := registerHandlers(exp)
	return &Relay{
		activeExporters: exp,
		mux:             mux,
	}
}

const ServerPort = 3333

// StartRelay starts the main relay function.
// Responsibilities: start listener for servers, start listeners for clients
func StartRelay() {
	exportersServed := InitExporterDB()
	mux := registerHandlers(exportersServed)

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

func registerHandlers(exportersServed *ExporterDB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(Listen, HandleServerLongTermConnection(exportersServed)) //listen
	mux.HandleFunc(Dial, HandleClientConnection(exportersServed))           //call
	mux.HandleFunc(Accept, HandleServerCallBackConnection)                  //accept
	return mux
}

// under construction
func HandleServerLongTermConnection(db *ExporterDB) http.HandlerFunc {

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
		db.AddExporter(exporterID, connectionRequests) //@todo define api for server requests

		go func() {
			<-r.Context().Done()
			db.RemoveExporter(exporterID)
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

func HandleClientConnection(db *ExporterDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cr ConnectionRequest
		err := json.NewDecoder(r.Body).Decode(&cr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		imd := InitImporterData(cr)

		err = db.NotifyExporter(cr.ExporterID, imd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var res = <-imd.resultNotificationCh
		if res.Message != NotePassed {
			http.Error(w, res.Error.Error(), http.StatusBadRequest)
			return
		}
		//hijack connection
		conn := hijackConn(w)
		if conn == nil {
			return
		}
		//wait for callback
		//create data that can be connected to
	}
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
