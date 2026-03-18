# go-poly-clob-client

Go client for the [Polymarket CLOB API](https://docs.polymarket.com). Provides market data, order creation, signing, and submission with full L0/L1/L2 authentication support.

## Install

```bash
go get github.com/jaxxjj/go-poly-clob-client
```

## Quick Start

### Read Market Data (no auth)

```go
client := polyclob.New("https://clob.polymarket.com", 137)

ob, _ := client.GetOrderBook(ctx, tokenID)
mid, _ := client.GetMidpoint(ctx, tokenID)
market, _ := client.GetMarket(ctx, conditionID)
```

### Place an Order

```go
// Create L1 client (can sign orders)
client, _ := polyclob.NewL1("https://clob.polymarket.com", 137, privateKeyHex)

// Derive L2 credentials (one-time)
creds, _ := client.CreateOrDeriveAPICreds(ctx, 0)
client.SetAPICreds(*creds)

// Create and post order
resp, _ := client.CreateAndPostOrder(ctx, model.OrderArgs{
    TokenID: tokenID,
    Price:   0.50,
    Size:    100,
    Side:    "BUY",
}, nil) // auto-resolves tick_size and neg_risk
```

## Authentication Levels

| Level | What you need | What you can do |
|-------|--------------|-----------------|
| **L0** | Host URL only | Market data, prices, order books |
| **L1** | + Private key | Sign orders, derive API keys |
| **L2** | + API credentials | Post orders, cancel, query balances |

```go
// L0
client := polyclob.New(host, chainID)

// L1
client, _ := polyclob.NewL1(host, chainID, privateKeyHex)

// L2
client, _ := polyclob.NewL2(host, chainID, privateKeyHex, creds)

// Or derive creds from L1 → L2
client, _ := polyclob.NewL1(host, chainID, privateKeyHex)
creds, _ := client.CreateOrDeriveAPICreds(ctx, 0)
client.SetAPICreds(*creds)
```

## API Reference

### Market Data (L0)

```go
GetOk(ctx) (string, error)
GetServerTime(ctx) (string, error)
GetOrderBook(ctx, tokenID) (*OrderBookSummary, error)
GetMidpoint(ctx, tokenID) (json.RawMessage, error)
GetPrice(ctx, tokenID, side) (json.RawMessage, error)
GetSpread(ctx, tokenID) (json.RawMessage, error)
GetTickSize(ctx, tokenID) (TickSize, error)        // cached
GetNegRisk(ctx, tokenID) (bool, error)             // cached
GetFeeRateBps(ctx, tokenID) (int, error)           // cached
GetMarket(ctx, conditionID) (json.RawMessage, error)
GetMarkets(ctx, cursor) (json.RawMessage, error)
GetLastTradePrice(ctx, tokenID) (json.RawMessage, error)
```

### Order Creation (L1)

```go
CreateOrder(ctx, OrderArgs, *CreateOrderOptions) (*SignedOrder, error)
CreateMarketOrder(ctx, MarketOrderArgs, *CreateOrderOptions) (*SignedOrder, error)
CreateAPIKey(ctx, nonce) (*ApiCreds, error)
DeriveAPIKey(ctx, nonce) (*ApiCreds, error)
CreateOrDeriveAPICreds(ctx, nonce) (*ApiCreds, error)
```

### Trading (L2)

```go
PostOrder(ctx, *SignedOrder, OrderType, postOnly) (json.RawMessage, error)
PostOrders(ctx, []PostOrdersArg) (json.RawMessage, error)
CreateAndPostOrder(ctx, OrderArgs, *CreateOrderOptions) (json.RawMessage, error)
Cancel(ctx, orderID) (json.RawMessage, error)
CancelOrders(ctx, []string) (json.RawMessage, error)
CancelAll(ctx) (json.RawMessage, error)
CancelMarketOrders(ctx, market, assetID) (json.RawMessage, error)
GetOrders(ctx, *OpenOrderParams) ([]json.RawMessage, error)
GetOrder(ctx, orderID) (json.RawMessage, error)
GetTrades(ctx, *TradeParams) ([]json.RawMessage, error)
GetBalanceAllowance(ctx, BalanceAllowanceParams) (json.RawMessage, error)
PostHeartbeat(ctx, heartbeatID) (json.RawMessage, error)
```

## Configuration Options

```go
polyclob.NewL1(host, chainID, privateKeyHex,
    polyclob.WithSignatureType(model.PolyProxy),     // proxy wallet (default for Polymarket UI wallets)
    polyclob.WithFunder("0x..."),                     // funder address (if different from signer)
    polyclob.WithHTTPClient(customHTTPClient),        // custom http.Client
    polyclob.WithTickSizeTTL(10 * time.Minute),       // tick size cache duration
)
```

> Most wallets created via the Polymarket UI use `PolyProxy` (signature type 1). Use `EOA` (0) only for raw externally-owned accounts.

## Architecture

```
polyclob (root)         Client + public API
  pkg/auth              L1 EIP-712 + L2 HMAC signing
  pkg/headers           HTTP header builders
  pkg/order             Amount calculation + order builder (wraps go-order-utils)
  pkg/transport         HTTP Doer interface
  pkg/cache             Generic TTL cache
  pkg/model             Types, endpoints, errors, contract config
```

Dependencies:
- [`go-order-utils`](https://github.com/polymarket/go-order-utils) — EIP-712 order signing
- [`go-ethereum`](https://github.com/ethereum/go-ethereum) — ECDSA, ABI encoding

## Networks

| Network | Chain ID | CLOB URL |
|---------|----------|----------|
| Polygon | 137 | `https://clob.polymarket.com` |
| Amoy (testnet) | 80002 | `https://clob.polymarket.com` |

## Testing

```bash
# Unit tests (no credentials needed)
make test

# Integration tests (hit real CLOB API)
POLYMARKET_PRIVATE_KEY=0x... go test -v -run TestIntegration -count=1
```

## Examples

See [`examples/`](./examples/) for runnable examples:
- [`market_data`](./examples/market_data/) — fetch prices and order books
- [`place_order`](./examples/place_order/) — create, sign, and post an order
- [`cancel_orders`](./examples/cancel_orders/) — cancel orders and check trades
