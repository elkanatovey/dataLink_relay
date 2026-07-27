package main

import (
	"bufio"
	"fmt"
	"github.com/elkanatovey/dataLink_relay/example/utils"
	"github.com/elkanatovey/dataLink_relay/pkg/mtls_endpoint"
	"github.com/elkanatovey/dataLink_relay/pkg/utils/logutils"
	log "github.com/sirupsen/logrus"
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

	tlsConfig, err := utils.ServerTLSConfig()
	if err != nil {
		log.Fatalf("tls config: %v", err)
	}

	relayPub, err := utils.LoadRelayPublicKey()
	if err != nil {
		log.Fatalf("relay key: %v", err)
	}

	relayAddress := fmt.Sprintf("localhost:%d", utils.ServerPort)

	listener, err := mtls_endpoint.ListenMTLS("tcp", utils.ExporterName, tlsConfig, relayAddress, mtls_endpoint.WithRelayKey(relayPub))
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
