// Package model defines shared data types for the Polymarket CLOB client.
package model

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// ApiCreds holds L2 API credentials derived from L1 EIP-712 authentication.
type ApiCreds struct {
	APIKey     string `json:"apiKey"`
	APISecret  string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// OrderType specifies the time-in-force behavior for an order.
type OrderType string

const (
	OrderTypeGTC OrderType = "GTC" // Good-Till-Cancelled
	OrderTypeFOK OrderType = "FOK" // Fill-or-Kill
	OrderTypeGTD OrderType = "GTD" // Good-Till-Date
	OrderTypeFAK OrderType = "FAK" // Fill-and-Kill
)

// TickSize represents the minimum price increment for a market.
type TickSize string

const (
	TickSize01    TickSize = "0.1"
	TickSize001   TickSize = "0.01"
	TickSize0001  TickSize = "0.001"
	TickSize00001 TickSize = "0.0001"
)

// SignatureType identifies the signing method used for orders.
type SignatureType = int

const (
	EOA            SignatureType = 0 // Standard externally owned account
	PolyProxy      SignatureType = 1 // Polymarket proxy wallet
	PolyGnosisSafe SignatureType = 2 // Gnosis Safe multisig
)

// Side represents the order direction.
type Side = int

const (
	Buy  Side = 0
	Sell Side = 1
)

// ZeroAddress is the Ethereum zero address used for public (taker-less) orders.
var ZeroAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")

// OrderArgs specifies the parameters for creating a limit order.
type OrderArgs struct {
	TokenID    string
	Price      float64
	Size       float64
	Side       string // "BUY" | "SELL"
	FeeRateBps int
	Nonce      int
	Expiration int
	Taker      string // defaults to ZeroAddress
}

// MarketOrderArgs specifies the parameters for creating a market order.
type MarketOrderArgs struct {
	TokenID    string
	Amount     float64 // BUY: USDC amount, SELL: shares to sell
	Side       string  // "BUY" | "SELL"
	Price      float64 // optional; 0 = auto-calculate from order book
	FeeRateBps int
	Nonce      int
	Taker      string
	OrderType  OrderType // default FOK
}

// CreateOrderOptions provides optional overrides for order creation.
type CreateOrderOptions struct {
	TickSize TickSize
	NegRisk  bool
}

// OrderLevel represents a single price level in an order book.
type OrderLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// OrderBookSummary is the full order book snapshot for a token.
type OrderBookSummary struct {
	Market         string       `json:"market"`
	AssetID        string       `json:"asset_id"`
	Timestamp      string       `json:"timestamp"`
	Bids           []OrderLevel `json:"bids"`
	Asks           []OrderLevel `json:"asks"`
	MinOrderSize   string       `json:"min_order_size"`
	NegRisk        bool         `json:"neg_risk"`
	TickSize       string       `json:"tick_size"`
	LastTradePrice string       `json:"last_trade_price"`
	Hash           string       `json:"hash"`
}

// AssetType identifies the type of asset for balance queries.
type AssetType string

const (
	AssetCollateral  AssetType = "COLLATERAL"
	AssetConditional AssetType = "CONDITIONAL"
)

// BalanceAllowanceParams specifies filters for balance/allowance queries.
type BalanceAllowanceParams struct {
	AssetType     AssetType
	TokenID       string
	SignatureType int
}

// OpenOrderParams specifies filters for querying open orders.
type OpenOrderParams struct {
	ID      string
	Market  string
	AssetID string
}

// TradeParams specifies filters for querying trade history.
type TradeParams struct {
	ID           string
	MakerAddress string
	Market       string
	AssetID      string
	Before       int
	After        int
}

// RoundConfig defines decimal precision for price, size, and amount.
type RoundConfig struct {
	Price  int
	Size   int
	Amount int
}

// RoundConfigForTickSize returns the rounding configuration for a given tick size.
func RoundConfigForTickSize(ts TickSize) RoundConfig {
	switch ts {
	case TickSize01:
		return RoundConfig{Price: 1, Size: 2, Amount: 3}
	case TickSize001:
		return RoundConfig{Price: 2, Size: 2, Amount: 4}
	case TickSize0001:
		return RoundConfig{Price: 3, Size: 2, Amount: 5}
	case TickSize00001:
		return RoundConfig{Price: 4, Size: 2, Amount: 6}
	default:
		return RoundConfig{Price: 2, Size: 2, Amount: 4}
	}
}

// ContractConfig holds the exchange contract addresses for a specific chain and market type.
type ContractConfig struct {
	Exchange          common.Address
	Collateral        common.Address
	ConditionalTokens common.Address
}

// Chain IDs.
const (
	ChainPolygon = 137
	ChainAmoy    = 80002
)

// Contract addresses per chain and market type.
var (
	PolygonConfig = ContractConfig{
		Exchange:          common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"),
		Collateral:        common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"),
		ConditionalTokens: common.HexToAddress("0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"),
	}
	PolygonNegRiskConfig = ContractConfig{
		Exchange:          common.HexToAddress("0xC5d563A36AE78145C45a50134d48A1215220f80a"),
		Collateral:        common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"),
		ConditionalTokens: common.HexToAddress("0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"),
	}
	AmoyConfig = ContractConfig{
		Exchange:          common.HexToAddress("0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40"),
		Collateral:        common.HexToAddress("0x9c4e1703476e875070ee25b56a58b008cfb8fa78"),
		ConditionalTokens: common.HexToAddress("0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB"),
	}
	AmoyNegRiskConfig = ContractConfig{
		Exchange:          common.HexToAddress("0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"),
		Collateral:        common.HexToAddress("0x9c4e1703476e875070ee25b56a58b008cfb8fa78"),
		ConditionalTokens: common.HexToAddress("0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB"),
	}
)

// GetContractConfig returns the contract config for a given chain and neg_risk flag.
// Returns an error for unsupported chain IDs.
func GetContractConfig(chainID int64, negRisk bool) (ContractConfig, error) {
	switch chainID {
	case ChainPolygon:
		if negRisk {
			return PolygonNegRiskConfig, nil
		}
		return PolygonConfig, nil
	case ChainAmoy:
		if negRisk {
			return AmoyNegRiskConfig, nil
		}
		return AmoyConfig, nil
	default:
		return ContractConfig{}, fmt.Errorf("unsupported chain ID: %d", chainID)
	}
}
