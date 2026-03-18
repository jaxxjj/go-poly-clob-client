package polyclob

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

// Integration tests hit the real Polymarket CLOB API.
// Skipped unless POLYMARKET_PRIVATE_KEY is set.
//
//	go test -v -run TestIntegration -count=1

func newIntegrationClient(t *testing.T) *Client {
	t.Helper()
	pk := os.Getenv("POLYMARKET_PRIVATE_KEY")
	if pk == "" {
		t.Skip("POLYMARKET_PRIVATE_KEY not set")
	}

	sigType := 0 // EOA
	if os.Getenv("POLYMARKET_SIGNATURE_TYPE") == "1" {
		sigType = model.PolyProxy
	}

	var opts []Option
	if sigType > 0 {
		opts = append(opts, WithSignatureType(sigType))
	}
	if funder := os.Getenv("POLYMARKET_FUNDER"); funder != "" {
		opts = append(opts, WithFunder(funder))
	}

	client, err := NewL1("https://clob.polymarket.com", 137, pk, opts...)
	if err != nil {
		t.Fatalf("NewL1: %v", err)
	}

	// Auto-derive L2 creds
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	creds, err := client.CreateOrDeriveAPICreds(ctx, 0)
	if err != nil {
		t.Fatalf("CreateOrDeriveAPICreds: %v", err)
	}
	client.SetAPICreds(*creds)
	t.Logf("address=%s apiKey=%s", client.GetAddress(), creds.APIKey)

	return client
}

// activeMarket returns a condition ID and token ID from a currently active market
// using the Gamma Data API (which returns active markets first).
func activeMarket(t *testing.T) (conditionID, tokenID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://gamma-api.polymarket.com/markets?closed=false&active=true&limit=1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("gamma API unavailable: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var markets []struct {
		ConditionID  string `json:"conditionId"`
		ClobTokenIDs string `json:"clobTokenIds"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil || len(markets) == 0 {
		t.Skip("no active market found via gamma API")
	}

	// clobTokenIds is a JSON-encoded string array like "[\"tok1\",\"tok2\"]"
	var tokenIDs []string
	_ = json.Unmarshal([]byte(markets[0].ClobTokenIDs), &tokenIDs)
	if len(tokenIDs) == 0 {
		t.Skip("no token IDs in market")
	}

	return markets[0].ConditionID, tokenIDs[0]
}

// ---------------------------------------------------------------------------
// L0 — Public endpoints
// ---------------------------------------------------------------------------

func TestIntegration_GetOk(t *testing.T) {
	client := newIntegrationClient(t)
	resp, err := client.GetOk(context.Background())
	if err != nil {
		t.Fatalf("GetOk: %v", err)
	}
	t.Logf("OK: %s", resp)
}

func TestIntegration_GetServerTime(t *testing.T) {
	client := newIntegrationClient(t)
	resp, err := client.GetServerTime(context.Background())
	if err != nil {
		t.Fatalf("GetServerTime: %v", err)
	}
	t.Logf("ServerTime: %s", resp)
}

func TestIntegration_GetMarket(t *testing.T) {
	client := newIntegrationClient(t)
	conditionID, _ := activeMarket(t)

	resp, err := client.GetMarket(context.Background(), conditionID)
	if err != nil {
		t.Fatalf("GetMarket: %v", err)
	}
	if len(resp) > 200 {
		t.Logf("Market: %s...", string(resp)[:200])
	} else {
		t.Logf("Market: %s", string(resp))
	}
}

func TestIntegration_GetTickSize(t *testing.T) {
	client := newIntegrationClient(t)
	_, tokenID := activeMarket(t)

	ts, err := client.GetTickSize(context.Background(), tokenID)
	if err != nil {
		t.Fatalf("GetTickSize: %v", err)
	}
	t.Logf("TickSize: %s", ts)

	if ts != model.TickSize01 && ts != model.TickSize001 && ts != model.TickSize0001 && ts != model.TickSize00001 {
		t.Errorf("unexpected tick size: %q", ts)
	}
}

func TestIntegration_GetNegRisk(t *testing.T) {
	client := newIntegrationClient(t)
	_, tokenID := activeMarket(t)

	nr, err := client.GetNegRisk(context.Background(), tokenID)
	if err != nil {
		t.Fatalf("GetNegRisk: %v", err)
	}
	t.Logf("NegRisk: %v", nr)
}

func TestIntegration_GetFeeRate(t *testing.T) {
	client := newIntegrationClient(t)
	_, tokenID := activeMarket(t)

	fr, err := client.GetFeeRateBps(context.Background(), tokenID)
	if err != nil {
		t.Fatalf("GetFeeRateBps: %v", err)
	}
	t.Logf("FeeRateBps: %d", fr)
}

func TestIntegration_GetOrderBook(t *testing.T) {
	client := newIntegrationClient(t)
	_, tokenID := activeMarket(t)

	ob, err := client.GetOrderBook(context.Background(), tokenID)
	if err != nil {
		t.Fatalf("GetOrderBook: %v", err)
	}
	t.Logf("OrderBook: %d bids, %d asks, tickSize=%s", len(ob.Bids), len(ob.Asks), ob.TickSize)
}

func TestIntegration_GetMidpoint(t *testing.T) {
	client := newIntegrationClient(t)
	_, tokenID := activeMarket(t)

	raw, err := client.GetMidpoint(context.Background(), tokenID)
	if err != nil {
		t.Fatalf("GetMidpoint: %v", err)
	}
	t.Logf("Midpoint: %s", string(raw))
}

func TestIntegration_GetPrice(t *testing.T) {
	client := newIntegrationClient(t)
	_, tokenID := activeMarket(t)

	for _, side := range []string{"BUY", "SELL"} {
		raw, err := client.GetPrice(context.Background(), tokenID, side)
		if err != nil {
			t.Errorf("GetPrice(%s): %v", side, err)
			continue
		}
		t.Logf("Price(%s): %s", side, string(raw))
	}
}

// ---------------------------------------------------------------------------
// L1 — EIP-712 authenticated
// ---------------------------------------------------------------------------

func TestIntegration_DeriveAPIKey(t *testing.T) {
	pk := os.Getenv("POLYMARKET_PRIVATE_KEY")
	if pk == "" {
		t.Skip("POLYMARKET_PRIVATE_KEY not set")
	}

	client, err := NewL1("https://clob.polymarket.com", 137, pk)
	if err != nil {
		t.Fatalf("NewL1: %v", err)
	}

	creds, err := client.CreateOrDeriveAPICreds(context.Background(), 0)
	if err != nil {
		t.Fatalf("CreateOrDeriveAPICreds: %v", err)
	}
	t.Logf("APIKey=%s Passphrase=%s Secret=%s...", creds.APIKey, creds.Passphrase, creds.APISecret[:10])
}

func TestIntegration_CreateOrder(t *testing.T) {
	client := newIntegrationClient(t)
	_, tokenID := activeMarket(t)

	signed, err := client.CreateOrder(context.Background(), model.OrderArgs{
		TokenID: tokenID,
		Price:   0.01,
		Size:    5,
		Side:    "BUY",
	}, nil)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Logf("Signed: maker=%s tokenId=%s makerAmt=%s takerAmt=%s",
		signed.Maker.Hex(), signed.TokenId.String(),
		signed.MakerAmount.String(), signed.TakerAmount.String())
}

// ---------------------------------------------------------------------------
// L2 — HMAC authenticated
// ---------------------------------------------------------------------------

func TestIntegration_GetBalance(t *testing.T) {
	client := newIntegrationClient(t)

	raw, err := client.GetBalanceAllowance(context.Background(), model.BalanceAllowanceParams{
		AssetType:     model.AssetCollateral,
		SignatureType: -1,
	})
	if err != nil {
		t.Fatalf("GetBalanceAllowance: %v", err)
	}
	t.Logf("Balance: %s", string(raw))
}

func TestIntegration_GetOpenOrders(t *testing.T) {
	client := newIntegrationClient(t)

	orders, err := client.GetOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	t.Logf("Open orders: %d", len(orders))
}

func TestIntegration_GetTrades(t *testing.T) {
	client := newIntegrationClient(t)

	trades, err := client.GetTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetTrades: %v", err)
	}
	t.Logf("Trades: %d", len(trades))
}

// TestIntegration_PlaceAndCancel creates an order at $0.01 (won't fill) and cancels it.
// Requires USDC deposited on Polymarket.
func TestIntegration_PlaceAndCancel(t *testing.T) {
	client := newIntegrationClient(t)
	ctx := context.Background()

	// Check balance
	raw, err := client.GetBalanceAllowance(ctx, model.BalanceAllowanceParams{
		AssetType:     model.AssetCollateral,
		SignatureType: -1,
	})
	if err != nil {
		t.Fatalf("GetBalanceAllowance: %v", err)
	}
	var bal struct {
		Balance string `json:"balance"`
	}
	_ = json.Unmarshal(raw, &bal)
	t.Logf("Balance: %s", bal.Balance)
	if bal.Balance == "" || bal.Balance == "0" {
		t.Skip("no USDC balance — skipping order test")
	}

	_, tokenID := activeMarket(t)

	// Create + post
	signed, err := client.CreateOrder(ctx, model.OrderArgs{
		TokenID: tokenID,
		Price:   0.01,
		Size:    5,
		Side:    "BUY",
	}, nil)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	resp, err := client.PostOrder(ctx, signed, model.OrderTypeGTC, false)
	if err != nil {
		t.Fatalf("PostOrder: %v", err)
	}

	var result struct {
		OrderID string `json:"orderID"`
	}
	_ = json.Unmarshal(resp, &result)
	t.Logf("Posted: orderID=%s", result.OrderID)

	if result.OrderID == "" {
		t.Fatal("empty order ID")
	}

	time.Sleep(1 * time.Second)

	// Cancel
	_, err = client.Cancel(ctx, result.OrderID)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	t.Log("Cancelled successfully")
}
