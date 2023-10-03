package main

import (
	"fmt"
	"github.ibm.com/mcnet-research/mbg_relay/example"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/client"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/utils/logutils"
	"os"
)

func main() {
	logutils.SetLogStyle()
	relayAddress := fmt.Sprintf("localhost:%d", example.ServerPort)

	conn, err := client.DialTCP(relayAddress, example.ImporterName, example.ExporterName)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}

	defer conn.Close()

	for {
		// Read user input from the terminal.
		fmt.Print("Enter a message (or 'exit' to quit): ")
		var userInput string
		fmt.Scanln(&userInput)

		if userInput == "exit" {
			fmt.Println("Exiting client.")
			return
		}

		// Send the user's input to the server.
		_, err := conn.Write([]byte(userInput + "\n"))
		if err != nil {
			fmt.Println("Error sending data:", err)
			return
		}

		// Receive and print the server's response.
		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			fmt.Println("Error receiving response:", err)
			return
		}

		fmt.Printf("Received from server: %s", response[:n])
	}

}
