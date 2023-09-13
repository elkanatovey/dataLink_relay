package main

import (
	"net/http"
	"os"
	"time"

	"fmt"
	"mbg-relay/relayconn"
)

func main() {

	go relayconn.StartRelay()

	time.Sleep(100 * time.Millisecond)
	requestURL := fmt.Sprintf("http://localhost:%d/clientconn", relayconn.ServerPort)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	fmt.Println("random number:", relayconn.MaintainConnection())
}
