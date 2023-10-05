package main

import (
	"context"
	"errors"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/utils/logutils"
	"net"
	//"os"
	"strconv"
	"time"

	"github.ibm.com/mcnet-research/mbg_relay/pkg/relay"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/tcp_endpoints"

	"net/http"

	"fmt"
)

const ServerPort = 3333
const ServerName = "foo"
const ClientName = "bar"

var untrustedRelay http.Server

// StartRelay starts the main relay function.
func StartRelay() {
	r := relay.NewRelay()
	untrustedRelay = http.Server{
		Addr:    fmt.Sprintf("localhost:%d", ServerPort),
		Handler: r.Mux,
	}
	if err := untrustedRelay.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http tcp_endpoints: %s\n", err)
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// Read data from the client
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	// Convert the received data to a string
	message := string(buffer[:n])

	fmt.Printf("Received message from client: %s\n", message)

	toWrite := "received message: " + message

	_, err = conn.Write([]byte(toWrite))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

}

func AcceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		go handleClient(conn)
	}

}

func main() {
	logutils.SetLogStyle()
	//logrus.SetOutput(io.Discard)
	relayAddress := fmt.Sprintf("localhost:%d", ServerPort)

	//start relay
	go StartRelay()
	time.Sleep(1000 * time.Millisecond)

	//start tcp_endpoints
	listener, err := tcp_endpoints.ListenRelay(relayAddress, ServerName)
	if err != nil {
		return
	}
	defer listener.Close()
	go AcceptConnections(listener)

	for i := 1; i < 5; i++ {
		conn, err := tcp_endpoints.DialTCP(relayAddress, ClientName+string(rune(i)), ServerName)
		// Message to send
		message := "Hello, server! from " + strconv.Itoa(i)

		// Send the message to the tcp_endpoints
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		received := string(buffer[:n])

		fmt.Printf("Received message from server: %s\n", received)

		conn.Close()
	}

	time.Sleep(1000 * time.Millisecond)
	listener.Close()
	err = untrustedRelay.Shutdown(context.Background())
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
	}
	time.Sleep(1000 * time.Millisecond)

}
