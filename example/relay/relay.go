package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/elkanatovey/dataLink_relay/example/utils"
	"github.com/elkanatovey/dataLink_relay/pkg/relay"
	"github.com/elkanatovey/dataLink_relay/pkg/utils/logutils"
	"net/http"
	"time"
)

// requireControlTLS makes server registration reachable only over the mTLS control listener. With it
// unset the plaintext listener also serves registration, which keeps the plain TCP examples working.
var requireControlTLS = flag.Bool("require-control-tls", false,
	"serve server registration only on the mTLS control listener")

func StartRelay() {
	logutils.SetLogStyle()
	kp, err := utils.LoadRelayKeyPair()
	if err != nil {
		fmt.Printf("error loading relay key: %s\n", err)
		return
	}
	r := relay.NewRelay(kp)

	// The control listener carries only the persistent registration connection. The dial and
	// callback hops stay plaintext: they are hijacked into the end-to-end encrypted tunnel, so
	// wrapping them here would just nest TLS inside TLS.
	controlTLS, err := utils.RelayControlTLSConfig()
	if err != nil {
		fmt.Printf("error loading relay control certs: %s\n", err)
		return
	}
	controlRelay := http.Server{
		Addr:              fmt.Sprintf("localhost:%d", utils.ControlPort),
		Handler:           r.ControlMux,
		TLSConfig:         controlTLS,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		if err := controlRelay.ListenAndServeTLS("", ""); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("error running relay control listener: %s\n", err)
			}
		}
	}()

	handler := r.Mux
	if *requireControlTLS {
		handler = r.DataMux
	}
	untrustedRelay := http.Server{
		Addr:              fmt.Sprintf("localhost:%d", utils.ServerPort),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := untrustedRelay.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http tcp_endpoints: %s\n", err)
		}
	}
}

func main() {
	flag.Parse()
	StartRelay()
}
