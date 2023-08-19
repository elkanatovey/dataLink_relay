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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const ServerPort = 3333

// StartRelay starts the main relay function.
// Responsibilities: start listener for servers, start listeners for clients
func StartRelay() {
	mux := http.NewServeMux()
	mux.HandleFunc("/serverconn", HandleServerLongTermConnection)     //listen
	mux.HandleFunc("/clientconn", HandleClientConnection)             //call
	mux.HandleFunc("/servercallback", HandleServerCallBackConnection) //accept

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

// under construction
func HandleServerLongTermConnection(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	fmt.Println("server connected to relay...")

	w.Header().Set("Content-Type", "text/event-stream")

	conReqCh := make(chan string)

	// @todo put channel + r.Context in db that client conns can access

	for connectionRequest := range conReqCh {
		event, err := formatServerSentEvent("connection-request", connectionRequest)
		if err != nil {
			fmt.Println(err)
			break
		}

		_, err = fmt.Fprint(w, event)
		if err != nil {
			fmt.Println(err)
			break
		}

		flusher.Flush()
	}

	fmt.Printf("server: %s /\n", r.Method)
}

// formatServerSentEvent takes name of an event and any kind of data and transforms
// into a server sent event payload structure.
// Data is sent as a json object, { "data": <your_data> }.
//
// Example:
//
//	Input:
//		event="connection-request"
//		data=servicefoo
//	Output:
//		event: connection-request\n
//		data: "{\"data\":servicefoo}"\n\n
func formatServerSentEvent(event string, data any) (string, error) {
	m := map[string]any{
		"data": data,
	}

	buff := bytes.NewBuffer([]byte{})

	encoder := json.NewEncoder(buff)

	err := encoder.Encode(m)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("event: %s\n", event))
	sb.WriteString(fmt.Sprintf("data: %v\n\n", buff.String()))

	return sb.String(), nil
}

func HandleClientConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("client: %s /\n", r.Method)
}

func HandleServerCallBackConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("servercallback: %s /\n", r.Method)
}

// proxy via which server and client connect

func MaintainConnection() int { return 1 }

// func ServerConnect() int { return 1 }

// func ClientConnect() int { return 1 }
