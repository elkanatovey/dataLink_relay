package mtls_endpoint

import (
	"crypto/tls"

	"github.com/elkanatovey/dataLink_relay/pkg/tcp_endpoints"
)

// Option customises how a RelayMTLSDialer or MTLSRelayListener talks to the relay. It is an alias
// for tcp_endpoints.Option, so options from either package are interchangeable.
type Option = tcp_endpoints.Option

// WithRelayKey seals the routing metadata (client and server IDs) to the relay's public key so the
// network cannot read who is connecting to whom. Without it, routing metadata is sent in plaintext.
// The relay still reads the metadata in order to route.
func WithRelayKey(relayPub [32]byte) Option {
	return tcp_endpoints.WithRelayKey(relayPub)
}

// WithRelayControlTLS sends the listener's persistent registration connection to the relay's mTLS
// control endpoint at addr, using cfg. See tcp_endpoints.WithRelayControlTLS.
//
// cfg authenticates this listener to the *relay*. It is a different peer from the client that the
// tls.Config passed to Listen authenticates, so the two should not share credentials: anything the
// relay accepts as a registration certificate could otherwise be minted by a client.
func WithRelayControlTLS(addr string, cfg *tls.Config) Option {
	return tcp_endpoints.WithRelayControlTLS(addr, cfg)
}
