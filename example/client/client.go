package main

import (
	"fmt"
	"mbg-relay/example"
	"mbg-relay/relayconn/client"
	"mbg-relay/relayconn/utils/logutils"
)

func main() {
	logutils.SetLogStyle()
	relayAddress := fmt.Sprintf("localhost:%d", example.ServerPort)

	conn, err := client.DialTCP(relayAddress, example.ImporterName, example.ExporterName)
	// Message to send
	message := "Hello, server!"

	// Send the message to the server
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
	conn.Close()
}
