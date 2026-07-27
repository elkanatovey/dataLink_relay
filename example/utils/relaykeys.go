package utils

import (
	"fmt"
	"os"

	"github.com/elkanatovey/dataLink_relay/pkg/api"
)

// Relay routing keypair produced by `go run ./example/utils/gencerts`. For local demos only. The
// private key stays with the relay; clients and servers only need the public key.
const (
	RelayKeyFile = "example/utils/certs/relay.key"
	RelayPubFile = "example/utils/certs/relay.pub"
)

// LoadRelayKeyPair loads the relay's X25519 keypair, used by the relay to open sealed routing metadata.
func LoadRelayKeyPair() (*api.RelayKeyPair, error) {
	priv, err := os.ReadFile(RelayKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read relay key: %w", err)
	}
	if len(priv) != 32 {
		return nil, fmt.Errorf("relay key %s: expected 32 bytes, got %d", RelayKeyFile, len(priv))
	}
	var p [32]byte
	copy(p[:], priv)
	return api.RelayKeyPairFromPrivate(p)
}

// LoadRelayPublicKey loads the relay's public key, used by clients and servers to seal routing metadata.
func LoadRelayPublicKey() ([32]byte, error) {
	var pub [32]byte
	data, err := os.ReadFile(RelayPubFile)
	if err != nil {
		return pub, fmt.Errorf("read relay public key: %w", err)
	}
	if len(data) != 32 {
		return pub, fmt.Errorf("relay public key %s: expected 32 bytes, got %d", RelayPubFile, len(data))
	}
	copy(pub[:], data)
	return pub, nil
}
