package auth

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/polymarket/go-order-utils/pkg/eip712"
	"github.com/polymarket/go-order-utils/pkg/signer"
)

// ClobAuth EIP-712 domain constants.
const (
	ClobDomainName = "ClobAuthDomain"
	ClobVersion    = "1"
	MsgToSign      = "This message attests that I control the given wallet"
)

// clobAuthTypeHash is the keccak256 of the ClobAuth EIP-712 struct type.
var clobAuthTypeHash = crypto.Keccak256Hash(
	[]byte("ClobAuth(address address,string timestamp,uint256 nonce,string message)"),
)

// clobAuthTypes defines the ABI types for encoding the ClobAuth struct.
var clobAuthTypes = []abi.Type{
	eip712.Bytes32, // typehash
	eip712.Address, // address
	eip712.Bytes32, // timestamp (keccak256 of string)
	eip712.Uint256, // nonce
	eip712.Bytes32, // message (keccak256 of string)
}

// SignClobAuth produces the EIP-712 signature for L1 CLOB authentication.
//
// The returned signature is a hex-encoded string with "0x" prefix.
func SignClobAuth(
	privateKey *ecdsa.PrivateKey,
	address common.Address,
	chainID int64,
	timestamp int64,
	nonce int64,
) (string, error) {
	// Build domain separator (no verifying contract for ClobAuth)
	nameHash := crypto.Keccak256Hash([]byte(ClobDomainName))
	versionHash := crypto.Keccak256Hash([]byte(ClobVersion))

	domainSeparator, err := eip712.BuildEIP712DomainSeparatorNoContract(
		nameHash, versionHash, big.NewInt(chainID),
	)
	if err != nil {
		return "", fmt.Errorf("build domain separator: %w", err)
	}

	// EIP-712 string fields are keccak256-hashed before encoding
	timestampHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("%d", timestamp)))
	messageHash := crypto.Keccak256Hash([]byte(MsgToSign))

	values := []interface{}{
		clobAuthTypeHash,
		address,
		timestampHash,
		big.NewInt(nonce),
		messageHash,
	}

	// Hash the typed data
	hash, err := eip712.HashTypedDataV4(domainSeparator, clobAuthTypes, values)
	if err != nil {
		return "", fmt.Errorf("hash typed data: %w", err)
	}

	// Sign with the private key
	sig, err := signer.Sign(privateKey, hash)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	return "0x" + common.Bytes2Hex(sig), nil
}
