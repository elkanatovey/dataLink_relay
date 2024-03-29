package main

import (
	"errors"
	"fmt"
	"github.com/elkanatovey/dataLink_relay/example/utils"
	"github.com/elkanatovey/dataLink_relay/pkg/relay"
	"github.com/elkanatovey/dataLink_relay/pkg/utils/logutils"
	"net/http"
)

func StartRelay() { //@todo currently incorrect
	logutils.SetLogStyle()
	r := relay.NewRelay()
	untrustedRelay := http.Server{
		Addr:    fmt.Sprintf("localhost:%d", utils.ServerPort),
		Handler: r.Mux,
	}
	if err := untrustedRelay.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http tcp_endpoints: %s\n", err)
		}
	}
}

func main() {

	//relayAddress := fmt.Sprintf("localhost:%d", ServerPort)
	StartRelay()

	fmt.Println("random number:", relay.MaintainConnection())
}
