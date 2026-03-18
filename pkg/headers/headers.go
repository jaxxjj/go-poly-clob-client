// Package headers builds authentication HTTP headers for the Polymarket CLOB API.
package headers

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/jaxxjj/go-poly-clob-client/pkg/auth"
	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

// Header key constants matching the Polymarket CLOB API spec.
const (
	HeaderAddress    = "POLY_ADDRESS"
	HeaderSignature  = "POLY_SIGNATURE"
	HeaderTimestamp  = "POLY_TIMESTAMP"
	HeaderNonce      = "POLY_NONCE"
	HeaderAPIKey     = "POLY_API_KEY"
	HeaderPassphrase = "POLY_PASSPHRASE"
)

// BuildL1 creates Level 1 (EIP-712) authentication headers.
func BuildL1(
	privateKey *ecdsa.PrivateKey,
	address common.Address,
	chainID int64,
	nonce int64,
) (http.Header, error) {
	timestamp := time.Now().Unix()

	sig, err := auth.SignClobAuth(privateKey, address, chainID, timestamp, nonce)
	if err != nil {
		return nil, fmt.Errorf("sign clob auth: %w", err)
	}

	h := make(http.Header)
	h.Set(HeaderAddress, address.Hex())
	h.Set(HeaderSignature, sig)
	h.Set(HeaderTimestamp, fmt.Sprintf("%d", timestamp))
	h.Set(HeaderNonce, fmt.Sprintf("%d", nonce))
	return h, nil
}

// BuildL2 creates Level 2 (HMAC-SHA256) authentication headers.
//
// body should be the pre-serialized JSON body (result of json.Marshal), or nil.
func BuildL2(
	address common.Address,
	creds model.ApiCreds,
	method, path string,
	body []byte,
) (http.Header, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// HMAC is computed over the base path only, without query parameters.
	sigPath := path
	if idx := strings.IndexByte(sigPath, '?'); idx != -1 {
		sigPath = sigPath[:idx]
	}

	sig, err := auth.SignHMAC(creds.APISecret, timestamp, method, sigPath, body)
	if err != nil {
		return nil, fmt.Errorf("sign hmac: %w", err)
	}

	h := make(http.Header)
	h.Set(HeaderAddress, address.Hex())
	h.Set(HeaderSignature, sig)
	h.Set(HeaderTimestamp, timestamp)
	h.Set(HeaderAPIKey, creds.APIKey)
	h.Set(HeaderPassphrase, creds.Passphrase)
	return h, nil
}
