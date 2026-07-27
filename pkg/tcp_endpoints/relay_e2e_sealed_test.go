package tcp_endpoints

import (
	"github.com/elkanatovey/dataLink_relay/pkg/api"
	"github.com/elkanatovey/dataLink_relay/pkg/relay"
	"io"
	"net/http/httptest"
	"testing"
	"time"
)

// TestRelayEndToEndSealed runs the full relay path with routing metadata sealed to the relay's key:
// the listener and dialer seal their requests, the relay unseals them, and bytes flow end to end.
func TestRelayEndToEndSealed(t *testing.T) {
	kp, err := api.GenerateRelayKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	r := relay.NewRelay(kp)
	relayServer := httptest.NewServer(r.Mux)
	defer relayServer.Close()
	relayAddr := relayServer.Listener.Addr().String()

	const serverID = "e2e-sealed-server"

	listener, err := ListenRelay("tcp", serverID, relayAddr, WithRelayKey(kp.Public))
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
		_, err = io.Copy(conn, conn)
		serverErr <- err
	}()

	clientConn, err := DialTCP("tcp", serverID, relayAddr, "e2e-sealed-client", WithRelayKey(kp.Public))
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
