package polyclob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

const testPK = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNew_L0(t *testing.T) {
	cl := New("https://clob.polymarket.com", 137)
	if cl.GetAddress() != "" {
		t.Error("L0 client should have empty address")
	}
}

func TestNewL1(t *testing.T) {
	cl, err := NewL1("https://clob.polymarket.com", 80002, testPK)
	if err != nil {
		t.Fatalf("NewL1: %v", err)
	}
	if cl.GetAddress() == "" {
		t.Error("L1 client should have non-empty address")
	}
	if cl.GetAddress() != "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" {
		t.Errorf("unexpected address: %s", cl.GetAddress())
	}
}

func TestNewL2(t *testing.T) {
	creds := model.ApiCreds{APIKey: "k", APISecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", Passphrase: "p"}
	cl, err := NewL2("https://clob.polymarket.com", 80002, testPK, creds)
	if err != nil {
		t.Fatalf("NewL2: %v", err)
	}
	if cl.creds == nil {
		t.Error("L2 client should have creds")
	}
}

func TestRequireL1_Fails_L0(t *testing.T) {
	cl := New("http://localhost", 137)
	err := cl.requireL1()
	if err == nil {
		t.Error("L0 client should fail requireL1")
	}
}

func TestRequireL2_Fails_L1(t *testing.T) {
	cl, _ := NewL1("http://localhost", 80002, testPK)
	err := cl.requireL2()
	if err == nil {
		t.Error("L1 client (no creds) should fail requireL2")
	}
}

func TestGetOk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`"OK"`))
	}))
	defer server.Close()

	cl := New(server.URL, 137)
	resp, err := cl.GetOk(context.Background())
	if err != nil {
		t.Fatalf("GetOk: %v", err)
	}
	if resp != `"OK"` {
		t.Errorf("resp = %q", resp)
	}
}

func TestGetOrderBook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token_id") != "123" {
			t.Errorf("token_id = %q", r.URL.Query().Get("token_id"))
		}
		ob := model.OrderBookSummary{
			Market:   "0xcond",
			AssetID:  "123",
			TickSize: "0.01",
			NegRisk:  true,
			Bids:     []model.OrderLevel{{Price: "0.50", Size: "100"}},
			Asks:     []model.OrderLevel{{Price: "0.55", Size: "200"}},
		}
		data, _ := json.Marshal(ob)
		w.WriteHeader(200)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	cl := New(server.URL, 137)
	ob, err := cl.GetOrderBook(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetOrderBook: %v", err)
	}
	if ob.Market != "0xcond" {
		t.Errorf("market = %q", ob.Market)
	}
	if len(ob.Bids) != 1 || ob.Bids[0].Price != "0.50" {
		t.Error("bids mismatch")
	}

	// Verify tick size was cached
	ts, ok := cl.tickSizeCache.Get("123")
	if !ok || ts != model.TickSize001 {
		t.Errorf("tick size not cached: %v, %v", ts, ok)
	}

	// Verify neg_risk was cached
	nr, ok := cl.negRiskCache.Get("123")
	if !ok || !nr {
		t.Error("neg_risk not cached")
	}
}

func TestGetTickSize(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"minimum_tick_size":0.001}`))
	}))
	defer server.Close()

	cl := New(server.URL, 137)
	ctx := context.Background()

	ts, err := cl.GetTickSize(ctx, "token1")
	if err != nil {
		t.Fatalf("GetTickSize: %v", err)
	}
	if ts != model.TickSize0001 {
		t.Errorf("tick size = %q, want 0.001", ts)
	}

	// Second call should hit cache
	ts2, _ := cl.GetTickSize(ctx, "token1")
	if ts2 != ts {
		t.Error("cache miss on second call")
	}
	if calls != 1 {
		t.Errorf("expected 1 server call (cached), got %d", calls)
	}
}

func TestGetNegRisk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"neg_risk":true}`))
	}))
	defer server.Close()

	cl := New(server.URL, 137)
	nr, err := cl.GetNegRisk(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetNegRisk: %v", err)
	}
	if !nr {
		t.Error("expected true")
	}
}

func TestGetContractAddresses(t *testing.T) {
	cl := New("http://localhost", 137)

	ex, err := cl.GetExchangeAddress(false)
	if err != nil {
		t.Fatalf("GetExchangeAddress: %v", err)
	}
	if ex != model.PolygonConfig.Exchange {
		t.Errorf("exchange = %s, want %s", ex.Hex(), model.PolygonConfig.Exchange.Hex())
	}

	exNR, err := cl.GetExchangeAddress(true)
	if err != nil {
		t.Fatalf("GetExchangeAddress(negRisk): %v", err)
	}
	if exNR != model.PolygonNegRiskConfig.Exchange {
		t.Errorf("neg_risk exchange = %s, want %s", exNR.Hex(), model.PolygonNegRiskConfig.Exchange.Hex())
	}

	col, err := cl.GetCollateralAddress(false)
	if err != nil {
		t.Fatalf("GetCollateralAddress: %v", err)
	}
	if col != model.PolygonConfig.Collateral {
		t.Errorf("collateral = %s", col.Hex())
	}
}

func TestGetContractAddresses_InvalidChain(t *testing.T) {
	cl := New("http://localhost", 99999)

	_, err := cl.GetExchangeAddress(false)
	if err == nil {
		t.Error("expected error for unsupported chain ID")
	}
}

func TestCreateOrder_RequiresL1(t *testing.T) {
	cl := New("http://localhost", 137)
	_, err := cl.CreateOrder(context.Background(), model.OrderArgs{}, nil)
	if err == nil {
		t.Error("expected L1 error from L0 client")
	}
}

func TestPostOrder_RequiresL2(t *testing.T) {
	cl, _ := NewL1("http://localhost", 80002, testPK)
	_, err := cl.PostOrder(context.Background(), nil, model.OrderTypeGTC, false)
	if err == nil {
		t.Error("expected L2 error from L1 client")
	}
}

func TestSetAPICreds(t *testing.T) {
	cl, _ := NewL1("http://localhost", 80002, testPK)

	// Initially L2 should fail
	if cl.requireL2() == nil {
		t.Error("expected L2 to fail before SetAPICreds")
	}

	cl.SetAPICreds(model.ApiCreds{
		APIKey:     "key",
		APISecret:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		Passphrase: "pass",
	})

	if cl.requireL2() != nil {
		t.Error("expected L2 to succeed after SetAPICreds")
	}
}

func TestPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		cursor := r.URL.Query().Get("next_cursor")

		var resp []byte
		if cursor == model.InitialCursor || cursor == "" {
			resp, _ = json.Marshal(map[string]any{
				"data":        []map[string]string{{"id": "order1"}, {"id": "order2"}},
				"next_cursor": "page2",
			})
		} else {
			resp, _ = json.Marshal(map[string]any{
				"data":        []map[string]string{{"id": "order3"}},
				"next_cursor": model.EndCursor,
			})
		}
		w.WriteHeader(200)
		_, _ = w.Write(resp)
	}))
	defer server.Close()

	creds := model.ApiCreds{APIKey: "k", APISecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", Passphrase: "p"}
	cl, _ := NewL2(server.URL, 80002, testPK, creds)

	orders, err := cl.GetOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("expected 3 orders (2 pages), got %d", len(orders))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (pagination), got %d", callCount)
	}
}

func TestSignedOrderToMap(t *testing.T) {
	// Create a real signed order to test serialization
	cl, _ := NewL1("http://localhost", 80002, testPK)

	args := model.OrderArgs{
		TokenID: "1234",
		Price:   0.50,
		Size:    100,
		Side:    "BUY",
	}
	opts := &model.CreateOrderOptions{TickSize: model.TickSize001}

	signed, err := cl.CreateOrder(context.Background(), args, opts)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	m := signedOrderToMap(signed)

	requiredFields := []string{"salt", "maker", "signer", "taker", "tokenId",
		"makerAmount", "takerAmount", "expiration", "nonce", "feeRateBps",
		"side", "signatureType", "signature"}

	for _, f := range requiredFields {
		if _, ok := m[f]; !ok {
			t.Errorf("missing field %q in order map", f)
		}
	}

	sig := m["signature"].(string)
	if len(sig) < 4 || sig[:2] != "0x" {
		t.Errorf("signature should start with 0x, got %q", sig[:10])
	}

	t.Logf("Order map: tokenId=%s maker=%s side=%s", m["tokenId"], m["maker"], m["side"])
}

func TestWithFunder(t *testing.T) {
	funder := "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	cl, _ := NewL1("http://localhost", 80002, testPK, WithFunder(funder))

	if cl.funder.Hex() != funder {
		t.Errorf("funder = %s, want %s", cl.funder.Hex(), funder)
	}
}

func TestValidatePrice(t *testing.T) {
	tests := []struct {
		price    float64
		tickSize model.TickSize
		wantErr  bool
	}{
		{0.50, model.TickSize001, false},     // valid
		{0.01, model.TickSize001, false},     // min valid
		{0.99, model.TickSize001, false},     // max valid
		{0.005, model.TickSize001, true},     // below min
		{0.995, model.TickSize001, true},     // above max
		{0.1, model.TickSize01, false},       // min for 0.1 tick
		{0.9, model.TickSize01, false},       // max for 0.1 tick
		{0.05, model.TickSize01, true},       // below min for 0.1 tick
		{0.0001, model.TickSize00001, false}, // min for 0.0001 tick
		{0.9999, model.TickSize00001, false}, // max for 0.0001 tick
	}

	for _, tt := range tests {
		err := validatePrice(tt.price, tt.tickSize)
		if (err != nil) != tt.wantErr {
			t.Errorf("validatePrice(%.4f, %s) error=%v, wantErr=%v", tt.price, tt.tickSize, err, tt.wantErr)
		}
	}
}

func TestClearTickSizeCache(t *testing.T) {
	cl := New("http://localhost", 137)
	cl.tickSizeCache.Set("tok1", model.TickSize001)
	cl.tickSizeCache.Set("tok2", model.TickSize0001)

	cl.ClearTickSizeCache("tok1")
	if _, ok := cl.tickSizeCache.Get("tok1"); ok {
		t.Error("tok1 should be cleared")
	}
	if _, ok := cl.tickSizeCache.Get("tok2"); !ok {
		t.Error("tok2 should still exist")
	}

	cl.ClearTickSizeCache()
	if _, ok := cl.tickSizeCache.Get("tok2"); ok {
		t.Error("tok2 should be cleared after ClearAll")
	}
}

func TestCreateOrder_PriceValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Tick size and neg_risk endpoints
		if r.URL.Path == "/tick-size" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"minimum_tick_size":0.01}`))
			return
		}
		if r.URL.Path == "/neg-risk" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
			return
		}
		if r.URL.Path == "/fee-rate" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"base_fee":0}`))
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cl, _ := NewL1(server.URL, 80002, testPK)

	// Price below tick_size should fail
	_, err := cl.CreateOrder(context.Background(), model.OrderArgs{
		TokenID: "123",
		Price:   0.005, // below 0.01
		Size:    100,
		Side:    "BUY",
	}, nil)
	if err == nil {
		t.Error("expected price validation error for price=0.005")
	} else {
		t.Logf("Expected error: %v", err)
	}

	// Price above 1-tick_size should fail
	_, err = cl.CreateOrder(context.Background(), model.OrderArgs{
		TokenID: "123",
		Price:   0.995, // above 0.99
		Size:    100,
		Side:    "BUY",
	}, nil)
	if err == nil {
		t.Error("expected price validation error for price=0.995")
	}
}
