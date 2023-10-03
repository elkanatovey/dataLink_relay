package relay

import (
	"errors"
	"fmt"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/api"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"net/http"
)

// getWaitingClientId calculates the id of a waiting request based on relevant client/server ids
func getWaitingClientId(cr api.ConnectionRequest) string {
	return cr.ClientID + cr.ServerID
}

// getCallingServerId calculates the id of a callback response based on relevant client/server ids
func getCallingServerId(ca api.ConnectionAccept) string {
	return ca.ClientID + ca.ServerID
}

// Hijack the HTTP connection and use the TCP session
func hijackConn(w http.ResponseWriter) net.Conn {
	// Check if we can hijack connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server doesn't support hijacking", http.StatusInternalServerError)
		return nil
	}
	w.WriteHeader(http.StatusOK) //should this be here?
	// Hijack the connection
	conn, _, _ := hj.Hijack()
	return conn
}

// uniteConnections glues an importer connection with an exporter connection
func uniteConnections(clientConn net.Conn, serverConn net.Conn) error {

	var eg errgroup.Group
	//fmt.Printf("uniteconn ")

	eg.Go(func() error {
		defer clientConn.Close()
		defer serverConn.Close()

		_, err := io.Copy(serverConn, clientConn)
		//fmt.Printf("server written bytes %s \n", nb)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("server->client: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer clientConn.Close()
		defer serverConn.Close()

		_, err := io.Copy(clientConn, serverConn)
		//fmt.Printf("client written bytes %s \n", cnb)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("client->server: %w", err)
		}

		return nil
	})

	return eg.Wait()
}
