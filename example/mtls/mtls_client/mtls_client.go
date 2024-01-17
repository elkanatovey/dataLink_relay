package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"github.com/elkanatovey/dataLink_relay/example/utils"
	log "github.com/sirupsen/logrus"

	"fmt"

	"github.com/elkanatovey/dataLink_relay/pkg/mtls_endpoint"
	"github.com/elkanatovey/dataLink_relay/pkg/utils/logutils"
	"os"
)

func main() {
	logutils.SetLogStyle()

	// load CA certificate file and add it to list of client CAs
	caCertFile, err := os.ReadFile(utils.CertFile)
	if err != nil {
		log.Fatalf("error reading CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertFile)
	cert, err := tls.LoadX509KeyPair(utils.CertFile, utils.KeyFile)
	if err != nil {
		fmt.Println("Error loading client certificates:", err)
		os.Exit(1)
	}
	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs:          caCertPool,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		MinVersion:         tls.VersionTLS12,
		CurvePreferences:   []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	relayAddress := fmt.Sprintf("localhost:%d", utils.ServerPort)

	conn, err := mtls_endpoint.DialMTLS("tcp", utils.ExporterName, tlsConfig, relayAddress, utils.ImporterName)
	if err != nil {
		fmt.Println("Error connecting to tcp_endpoints:", err)
		os.Exit(1)
	}

	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read user input from the terminal.
		fmt.Print("Enter a message (or 'exit' to quit): ")
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		if userInput == "exit\n" {
			fmt.Println("Exiting client.")
			return
		}

		// Send the user's input to the tcp_endpoints.
		_, err = conn.Write([]byte(userInput))
		if err != nil {
			fmt.Println("Error sending data:", err)
			return
		}

		// Receive and print the tcp_endpoints's response.
		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			fmt.Println("Error receiving response:", err)
			return
		}

		fmt.Printf("Received from server: %s", response[:n])
	}

}
