package api

import (
	"bytes"
	"testing"
)

func TestSealOpenRoundTrip(t *testing.T) {
	kp, err := GenerateRelayKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	req := ConnectionRequest{ClientID: "alice", ServerID: "bob", Data: "payload"}

	blob, err := SealRouting(req, &kp.Public)
	if err != nil {
		t.Fatal(err)
	}
	// the sealed blob must not leak the routing identities in the clear
	if bytes.Contains(blob, []byte("alice")) || bytes.Contains(blob, []byte("bob")) {
		t.Fatal("sealed routing blob leaks identities")
	}

	var got ConnectionRequest
	if err := OpenRouting(blob, []*RelayKeyPair{kp}, &got); err != nil {
		t.Fatal(err)
	}
	if got != req {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", got, req)
	}
}

func TestOpenRoutingTrialsKeyring(t *testing.T) {
	oldKey, err := GenerateRelayKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	newKey, err := GenerateRelayKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	req := ListenRequest{ServerID: "svc"}
	blob, err := SealRouting(req, &oldKey.Public) // sealed to the retiring key
	if err != nil {
		t.Fatal(err)
	}

	// a relay mid-rotation (new key first, old still present) opens it
	var got ListenRequest
	if err := OpenRouting(blob, []*RelayKeyPair{newKey, oldKey}, &got); err != nil {
		t.Fatalf("trial decrypt failed: %v", err)
	}
	if got != req {
		t.Fatalf("mismatch: %+v", got)
	}

	// once the old key is fully retired it no longer opens
	if err := OpenRouting(blob, []*RelayKeyPair{newKey}, &got); err == nil {
		t.Fatal("expected failure opening with retired keyring")
	}
}

func TestEncodeRoutingPlaintextFallback(t *testing.T) {
	req := ConnectionAccept{ClientID: "a", ServerID: "b"}
	plain, err := EncodeRouting(req, nil) // nil key => plaintext JSON
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(plain, []byte(`"ClientID":"a"`)) {
		t.Fatalf("expected plaintext JSON, got %s", plain)
	}
}

func TestRelayKeyPairFromPrivate(t *testing.T) {
	kp, err := GenerateRelayKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	kp2, err := RelayKeyPairFromPrivate(kp.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	if kp2.Public != kp.Public {
		t.Fatal("derived public key does not match original")
	}
}
