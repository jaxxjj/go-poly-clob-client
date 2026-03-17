package headers

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

const testPrivateKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestBuildL1(t *testing.T) {
	privateKey, err := crypto.HexToECDSA(testPrivateKeyHex)
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	h, err := BuildL1(privateKey, address, 80002, 0)
	if err != nil {
		t.Fatalf("BuildL1: %v", err)
	}

	required := []string{HeaderAddress, HeaderSignature, HeaderTimestamp, HeaderNonce}
	for _, key := range required {
		if h.Get(key) == "" {
			t.Errorf("missing header: %s", key)
		}
	}

	if h.Get(HeaderAddress) != address.Hex() {
		t.Errorf("address mismatch: got %s, want %s", h.Get(HeaderAddress), address.Hex())
	}
	if h.Get(HeaderNonce) != "0" {
		t.Errorf("nonce mismatch: got %s, want 0", h.Get(HeaderNonce))
	}

	t.Logf("L1 headers: address=%s sig=%s...", h.Get(HeaderAddress), h.Get(HeaderSignature)[:20])
}

func TestBuildL2(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA(testPrivateKeyHex)
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	creds := model.ApiCreds{
		APIKey:     "test-api-key",
		APISecret:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		Passphrase: "test-passphrase",
	}

	body := []byte(`{"orderID":"123"}`)
	h, err := BuildL2(address, creds, "POST", "/order", body)
	if err != nil {
		t.Fatalf("BuildL2: %v", err)
	}

	required := []string{HeaderAddress, HeaderSignature, HeaderTimestamp, HeaderAPIKey, HeaderPassphrase}
	for _, key := range required {
		if h.Get(key) == "" {
			t.Errorf("missing header: %s", key)
		}
	}

	if h.Get(HeaderAPIKey) != "test-api-key" {
		t.Errorf("api key mismatch: got %s", h.Get(HeaderAPIKey))
	}
	if h.Get(HeaderPassphrase) != "test-passphrase" {
		t.Errorf("passphrase mismatch: got %s", h.Get(HeaderPassphrase))
	}

	// L2 should NOT have nonce header
	if h.Get(HeaderNonce) != "" {
		t.Error("L2 headers should not include nonce")
	}

	t.Logf("L2 headers: api_key=%s sig=%s...", h.Get(HeaderAPIKey), h.Get(HeaderSignature)[:20])
}

func TestBuildL2_NilBody(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA(testPrivateKeyHex)
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	creds := model.ApiCreds{
		APIKey:     "key",
		APISecret:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		Passphrase: "pass",
	}

	h, err := BuildL2(address, creds, "GET", "/data/orders", nil)
	if err != nil {
		t.Fatalf("BuildL2: %v", err)
	}
	if h.Get(HeaderSignature) == "" {
		t.Error("expected non-empty signature for GET request")
	}
}
