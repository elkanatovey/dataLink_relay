package main

import (
	//"bytes"
	//"context"
	//"encoding/json"
	"errors"
	"time"

	//"mbg-relay/relayconn/api"
	//client2 "mbg-relay/relayconn/client"
	"mbg-relay/relayconn/relay"
	"mbg-relay/relayconn/server"
	//"net"
	"net/http"
	//"os"
	//"time"

	"fmt"
)

const serverPort = 3333
const exporterName = "foo"
const importerName = "bar"

// StartRelay starts the main relay function.
// Responsibilities: start listener for servers, start listeners for clients
func StartRelay() { //@todo currently incorrect
	r := relay.NewRelay()
	untrustedRelay := http.Server{
		Addr:    fmt.Sprintf("localhost:%d", serverPort),
		Handler: r.Mux,
	}
	if err := untrustedRelay.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
	}
}

func main() {
	relayAddress := fmt.Sprintf("localhost:%d", serverPort)
	go StartRelay()

	listener, err := server.Listen(relayAddress, exporterName)
	if err != nil {
		return
	}

	time.Sleep(1000 * time.Millisecond)
	listener.Close()

	fmt.Println("random number:", relay.MaintainConnection())
}
