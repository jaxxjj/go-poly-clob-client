package auth

import (
	"testing"
)

func TestSignHMAC(t *testing.T) {
	// Known test vector: secret is base64url-encoded 32 zero bytes
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	timestamp := "1000000"
	method := "GET"
	path := "/order"

	sig, err := SignHMAC(secret, timestamp, method, path, nil)
	if err != nil {
		t.Fatalf("SignHMAC: %v", err)
	}
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	t.Logf("HMAC sig (no body): %s", sig)
}

func TestSignHMAC_WithBody(t *testing.T) {
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	timestamp := "1000000"
	method := "POST"
	path := "/order"
	body := []byte(`{"orderID":"123"}`)

	sig, err := SignHMAC(secret, timestamp, method, path, body)
	if err != nil {
		t.Fatalf("SignHMAC: %v", err)
	}
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	t.Logf("HMAC sig (with body): %s", sig)
}

func TestSignHMAC_Deterministic(t *testing.T) {
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	timestamp := "1000000"
	method := "DELETE"
	path := "/order"
	body := []byte(`{"orderID":"abc"}`)

	sig1, _ := SignHMAC(secret, timestamp, method, path, body)
	sig2, _ := SignHMAC(secret, timestamp, method, path, body)
	if sig1 != sig2 {
		t.Errorf("expected deterministic signatures, got %q and %q", sig1, sig2)
	}
}

func TestSignHMAC_DifferentInputsDifferentOutput(t *testing.T) {
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	sig1, _ := SignHMAC(secret, "1000000", "GET", "/order", nil)
	sig2, _ := SignHMAC(secret, "1000001", "GET", "/order", nil)
	if sig1 == sig2 {
		t.Error("different timestamps should produce different signatures")
	}

	sig3, _ := SignHMAC(secret, "1000000", "POST", "/order", nil)
	if sig1 == sig3 {
		t.Error("different methods should produce different signatures")
	}
}

func TestSignHMAC_InvalidSecret(t *testing.T) {
	_, err := SignHMAC("not-valid-base64!!!", "1000000", "GET", "/", nil)
	if err == nil {
		t.Error("expected error for invalid base64 secret")
	}
}
