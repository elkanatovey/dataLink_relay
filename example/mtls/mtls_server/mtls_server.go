package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"github.ibm.com/mcnet-research/mbg_relay/example/utils"

	//"crypto/x509"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/mtls_endpoint"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/utils/logutils"
	//"io"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())

	//// Create a buffered reader to read messages from the client.
	//_, err := io.Copy(conn, conn)
	//if err != nil {
	//	fmt.Println("Error writing to connection:", err)
	//	return
	//}

	reader := bufio.NewReader(conn)

	for {
		// Read a message from the client.
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}

		// Trim any leading/trailing whitespace and print the message.
		message = strings.TrimSpace(message)
		fmt.Printf("Received from client %s: %s\n", conn.RemoteAddr(), message)

		// Echo the message back to the client.
		_, err = conn.Write([]byte(message + "\n"))
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			return
		}
	}
}

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
		ClientCAs:        caCertPool,
		ClientAuth:       tls.RequireAndVerifyClientCert,
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		Certificates:     []tls.Certificate{cert},
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

	listener, err := mtls_endpoint.ListenMTLS("tcp", utils.ExporterName, tlsConfig, relayAddress)
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
