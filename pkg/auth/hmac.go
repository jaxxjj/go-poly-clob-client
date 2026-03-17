// Package auth implements L1 (EIP-712) and L2 (HMAC-SHA256) authentication for the Polymarket CLOB API.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

// SignHMAC produces the HMAC-SHA256 signature for L2 authentication.
//
// secret is the base64url-encoded API secret from L1 credential derivation.
// body is the pre-serialized JSON body (from json.Marshal), or nil for bodyless requests.
//
// The signature message is: timestamp + method + path [+ body]
func SignHMAC(secret, timestamp, method, path string, body []byte) (string, error) {
	// Use RawURLEncoding to tolerate secrets without padding (matching Python's urlsafe_b64decode).
	key, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(secret, "="))
	if err != nil {
		return "", fmt.Errorf("decode api secret: %w", err)
	}

	message := timestamp + method + path
	if len(body) > 0 {
		message += string(body)
	}

	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))

	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}
