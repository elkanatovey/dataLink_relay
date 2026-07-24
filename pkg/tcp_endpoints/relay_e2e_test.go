package tcp_endpoints

import (
	"github.com/elkanatovey/dataLink_relay/pkg/relay"
	"io"
	"net/http/httptest"
	"testing"
	"time"
)

// TestRelayEndToEnd exercises the full path: a RelayListener registers with the relay,
// a RelayDialer connects through it, and bytes are echoed back over the glued socket.
func TestRelayEndToEnd(t *testing.T) {
	r := relay.NewRelay()
	relayServer := httptest.NewServer(r.Mux)
	defer relayServer.Close()
	relayAddr := relayServer.Listener.Addr().String()

	const serverID = "e2e-server"

	listener, err := ListenRelay("tcp", serverID, relayAddr)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		_, err = io.Copy(conn, conn) // echo until client closes
		serverErr <- err
	}()

	clientConn, err := DialTCP("tcp", serverID, relayAddr, "e2e-client")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer clientConn.Close()

	msg := []byte("ping")
	if _, err := clientConn.Write(msg); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := make([]byte, len(msg))
	if _, err := io.ReadFull(clientConn, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(msg) {
		t.Fatalf("echo mismatch: got %q want %q", got, msg)
	}

	// closing the client releases the server-side echo copy
	clientConn.Close()
	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("server: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not finish after client close")
	}
}
