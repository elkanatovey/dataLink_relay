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

	// ServerTLSName must match a SAN in the demo server certificate.
	ServerTLSName = "foo"
)

func caCertPool() (*x509.CertPool, error) {
	caPEM, err := os.ReadFile(CACertFile)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("no certificates parsed from %s", CACertFile)
	}
	return pool, nil
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
