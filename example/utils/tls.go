package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Demo PKI produced by `go run ./example/utils/gencerts`. For local demos only.
const (
	CACertFile     = "example/utils/certs/ca-cert.pem"
	ServerCertFile = "example/utils/certs/server-cert.pem"
	ServerKeyFile  = "example/utils/certs/server-key.pem"
	ClientCertFile = "example/utils/certs/client-cert.pem"
	ClientKeyFile  = "example/utils/certs/client-key.pem"

	// Control-plane PKI, used only between a listening server and the relay. Separate from the
	// end-to-end PKI above so that a client certificate cannot be used to register a server id.
	ControlCACertFile    = "example/utils/certs/control-ca-cert.pem"
	RelayControlCertFile = "example/utils/certs/relay-control-cert.pem"
	RelayControlKeyFile  = "example/utils/certs/relay-control-key.pem"
	ExporterCtrlCertFile = "example/utils/certs/exporter-control-cert.pem"
	ExporterCtrlKeyFile  = "example/utils/certs/exporter-control-key.pem"
	RelayControlTLSName  = "localhost"

	// ServerTLSName must match a SAN in the demo server certificate.
	ServerTLSName = "foo"
)

func certPool(file string) (*x509.CertPool, error) {
	caPEM, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("no certificates parsed from %s", file)
	}
	return pool, nil
}

func caCertPool() (*x509.CertPool, error) {
	return certPool(CACertFile)
}

// ClientTLSConfig returns the demo client mTLS config. It verifies the server against the demo CA,
// so it does not use InsecureSkipVerify.
func ClientTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(ClientCertFile, ClientKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load client cert: %w", err)
	}
	pool, err := caCertPool()
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   ServerTLSName,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// ServerTLSConfig returns the demo server mTLS config. It requires and verifies a client
// certificate issued by the demo CA.
func ServerTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(ServerCertFile, ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server cert: %w", err)
	}
	pool, err := caCertPool()
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// RelayControlTLSConfig returns the config the relay serves its mTLS control listener with. It
// requires a client certificate from the control CA, which is not the CA that signs client or
// server certificates for the end-to-end session.
func RelayControlTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(RelayControlCertFile, RelayControlKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load relay control cert: %w", err)
	}
	pool, err := certPool(ControlCACertFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// ExporterControlTLSConfig returns the config a listening server uses for its registration
// connection to the relay. The certificate's SAN is the server id it is allowed to register.
func ExporterControlTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(ExporterCtrlCertFile, ExporterCtrlKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load exporter control cert: %w", err)
	}
	pool, err := certPool(ControlCACertFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   RelayControlTLSName,
		MinVersion:   tls.VersionTLS13,
	}, nil
}
