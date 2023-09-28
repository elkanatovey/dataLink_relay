package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"mbg-relay/relayconn/api"
	"mbg-relay/relayconn/relay"
	"net/http"
	"os"
	"time"

	"fmt"
)

const ServerPort = 3333

// StartRelay starts the main relay function.
// Responsibilities: start listener for servers, start listeners for clients
func StartRelay() { //@todo currently incorrect
	r := relay.NewRelay()
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
	requestURL := fmt.Sprintf("%slocalhost:%d%s", api.TCP, ServerPort, api.Dial)
	cr := api.ConnectionRequest{ImporterID: "123", ExporterID: "456"}
	reqBodyBytes, _ := json.Marshal(cr)
	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(reqBodyBytes))
	client := &http.Client{}
	response, err := client.Do(req)

	//res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", response.StatusCode)

	fmt.Println("random number:", relay.MaintainConnection())
}
