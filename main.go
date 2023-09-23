package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"fmt"
	"mbg-relay/relayconn"
)

const ServerPort = 3333

// StartRelay starts the main relay function.
// Responsibilities: start listener for servers, start listeners for clients
func StartRelay() { //@todo currently incorrect
	r := relayconn.NewRelay()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", ServerPort),
		Handler: r.Mux,
	}
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
	}
}

func main() {

	go StartRelay()

	time.Sleep(100 * time.Millisecond)
	requestURL := fmt.Sprintf("http://localhost:%d%s", ServerPort, relayconn.Dial)
	cr := relayconn.ConnectionRequest{"a", "123", "456"}
	reqBodyBytes, _ := json.Marshal(cr)
	req, err := http.NewRequest("GET", requestURL, bytes.NewReader(reqBodyBytes))
	client := &http.Client{}
	response, err := client.Do(req)

	//res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", response.StatusCode)

	fmt.Println("random number:", relayconn.MaintainConnection())
}
