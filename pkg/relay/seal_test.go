package relay

import (
	"encoding/json"
	"testing"

	"github.com/elkanatovey/dataLink_relay/pkg/api"
)

func TestDecodeRoutingSealedAndPlaintext(t *testing.T) {
	kp, err := api.GenerateRelayKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	d := initRelayData()
	d.SetRoutingKeys([]*api.RelayKeyPair{kp})

	req := api.ConnectionRequest{ClientID: "c", ServerID: "s"}

	// a sealed body opens with the keyring
	sealed, err := api.SealRouting(req, &kp.Public)
	if err != nil {
		t.Fatal(err)
	}
	var got api.ConnectionRequest
	if err := d.decodeRouting(sealed, &got); err != nil {
		t.Fatalf("decode sealed: %v", err)
	}
	if got != req {
		t.Fatalf("sealed mismatch: got %+v want %+v", got, req)
	}

	// a plaintext body still opens via the fallback
	plain, _ := json.Marshal(req)
	var got2 api.ConnectionRequest
	if err := d.decodeRouting(plain, &got2); err != nil {
		t.Fatalf("decode plaintext: %v", err)
	}
	if got2 != req {
		t.Fatalf("plaintext mismatch: got %+v want %+v", got2, req)
	}
}
