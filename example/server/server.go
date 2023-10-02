package main

import (
	"bufio"
	"fmt"
	"mbg-relay/example"
	"mbg-relay/pkg/server"
	"mbg-relay/pkg/utils/logutils"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())

	// Create a buffered reader to read messages from the client.
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
		go handleConnection(conn)
	}
}
