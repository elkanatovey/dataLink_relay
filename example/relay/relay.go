package main

import (
	"errors"
	"fmt"
	"github.com/elkanatovey/dataLink_relay/example/utils"
	"github.com/elkanatovey/dataLink_relay/pkg/relay"
	"github.com/elkanatovey/dataLink_relay/pkg/utils/logutils"
	"net/http"
	"time"
)

func StartRelay() {
	logutils.SetLogStyle()
	kp, err := utils.LoadRelayKeyPair()
	if err != nil {
		fmt.Printf("error loading relay key: %s\n", err)
		return
	}
	r := relay.NewRelay(kp)
	untrustedRelay := http.Server{
		Addr:              fmt.Sprintf("localhost:%d", utils.ServerPort),
		Handler:           r.Mux,
		ReadHeaderTimeout: 10 * time.Second,
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
}
