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
	opts     []tcp_endpoints.Option
}

func (r RelayMTLSDialer) Dial(network, address string, config *tls.Config) (net.Conn, error) {
	if network != "tcp" {
		return nil, errors.New("only tcp supported")
	}
	return dial(context.Background(), r.relayIP, r.clientID, address, config, r.opts...)
}

// DialMTLS dials the given network address via the given relay
func DialMTLS(network, address string, config *tls.Config, relayIP string, clientName string, opts ...tcp_endpoints.Option) (net.Conn, error) {
	dialer := RelayMTLSDialer{relayIP: relayIP, clientID: clientName, opts: opts}
	return dialer.Dial(network, address, config)

}

func dial(ctx context.Context, relayIP string, clientName string, serverName string, config *tls.Config, opts ...tcp_endpoints.Option) (net.Conn, error) {
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
