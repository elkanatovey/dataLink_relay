package tcp_endpoints

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elkanatovey/dataLink_relay/pkg/relay"
)

// controlPKI is a throwaway CA plus the two leaves the control plane needs.
type controlPKI struct {
	pool  *x509.CertPool
	relay tls.Certificate
}

// newControlPKI mints a CA, a relay server leaf valid for 127.0.0.1, and returns both along with a
// factory for client leaves carrying a given server id as their SAN.
func newControlPKI(t *testing.T) (*controlPKI, func(t *testing.T, serverID string) tls.Certificate) {
	t.Helper()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "control test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	leaf := func(t *testing.T, eku x509.ExtKeyUsage, dns []string, ips []net.IP) tls.Certificate {
		t.Helper()
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
		if err != nil {
			t.Fatal(err)
		}
		tmpl := &x509.Certificate{
			SerialNumber: serial,
			Subject:      pkix.Name{CommonName: "control test leaf"},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(time.Hour),
			KeyUsage:     x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  []x509.ExtKeyUsage{eku},
			DNSNames:     dns,
			IPAddresses:  ips,
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
		if err != nil {
			t.Fatal(err)
		}
		return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
	}

	pki := &controlPKI{
		pool:  pool,
		relay: leaf(t, x509.ExtKeyUsageServerAuth, []string{"localhost"}, []net.IP{net.ParseIP("127.0.0.1")}),
	}
	clientFor := func(t *testing.T, serverID string) tls.Certificate {
		return leaf(t, x509.ExtKeyUsageClientAuth, []string{serverID}, nil)
	}
	return pki, clientFor
}

func (p *controlPKI) relayConfig() *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{p.relay},
		ClientCAs:    p.pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}
}

func (p *controlPKI) exporterConfig(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      p.pool,
		MinVersion:   tls.VersionTLS13,
	}
}

// startSplitRelay serves the hijacked hops on a plaintext listener and registration on an mTLS
// listener, which is the deployment where registration cannot be done without a certificate.
func startSplitRelay(t *testing.T, pki *controlPKI) (dataAddr, controlAddr string) {
	t.Helper()
	r := relay.NewRelay()

	dataSrv := httptest.NewServer(r.DataMux)
	t.Cleanup(dataSrv.Close)

	controlSrv := httptest.NewUnstartedServer(r.ControlMux)
	controlSrv.TLS = pki.relayConfig()
	controlSrv.StartTLS()
	t.Cleanup(controlSrv.Close)

	return dataSrv.Listener.Addr().String(), controlSrv.Listener.Addr().String()
}

// TestRelayControlMTLSEndToEnd registers over the mTLS control listener and then moves bytes over
// the plaintext data hops, confirming the split listeners interoperate.
func TestRelayControlMTLSEndToEnd(t *testing.T) {
	const serverID = "control-mtls-server"

	pki, clientFor := newControlPKI(t)
	dataAddr, controlAddr := startSplitRelay(t, pki)

	listener, err := ListenRelay("tcp", serverID, dataAddr,
		WithRelayControlTLS(controlAddr, pki.exporterConfig(clientFor(t, serverID))))
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

	clientConn, err := DialTCP("tcp", serverID, dataAddr, "control-mtls-client")
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

// TestRelayControlMTLSRejectsForeignServerID is the property the control plane exists for: holding a
// valid certificate is not enough, it must actually name the server id being registered.
func TestRelayControlMTLSRejectsForeignServerID(t *testing.T) {
	pki, clientFor := newControlPKI(t)
	dataAddr, controlAddr := startSplitRelay(t, pki)

	// A well formed certificate from the control CA, but issued for a different server id.
	cfg := pki.exporterConfig(clientFor(t, "some-other-server"))

	listener, err := ListenRelay("tcp", "victim-server", dataAddr, WithRelayControlTLS(controlAddr, cfg))
	if err == nil {
		listener.Close()
		t.Fatal("expected registration to be refused for a server id the certificate does not cover")
	}
	// The certificate itself is trusted, so this has to be an authorisation refusal. Asserting on it
	// keeps the test from passing because of an unrelated transport or handshake failure.
	if !strings.Contains(err.Error(), http.StatusText(http.StatusForbidden)) {
		t.Fatalf("expected a forbidden response, got: %v", err)
	}
}

// TestRelayControlMuxRequiresCertificate guards against serving ControlMux on a plaintext listener.
func TestRelayControlMuxRequiresCertificate(t *testing.T) {
	r := relay.NewRelay()
	srv := httptest.NewServer(r.ControlMux)
	defer srv.Close()

	listener, err := ListenRelay("tcp", "plaintext-server", srv.Listener.Addr().String())
	if err == nil {
		listener.Close()
		t.Fatal("expected plaintext registration against ControlMux to be refused")
	}
	// The route exists on this mux, so a refusal here means the certificate check rejected it rather
	// than the request having gone somewhere else.
	if !strings.Contains(err.Error(), http.StatusText(http.StatusForbidden)) {
		t.Fatalf("expected a forbidden response, got: %v", err)
	}
}
