package main

import (
	"fmt"
	"mbg-relay/example"
	"mbg-relay/relayconn/server"
	"mbg-relay/relayconn/utils/logutils"
	"net"
	"os"
)

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
}

func main() {
	logutils.SetLogStyle()
	relayAddress := fmt.Sprintf("localhost:%d", example.ServerPort)

	listener, err := server.Listen(relayAddress, example.ExporterName)
	if err != nil {
		return
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleClient(conn)
	}
}
