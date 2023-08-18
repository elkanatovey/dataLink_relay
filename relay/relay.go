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
	"errors"
	"fmt"
	"net/http"
)

const serverPort = 3333

// StartRelay starts the main relay function.
// Responsibilities: start listener for servers, start listeners for clients
func StartRelay() {
	mux := http.NewServeMux()
	mux.HandleFunc("/serverconn", HandleServerLongTermConnection)
	mux.HandleFunc("/clientconn", HandleClientConnection)
	mux.HandleFunc("/servercallback", HandleServerCallBackConnection)
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
	}
}

func HandleServerLongTermConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("server: %s /\n", r.Method)
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
