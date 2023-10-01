package main

import (
	"errors"
	"fmt"
	"mbg-relay/example"
	"mbg-relay/relayconn/relay"
	"mbg-relay/relayconn/utils/logutils"
	"net/http"
)

func StartRelay() { //@todo currently incorrect
	logutils.SetLogStyle()
	r := relay.NewRelay()
	untrustedRelay := http.Server{
		Addr:    fmt.Sprintf("localhost:%d", example.ServerPort),
		Handler: r.Mux,
	}
	if err := untrustedRelay.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
	}
}

func main() {

	//relayAddress := fmt.Sprintf("localhost:%d", ServerPort)
	StartRelay()

	fmt.Println("random number:", relay.MaintainConnection())
}
