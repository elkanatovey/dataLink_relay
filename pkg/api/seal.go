package api

import (
	"crypto/rand"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

// RelayKeyPair is an X25519 keypair the relay uses to open routing metadata that clients and
// servers seal to it. Sealing hides the routing metadata (who is connecting to whom) from the
// network; the relay still reads it in order to route.
type RelayKeyPair struct {
	Public  [32]byte
	private [32]byte
}

// GenerateRelayKeyPair returns a fresh X25519 keypair for sealing routing metadata.
func GenerateRelayKeyPair() (*RelayKeyPair, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &RelayKeyPair{Public: *pub, private: *priv}, nil
}

// RelayKeyPairFromPrivate reconstructs a keypair from a 32-byte X25519 private key.
func RelayKeyPairFromPrivate(private [32]byte) (*RelayKeyPair, error) {
	pub, err := curve25519.X25519(private[:], curve25519.Basepoint)
	if err != nil {
		return nil, err
	}
	kp := &RelayKeyPair{private: private}
	copy(kp.Public[:], pub)
	return kp, nil
}

// PrivateKey returns the raw private key bytes, for persisting a relay identity.
func (kp *RelayKeyPair) PrivateKey() [32]byte {
	return kp.private
}

// SealRouting seals v (marshalled as JSON) to relayPub. The result is opaque to anyone without
// the matching relay private key; only the relay can open it.
func SealRouting(v any, relayPub *[32]byte) ([]byte, error) {
	plain, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return box.SealAnonymous(nil, plain, relayPub, rand.Reader)
}

// EncodeRouting seals v to relayPub when it is non-nil, otherwise it marshals v as plaintext JSON.
// It is the client/server-side counterpart for sending routing metadata to the relay.
func EncodeRouting(v any, relayPub *[32]byte) ([]byte, error) {
	if relayPub != nil {
		return SealRouting(v, relayPub)
	}
	return json.Marshal(v)
}

// OpenRouting trial-decrypts a sealed blob against the keyring and unmarshals it into v. Trying
// multiple keys lets the relay rotate its key without dropping clients that still use the old one.
func OpenRouting(blob []byte, ring []*RelayKeyPair, v any) error {
	for _, kp := range ring {
		if plain, ok := box.OpenAnonymous(nil, blob, &kp.Public, &kp.private); ok {
			return json.Unmarshal(plain, v)
		}
	}
	return errors.New("api: sealed routing message could not be opened with any relay key")
}
