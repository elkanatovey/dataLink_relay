package mtls_endpoint

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/elkanatovey/dataLink_relay/pkg/tcp_endpoints"
	"net"
)

// RelayMTLSDialer connects to a server via a relay
type RelayMTLSDialer struct {
	relayIP  string
	clientID string
	opts     []Option
}

func (r RelayMTLSDialer) Dial(network, address string, config *tls.Config) (net.Conn, error) {
	if network != "tcp" {
		return nil, errors.New("only tcp supported")
	}
	return dial(context.Background(), r.relayIP, r.clientID, address, config, r.opts...)
}

// NewRelayMTLSDialer creates a RelayMTLSDialer that maintains the tls.Dial api. To dial call
// RelayMTLSDialer.Dial. relayIP is the address of the relay via which we dial, clientName is the id
// this client presents to the relay
func NewRelayMTLSDialer(relayIP string, clientName string, opts ...Option) RelayMTLSDialer {
	return RelayMTLSDialer{relayIP: relayIP, clientID: clientName, opts: opts}
}

// DialMTLS dials the given network address via the given relay
func DialMTLS(network, address string, config *tls.Config, relayIP string, clientName string, opts ...Option) (net.Conn, error) {
	return NewRelayMTLSDialer(relayIP, clientName, opts...).Dial(network, address, config)
}

func dial(ctx context.Context, relayIP string, clientName string, serverName string, config *tls.Config, opts ...Option) (net.Conn, error) {
	rawConn, err := tcp_endpoints.DialTCP("tcp", serverName, relayIP, clientName, opts...)
	if err != nil {
		return nil, err
	}
	conn := tls.Client(rawConn, config)
	if err := conn.HandshakeContext(ctx); err != nil {
		rawConn.Close()
		return nil, err
	}
	return conn, nil
}
