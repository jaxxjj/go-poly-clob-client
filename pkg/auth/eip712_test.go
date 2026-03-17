package auth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Test private key from go-order-utils test suite (Hardhat account #0)
const testPrivateKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestSignClobAuth(t *testing.T) {
	privateKey, err := crypto.HexToECDSA(testPrivateKeyHex)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	expectedAddr := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	if address != expectedAddr {
		t.Fatalf("address mismatch: got %s, want %s", address.Hex(), expectedAddr.Hex())
	}

	sig, err := SignClobAuth(privateKey, address, 80002, 10000000, 23)
	if err != nil {
		t.Fatalf("SignClobAuth: %v", err)
	}

	if len(sig) < 4 {
		t.Fatal("signature too short")
	}
	if sig[:2] != "0x" {
		t.Error("signature should start with 0x")
	}
	t.Logf("ClobAuth signature: %s", sig)

	// Signature should be 65 bytes = 130 hex chars + "0x" = 132 chars
	if len(sig) != 132 {
		t.Errorf("expected 132 char signature (0x + 130), got %d", len(sig))
	}
}

func TestSignClobAuth_Deterministic(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA(testPrivateKeyHex)
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	sig1, err := SignClobAuth(privateKey, address, 80002, 10000000, 0)
	if err != nil {
		t.Fatalf("first sign: %v", err)
	}

	sig2, err := SignClobAuth(privateKey, address, 80002, 10000000, 0)
	if err != nil {
		t.Fatalf("second sign: %v", err)
	}

	if sig1 != sig2 {
		t.Errorf("expected deterministic signatures:\n  got %s\n  and %s", sig1, sig2)
	}
}

func TestSignClobAuth_DifferentInputsDifferentOutput(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA(testPrivateKeyHex)
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	sig1, _ := SignClobAuth(privateKey, address, 80002, 10000000, 0)
	sig2, _ := SignClobAuth(privateKey, address, 80002, 10000001, 0)
	if sig1 == sig2 {
		t.Error("different timestamps should produce different signatures")
	}

	sig3, _ := SignClobAuth(privateKey, address, 80002, 10000000, 1)
	if sig1 == sig3 {
		t.Error("different nonces should produce different signatures")
	}

	sig4, _ := SignClobAuth(privateKey, address, 137, 10000000, 0)
	if sig1 == sig4 {
		t.Error("different chain IDs should produce different signatures")
	}
}
