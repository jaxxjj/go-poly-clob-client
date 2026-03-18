// Package polyclob provides a Go client for the Polymarket CLOB (Central Limit Order Book) API.
//
// Three authentication levels are supported:
//   - L0 (public): market data, prices, order books
//   - L1 (EIP-712): API key creation/derivation, order signing
//   - L2 (HMAC-SHA256): order posting, cancellation, balance queries
package polyclob

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/jaxxjj/go-poly-clob-client/pkg/cache"
	"github.com/jaxxjj/go-poly-clob-client/pkg/headers"
	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
	"github.com/jaxxjj/go-poly-clob-client/pkg/order"
	"github.com/jaxxjj/go-poly-clob-client/pkg/transport"
	gomodel "github.com/polymarket/go-order-utils/pkg/model"
)

// Client is the Polymarket CLOB client.
type Client struct {
	host       string
	chainID    int64
	privateKey *ecdsa.PrivateKey
	address    common.Address
	creds      *model.ApiCreds
	sigType    int
	funder     common.Address

	http         transport.Doer
	orderBuilder *order.Builder

	tickSizeCache *cache.TTL[model.TickSize]
	negRiskCache  *cache.TTL[bool]
	feeRateCache  *cache.TTL[int]
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) {
		cl.http = transport.NewHTTPClient(c)
	}
}

// WithCreds sets L2 API credentials.
func WithCreds(creds model.ApiCreds) Option {
	return func(cl *Client) {
		cl.creds = &creds
	}
}

// WithSignatureType sets the order signature type (EOA, PolyProxy, GnosisSafe).
func WithSignatureType(st int) Option {
	return func(cl *Client) {
		cl.sigType = st
	}
}

// WithFunder sets the funder address (for proxy wallets).
func WithFunder(addr string) Option {
	return func(cl *Client) {
		cl.funder = common.HexToAddress(addr)
	}
}

// WithTickSizeTTL sets the tick size cache TTL.
func WithTickSizeTTL(d time.Duration) Option {
	return func(cl *Client) {
		cl.tickSizeCache = cache.New[model.TickSize](d)
	}
}

// New creates a public (L0) client. No private key needed.
func New(host string, chainID int64, opts ...Option) *Client {
	cl := &Client{
		host:          strings.TrimRight(host, "/"),
		chainID:       chainID,
		http:          transport.NewHTTPClient(nil),
		tickSizeCache: cache.New[model.TickSize](5 * time.Minute),
		negRiskCache:  cache.New[bool](5 * time.Minute),
		feeRateCache:  cache.New[int](5 * time.Minute),
	}
	for _, opt := range opts {
		opt(cl)
	}
	return cl
}

// NewL1 creates an L1-authenticated client (can sign orders and derive API keys).
func NewL1(host string, chainID int64, privateKeyHex string, opts ...Option) (*Client, error) {
	privateKeyHex = strings.TrimPrefix(strings.ToLower(privateKeyHex), "0x")
	pk, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	cl := New(host, chainID, opts...)
	cl.privateKey = pk
	cl.address = crypto.PubkeyToAddress(pk.PublicKey)
	if cl.funder == (common.Address{}) {
		cl.funder = cl.address
	}
	cl.orderBuilder = order.NewBuilder(pk, chainID, cl.sigType, cl.funder.Hex())
	return cl, nil
}

// NewL2 creates a fully authenticated client (can post orders, query balances, etc.).
func NewL2(host string, chainID int64, privateKeyHex string, creds model.ApiCreds, opts ...Option) (*Client, error) {
	opts = append(opts, WithCreds(creds))
	return NewL1(host, chainID, privateKeyHex, opts...)
}

// SetAPICreds sets L2 credentials after construction (e.g., after deriving them).
func (c *Client) SetAPICreds(creds model.ApiCreds) {
	c.creds = &creds
}

// GetAddress returns the signer's Ethereum address, or empty string if L0.
func (c *Client) GetAddress() string {
	if c.privateKey == nil {
		return ""
	}
	return c.address.Hex()
}

// GetCollateralAddress returns the USDC collateral token address for the configured chain.
func (c *Client) GetCollateralAddress(negRisk bool) (common.Address, error) {
	cfg, err := model.GetContractConfig(c.chainID, negRisk)
	if err != nil {
		return common.Address{}, err
	}
	return cfg.Collateral, nil
}

// GetExchangeAddress returns the exchange contract address.
func (c *Client) GetExchangeAddress(negRisk bool) (common.Address, error) {
	cfg, err := model.GetContractConfig(c.chainID, negRisk)
	if err != nil {
		return common.Address{}, err
	}
	return cfg.Exchange, nil
}

// ---------------------------------------------------------------------------
// L0 — Public endpoints
// ---------------------------------------------------------------------------

// GetOk checks the API health.
func (c *Client) GetOk(ctx context.Context) (string, error) {
	body, err := c.get(ctx, model.EndpointOk, nil)
	return string(body), err
}

// GetServerTime returns the server timestamp.
func (c *Client) GetServerTime(ctx context.Context) (string, error) {
	body, err := c.get(ctx, model.EndpointServerTime, nil)
	return string(body), err
}

// GetMidpoint returns the mid-market price for a token.
func (c *Client) GetMidpoint(ctx context.Context, tokenID string) (json.RawMessage, error) {
	return c.get(ctx, model.EndpointMidpoint+"?token_id="+tokenID, nil)
}

// GetPrice returns the best price for a token and side.
func (c *Client) GetPrice(ctx context.Context, tokenID, side string) (json.RawMessage, error) {
	return c.get(ctx, model.EndpointPrice+"?token_id="+tokenID+"&side="+side, nil)
}

// GetSpread returns the bid-ask spread for a token.
func (c *Client) GetSpread(ctx context.Context, tokenID string) (json.RawMessage, error) {
	return c.get(ctx, model.EndpointSpread+"?token_id="+tokenID, nil)
}

// GetOrderBook returns the full order book for a token.
func (c *Client) GetOrderBook(ctx context.Context, tokenID string) (*model.OrderBookSummary, error) {
	body, err := c.get(ctx, model.EndpointBook+"?token_id="+tokenID, nil)
	if err != nil {
		return nil, err
	}

	var ob model.OrderBookSummary
	if err := json.Unmarshal(body, &ob); err != nil {
		return nil, fmt.Errorf("parse order book: %w", err)
	}

	// Opportunistically cache tick size
	if ob.TickSize != "" {
		c.tickSizeCache.Set(tokenID, model.TickSize(ob.TickSize))
	}
	if ob.NegRisk {
		c.negRiskCache.Set(tokenID, ob.NegRisk)
	}

	return &ob, nil
}

// GetTickSize returns the minimum tick size for a token (cached).
func (c *Client) GetTickSize(ctx context.Context, tokenID string) (model.TickSize, error) {
	if ts, ok := c.tickSizeCache.Get(tokenID); ok {
		return ts, nil
	}

	body, err := c.get(ctx, model.EndpointTickSize+"?token_id="+tokenID, nil)
	if err != nil {
		return "", err
	}

	raw, err := parseJSONField[float64](body, "minimum_tick_size")
	if err != nil {
		return "", fmt.Errorf("parse tick size: %w", err)
	}
	ts := model.TickSize(strconv.FormatFloat(raw, 'f', -1, 64))
	c.tickSizeCache.Set(tokenID, ts)
	return ts, nil
}

// GetNegRisk returns whether a token uses the neg-risk exchange (cached).
func (c *Client) GetNegRisk(ctx context.Context, tokenID string) (bool, error) {
	if nr, ok := c.negRiskCache.Get(tokenID); ok {
		return nr, nil
	}

	body, err := c.get(ctx, model.EndpointNegRisk+"?token_id="+tokenID, nil)
	if err != nil {
		return false, err
	}

	nr, err := parseJSONField[bool](body, "neg_risk")
	if err != nil {
		return false, err
	}
	c.negRiskCache.Set(tokenID, nr)
	return nr, nil
}

// GetFeeRateBps returns the fee rate in basis points for a token (cached).
func (c *Client) GetFeeRateBps(ctx context.Context, tokenID string) (int, error) {
	if fr, ok := c.feeRateCache.Get(tokenID); ok {
		return fr, nil
	}

	body, err := c.get(ctx, model.EndpointFeeRate+"?token_id="+tokenID, nil)
	if err != nil {
		return 0, err
	}

	frFloat, err := parseJSONField[float64](body, "base_fee")
	if err != nil {
		return 0, err
	}
	fr := int(frFloat)
	c.feeRateCache.Set(tokenID, fr)
	return fr, nil
}

// GetMarket returns market details by condition ID.
func (c *Client) GetMarket(ctx context.Context, conditionID string) (json.RawMessage, error) {
	return c.get(ctx, model.EndpointMarkets+"/"+conditionID, nil)
}

// GetMarkets returns paginated market listings.
func (c *Client) GetMarkets(ctx context.Context, cursor string) (json.RawMessage, error) {
	if cursor == "" {
		cursor = model.InitialCursor
	}
	return c.get(ctx, model.EndpointMarkets+"?next_cursor="+cursor, nil)
}

// GetLastTradePrice returns the last traded price for a token.
func (c *Client) GetLastTradePrice(ctx context.Context, tokenID string) (json.RawMessage, error) {
	return c.get(ctx, model.EndpointLastTradePrice+"?token_id="+tokenID, nil)
}

// ---------------------------------------------------------------------------
// L1 — EIP-712 authenticated
// ---------------------------------------------------------------------------

// CreateAPIKey creates new L2 API credentials.
func (c *Client) CreateAPIKey(ctx context.Context, nonce int64) (*model.ApiCreds, error) {
	if err := c.requireL1(); err != nil {
		return nil, err
	}

	h, err := headers.BuildL1(c.privateKey, c.address, c.chainID, nonce)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, "POST", model.EndpointCreateAPIKey, h, nil)
	if err != nil {
		return nil, err
	}

	var creds model.ApiCreds
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("parse api creds: %w", err)
	}
	return &creds, nil
}

// DeriveAPIKey derives existing L2 API credentials (idempotent).
func (c *Client) DeriveAPIKey(ctx context.Context, nonce int64) (*model.ApiCreds, error) {
	if err := c.requireL1(); err != nil {
		return nil, err
	}

	h, err := headers.BuildL1(c.privateKey, c.address, c.chainID, nonce)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, "GET", model.EndpointDeriveAPIKey, h, nil)
	if err != nil {
		return nil, err
	}

	var creds model.ApiCreds
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("parse api creds: %w", err)
	}
	return &creds, nil
}

// CreateOrDeriveAPICreds tries to create API credentials, falling back to derive.
func (c *Client) CreateOrDeriveAPICreds(ctx context.Context, nonce int64) (*model.ApiCreds, error) {
	creds, err := c.CreateAPIKey(ctx, nonce)
	if err == nil {
		return creds, nil
	}
	return c.DeriveAPIKey(ctx, nonce)
}

// CreateOrder creates and signs a limit order (does not post it).
func (c *Client) CreateOrder(ctx context.Context, args model.OrderArgs, opts *model.CreateOrderOptions) (*gomodel.SignedOrder, error) {
	if err := c.requireL1(); err != nil {
		return nil, err
	}

	resolved, err := c.resolveOrderOptions(ctx, args.TokenID, opts)
	if err != nil {
		return nil, err
	}

	// Validate price bounds: [tick_size, 1 - tick_size]
	if err := validatePrice(args.Price, resolved.TickSize); err != nil {
		return nil, err
	}

	if args.FeeRateBps == 0 {
		fr, frErr := c.GetFeeRateBps(ctx, args.TokenID)
		if frErr == nil {
			args.FeeRateBps = fr
		}
	}

	return c.orderBuilder.CreateOrder(args, resolved)
}

// CreateMarketOrder creates and signs a market order.
func (c *Client) CreateMarketOrder(ctx context.Context, args model.MarketOrderArgs, opts *model.CreateOrderOptions) (*gomodel.SignedOrder, error) {
	if err := c.requireL1(); err != nil {
		return nil, err
	}

	// Auto-calculate price from order book if not provided
	if args.Price <= 0 {
		ob, err := c.GetOrderBook(ctx, args.TokenID)
		if err != nil {
			return nil, fmt.Errorf("get order book for price: %w", err)
		}

		ot := args.OrderType
		if ot == "" {
			ot = model.OrderTypeFOK
		}

		if args.Side == "BUY" {
			args.Price, err = order.CalculateBuyMarketPrice(ob.Asks, args.Amount, ot)
		} else {
			args.Price, err = order.CalculateSellMarketPrice(ob.Bids, args.Amount, ot)
		}
		if err != nil {
			return nil, fmt.Errorf("calculate market price: %w", err)
		}
	}

	resolved, err := c.resolveOrderOptions(ctx, args.TokenID, opts)
	if err != nil {
		return nil, err
	}

	// Validate price bounds
	if err := validatePrice(args.Price, resolved.TickSize); err != nil {
		return nil, err
	}

	if args.FeeRateBps == 0 {
		fr, frErr := c.GetFeeRateBps(ctx, args.TokenID)
		if frErr == nil {
			args.FeeRateBps = fr
		}
	}

	return c.orderBuilder.CreateMarketOrder(args, resolved)
}

// ---------------------------------------------------------------------------
// L2 — HMAC authenticated
// ---------------------------------------------------------------------------

// PostOrder submits a signed order to the CLOB.
func (c *Client) PostOrder(ctx context.Context, signed *gomodel.SignedOrder, orderType model.OrderType, postOnly bool) (json.RawMessage, error) {
	if err := c.requireL2(); err != nil {
		return nil, err
	}

	if postOnly && orderType != model.OrderTypeGTC && orderType != model.OrderTypeGTD {
		return nil, fmt.Errorf("post_only orders can only be of type GTC or GTD")
	}

	payload := map[string]any{
		"order":     signedOrderToMap(signed),
		"owner":     c.creds.APIKey,
		"orderType": string(orderType),
		"postOnly":  postOnly,
	}

	return c.postL2(ctx, model.EndpointOrder, payload)
}

// CreateAndPostOrder creates, signs, and posts an order in one call.
func (c *Client) CreateAndPostOrder(ctx context.Context, args model.OrderArgs, opts *model.CreateOrderOptions) (json.RawMessage, error) {
	signed, err := c.CreateOrder(ctx, args, opts)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, signed, model.OrderTypeGTC, false)
}

// PostOrders submits multiple signed orders in a single request.
func (c *Client) PostOrders(ctx context.Context, orders []PostOrdersArg) (json.RawMessage, error) {
	if err := c.requireL2(); err != nil {
		return nil, err
	}

	payload := make([]map[string]any, len(orders))
	for i, arg := range orders {
		payload[i] = map[string]any{
			"order":     signedOrderToMap(arg.Order),
			"owner":     c.creds.APIKey,
			"orderType": string(arg.OrderType),
			"postOnly":  arg.PostOnly,
		}
	}

	return c.postL2(ctx, model.EndpointOrders, payload)
}

// PostOrdersArg holds a single order for batch posting.
type PostOrdersArg struct {
	Order     *gomodel.SignedOrder
	OrderType model.OrderType
	PostOnly  bool
}

// Cancel cancels an order by ID.
func (c *Client) Cancel(ctx context.Context, orderID string) (json.RawMessage, error) {
	return c.deleteL2(ctx, model.EndpointOrder, map[string]any{"orderID": orderID})
}

// CancelOrders cancels multiple orders by ID.
func (c *Client) CancelOrders(ctx context.Context, orderIDs []string) (json.RawMessage, error) {
	return c.deleteL2(ctx, model.EndpointOrders, orderIDs)
}

// CancelAll cancels all open orders.
func (c *Client) CancelAll(ctx context.Context) (json.RawMessage, error) {
	return c.deleteL2(ctx, model.EndpointCancelAll, nil)
}

// CancelMarketOrders cancels all orders for a specific market or asset.
func (c *Client) CancelMarketOrders(ctx context.Context, market, assetID string) (json.RawMessage, error) {
	return c.deleteL2(ctx, model.EndpointCancelMarketOrders, map[string]any{
		"market":   market,
		"asset_id": assetID,
	})
}

// PostHeartbeat sends a heartbeat to prevent auto-cancellation of orders.
func (c *Client) PostHeartbeat(ctx context.Context, heartbeatID string) (json.RawMessage, error) {
	return c.postL2(ctx, model.EndpointHeartbeat, map[string]any{
		"heartbeat_id": heartbeatID,
	})
}

// ClearTickSizeCache clears cached tick sizes. If no keys given, clears all.
func (c *Client) ClearTickSizeCache(tokenIDs ...string) {
	c.tickSizeCache.Clear(tokenIDs...)
}

// GetOrders returns all open orders (paginated, fetches all pages).
func (c *Client) GetOrders(ctx context.Context, params *model.OpenOrderParams) ([]json.RawMessage, error) {
	if err := c.requireL2(); err != nil {
		return nil, err
	}

	path := model.EndpointDataOrders
	if params != nil {
		qs := buildOrderQuery(params)
		if qs != "" {
			path += "?" + qs
		}
	}

	return c.paginateL2(ctx, path)
}

// GetOrder returns a single order by ID.
func (c *Client) GetOrder(ctx context.Context, orderID string) (json.RawMessage, error) {
	if err := c.requireL2(); err != nil {
		return nil, err
	}

	h, err := c.l2Headers("GET", model.EndpointDataOrder+"/"+orderID, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "GET", model.EndpointDataOrder+"/"+orderID, h, nil)
}

// GetTrades returns trade history (paginated, fetches all pages).
func (c *Client) GetTrades(ctx context.Context, params *model.TradeParams) ([]json.RawMessage, error) {
	if err := c.requireL2(); err != nil {
		return nil, err
	}

	path := model.EndpointDataTrades
	if params != nil {
		qs := buildTradeQuery(params)
		if qs != "" {
			path += "?" + qs
		}
	}

	return c.paginateL2(ctx, path)
}

// GetBalanceAllowance returns the balance and allowance for a given asset.
func (c *Client) GetBalanceAllowance(ctx context.Context, params model.BalanceAllowanceParams) (json.RawMessage, error) {
	if err := c.requireL2(); err != nil {
		return nil, err
	}

	sigType := params.SignatureType
	if sigType < 0 {
		sigType = c.sigType
	}

	path := fmt.Sprintf("%s?asset_type=%s&signature_type=%d",
		model.EndpointBalanceAllowance, params.AssetType, sigType)
	if params.TokenID != "" {
		path += "&token_id=" + params.TokenID
	}

	h, err := c.l2Headers("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "GET", path, h, nil)
}

// GetAPIKeys lists all API keys for the authenticated user.
func (c *Client) GetAPIKeys(ctx context.Context) (json.RawMessage, error) {
	if err := c.requireL2(); err != nil {
		return nil, err
	}

	h, err := c.l2Headers("GET", model.EndpointGetAPIKeys, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "GET", model.EndpointGetAPIKeys, h, nil)
}

// DeleteAPIKey deletes the current API key.
func (c *Client) DeleteAPIKey(ctx context.Context) (json.RawMessage, error) {
	return c.deleteL2(ctx, model.EndpointDeleteAPIKey, nil)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (c *Client) requireL1() error {
	if c.privateKey == nil {
		return &model.ErrNotAuthenticated{Level: "L1"}
	}
	return nil
}

func (c *Client) requireL2() error {
	if err := c.requireL1(); err != nil {
		return err
	}
	if c.creds == nil {
		return &model.ErrNotAuthenticated{Level: "L2"}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, h http.Header) ([]byte, error) {
	return c.doRequest(ctx, "GET", path, h, nil)
}

func (c *Client) doRequest(ctx context.Context, method, path string, h http.Header, body []byte) ([]byte, error) {
	url := c.host + path
	return c.http.Do(ctx, method, url, h, body)
}

func (c *Client) l2Headers(method, path string, body []byte) (http.Header, error) {
	return headers.BuildL2(c.address, *c.creds, method, path, body)
}

func (c *Client) postL2(ctx context.Context, path string, payload any) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	h, err := c.l2Headers("POST", path, body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "POST", path, h, body)
}

func (c *Client) deleteL2(ctx context.Context, path string, payload any) (json.RawMessage, error) {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
	}

	h, err := c.l2Headers("DELETE", path, body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "DELETE", path, h, body)
}

func (c *Client) paginateL2(ctx context.Context, basePath string) ([]json.RawMessage, error) {
	var all []json.RawMessage
	cursor := model.InitialCursor

	for {
		sep := "?"
		if strings.Contains(basePath, "?") {
			sep = "&"
		}
		path := basePath + sep + "next_cursor=" + cursor

		h, err := c.l2Headers("GET", path, nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(ctx, "GET", path, h, nil)
		if err != nil {
			return nil, err
		}

		var page struct {
			Data       []json.RawMessage `json:"data"`
			NextCursor string            `json:"next_cursor"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("parse page: %w", err)
		}

		all = append(all, page.Data...)

		if page.NextCursor == "" || page.NextCursor == model.EndCursor {
			break
		}
		cursor = page.NextCursor
	}

	return all, nil
}

func (c *Client) resolveOrderOptions(ctx context.Context, tokenID string, opts *model.CreateOrderOptions) (model.CreateOrderOptions, error) {
	var resolved model.CreateOrderOptions

	if opts != nil {
		resolved = *opts
	}

	if resolved.TickSize == "" {
		ts, err := c.GetTickSize(ctx, tokenID)
		if err != nil {
			return resolved, fmt.Errorf("get tick size: %w", err)
		}
		resolved.TickSize = ts
	}

	if !resolved.NegRisk {
		nr, err := c.GetNegRisk(ctx, tokenID)
		if err != nil {
			// Non-fatal: default to false
			nr = false
		}
		resolved.NegRisk = nr
	}

	return resolved, nil
}

func parseJSONField[T any](body []byte, key string) (T, error) {
	var obj map[string]json.RawMessage
	var zero T
	if err := json.Unmarshal(body, &obj); err != nil {
		return zero, fmt.Errorf("parse response: %w", err)
	}
	raw, ok := obj[key]
	if !ok {
		return zero, fmt.Errorf("missing field %q in response", key)
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return zero, fmt.Errorf("parse field %q: %w", key, err)
	}
	return v, nil
}

func validatePrice(price float64, ts model.TickSize) error {
	tsFloat, err := strconv.ParseFloat(string(ts), 64)
	if err != nil || tsFloat == 0 {
		return nil // unknown tick size, skip validation
	}
	minPrice := tsFloat
	maxPrice := 1.0 - tsFloat
	if price < minPrice || price > maxPrice {
		return fmt.Errorf("price %.4f out of valid range [%.4f, %.4f] for tick size %s",
			price, minPrice, maxPrice, ts)
	}
	return nil
}

// signedOrderToMap serializes a SignedOrder to the JSON format expected by the CLOB API.
//
// Field types must match py-order-utils SignedOrder.dict() exactly:
//   - salt: integer
//   - side: string "BUY"/"SELL" (converted from 0/1)
//   - signatureType: integer
//   - makerAmount, takerAmount, tokenId, expiration, nonce, feeRateBps: string
func signedOrderToMap(s *gomodel.SignedOrder) map[string]any {
	side := "BUY"
	if s.Side.Int64() == 1 {
		side = "SELL"
	}

	return map[string]any{
		"salt":          s.Salt.Int64(),
		"maker":         s.Maker.Hex(),
		"signer":        s.Signer.Hex(),
		"taker":         s.Taker.Hex(),
		"tokenId":       s.TokenId.String(),
		"makerAmount":   s.MakerAmount.String(),
		"takerAmount":   s.TakerAmount.String(),
		"expiration":    s.Expiration.String(),
		"nonce":         s.Nonce.String(),
		"feeRateBps":    s.FeeRateBps.String(),
		"side":          side,
		"signatureType": int(s.SignatureType.Int64()),
		"signature":     "0x" + common.Bytes2Hex(s.Signature),
	}
}

func buildOrderQuery(p *model.OpenOrderParams) string {
	var parts []string
	if p.ID != "" {
		parts = append(parts, "id="+p.ID)
	}
	if p.Market != "" {
		parts = append(parts, "market="+p.Market)
	}
	if p.AssetID != "" {
		parts = append(parts, "asset_id="+p.AssetID)
	}
	return strings.Join(parts, "&")
}

func buildTradeQuery(p *model.TradeParams) string {
	var parts []string
	if p.ID != "" {
		parts = append(parts, "id="+p.ID)
	}
	if p.MakerAddress != "" {
		parts = append(parts, "maker_address="+p.MakerAddress)
	}
	if p.Market != "" {
		parts = append(parts, "market="+p.Market)
	}
	if p.AssetID != "" {
		parts = append(parts, "asset_id="+p.AssetID)
	}
	if p.Before > 0 {
		parts = append(parts, fmt.Sprintf("before=%d", p.Before))
	}
	if p.After > 0 {
		parts = append(parts, fmt.Sprintf("after=%d", p.After))
	}
	return strings.Join(parts, "&")
}
