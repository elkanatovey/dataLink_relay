package tcp_endpoints

// options configure how a RelayDialer or RelayListener talks to the relay.
type options struct {
	relayPub *[32]byte
}

// Option customises a RelayDialer or RelayListener.
type Option func(*options)

// WithRelayKey seals the routing metadata (client and server IDs) to the relay's public key so the
// network cannot read who is connecting to whom. Without it, routing metadata is sent in plaintext,
// exactly as before. The relay still reads the metadata in order to route.
func WithRelayKey(relayPub [32]byte) Option {
	return func(o *options) { o.relayPub = &relayPub }
}

func applyOptions(opts []Option) options {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
