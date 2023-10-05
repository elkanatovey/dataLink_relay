package mtls_endpoint

import (
	"context"
	"crypto/tls"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/tcp_endpoints"
	"net"
)

// DialMTLS interprets a nil configuration as equivalent to the zero
// configuration; see the documentation of tls.Config for the defaults.
func DialMTLS(ctx context.Context, relayIP string, clientName string, serverName string, config *tls.Config) (net.Conn, error) {
	return dial(context.Background(), relayIP, clientName, serverName, config)
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
