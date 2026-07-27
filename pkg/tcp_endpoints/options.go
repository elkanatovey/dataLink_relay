package tcp_endpoints

import "crypto/tls"

// options configure how a RelayDialer or RelayListener talks to the relay.
type options struct {
	relayPub    *[32]byte
	controlAddr string
	controlTLS  *tls.Config
}

// Option customises a RelayDialer or RelayListener.
type Option func(*options)

// WithRelayKey seals the routing metadata (client and server IDs) to the relay's public key so the
// network cannot read who is connecting to whom. Without it, routing metadata is sent in plaintext,
// exactly as before. The relay still reads the metadata in order to route.
func WithRelayKey(relayPub [32]byte) Option {
	return func(o *options) { o.relayPub = &relayPub }
}

// WithRelayControlTLS sends the listener's persistent registration connection to the relay's mTLS
// control endpoint at addr, using cfg. This encrypts and authenticates that connection, and lets the
// relay refuse a server id the certificate does not cover.
//
// cfg authenticates the listener to the *relay* and is unrelated to any tls.Config used for the
// end-to-end session with a client: those are different peers and should use different credentials.
// The client dial and server callback hops are unaffected, so tunnel traffic is never nested inside
// a second layer of TLS.
func WithRelayControlTLS(addr string, cfg *tls.Config) Option {
	return func(o *options) {
		o.controlAddr = addr
		o.controlTLS = cfg
	}
}

func applyOptions(opts []Option) options {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
