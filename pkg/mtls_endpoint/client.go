package mtls_endpoint

import (
	"context"
	"crypto/tls"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/tcp_endpoints"
	"net"
)

// RelayMTLSDialer connects to a server via a relay
type RelayMTLSDialer struct {
	relayIP  string
	clientID string
}

func (r RelayMTLSDialer) Dial(network, address string, config *tls.Config) (net.Conn, error) {
	if network != "tcp" {
		panic(" only tcp supported")
	}
	return dial(context.Background(), r.relayIP, r.clientID, address, config)
}

// DialMTLS dials the given network address via the given relay
func DialMTLS(network, address string, config *tls.Config, relayIP string, clientName string) (net.Conn, error) {
	dialer := RelayMTLSDialer{relayIP: relayIP, clientID: clientName}
	return dialer.Dial(network, address, config)

}

func dial(ctx context.Context, relayIP string, clientName string, serverName string, config *tls.Config) (net.Conn, error) {
	rawConn, err := tcp_endpoints.DialTCP(relayIP, clientName, serverName)
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
