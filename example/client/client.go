package main

import (
	"bufio"
	"fmt"
	"github.ibm.com/mcnet-research/mbg_relay/example"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/tcp_endpoints"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/utils/logutils"
	"os"
)

func main() {
	logutils.SetLogStyle()
	relayAddress := fmt.Sprintf("localhost:%d", example.ServerPort)

	conn, err := tcp_endpoints.DialTCP(relayAddress, example.ImporterName, example.ExporterName)
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

		fmt.Printf("Received from tcp_endpoints: %s", response[:n])
	}

}
