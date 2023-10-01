package main

import (
	"fmt"
	"io/ioutil"
	"mbg-relay/example"
	"mbg-relay/relayconn/client"
	"mbg-relay/relayconn/relay"
	"net"
)

func ReadNWrite(conn net.Conn) {
	message := "Test Request\n"
	_, write_err := conn.Write([]byte(message))
	if write_err != nil {
		fmt.Println("failed:", write_err)
		return
	}
	conn.(*net.TCPConn).CloseWrite()

	buf, read_err := ioutil.ReadAll(conn)
	if read_err != nil {
		fmt.Println("failed:", read_err)
		return
	}
	fmt.Println(string(buf))
}

func main() {
	relayAddress := fmt.Sprintf("localhost:%d", example.ServerPort)

	conn, err := client.DialTCP(relayAddress, example.ImporterName, example.ExporterName)
	// Message to send
	//message := "Hello, server!"

	// Send the message to the server
	//_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
	conn.Close()
	fmt.Println("random number:", relay.MaintainConnection())
}
