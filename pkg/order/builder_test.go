package order

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

const testPrivateKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestBuilder_CreateOrder_BUY(t *testing.T) {
	pk, _ := crypto.HexToECDSA(testPrivateKeyHex)
	b := NewBuilder(pk, 80002, model.EOA, "")

	args := model.OrderArgs{
		TokenID: "1234",
		Price:   0.50,
		Size:    100,
		Side:    "BUY",
	}
	opts := model.CreateOrderOptions{
		TickSize: model.TickSize001,
		NegRisk:  false,
	}

	signed, err := b.CreateOrder(args, opts)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	if signed == nil {
		t.Fatal("expected non-nil SignedOrder")
	}
	if len(signed.Signature) != 65 {
		t.Errorf("signature length = %d, want 65", len(signed.Signature))
	}
	if signed.TokenId == nil || signed.TokenId.Int64() != 1234 {
		t.Errorf("tokenId = %v, want 1234", signed.TokenId)
	}

	t.Logf("SignedOrder: maker=%s taker=%s makerAmt=%s takerAmt=%s",
		signed.Maker.Hex(), signed.Taker.Hex(),
		signed.MakerAmount.String(), signed.TakerAmount.String())
}

func TestBuilder_CreateOrder_SELL(t *testing.T) {
	pk, _ := crypto.HexToECDSA(testPrivateKeyHex)
	b := NewBuilder(pk, 80002, model.EOA, "")

	args := model.OrderArgs{
		TokenID: "5678",
		Price:   0.75,
		Size:    200,
		Side:    "SELL",
	}
	opts := model.CreateOrderOptions{
		TickSize: model.TickSize001,
		NegRisk:  false,
	}

	signed, err := b.CreateOrder(args, opts)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	if signed.Side.Int64() != 1 { // SELL
		t.Errorf("side = %d, want 1 (SELL)", signed.Side.Int64())
	}
}

func TestBuilder_CreateOrder_NegRisk(t *testing.T) {
	pk, _ := crypto.HexToECDSA(testPrivateKeyHex)
	b := NewBuilder(pk, 80002, model.EOA, "")

	args := model.OrderArgs{
		TokenID: "9999",
		Price:   0.30,
		Size:    50,
		Side:    "BUY",
	}
	opts := model.CreateOrderOptions{
		TickSize: model.TickSize001,
		NegRisk:  true,
	}

	signed, err := b.CreateOrder(args, opts)
	if err != nil {
		t.Fatalf("CreateOrder (neg_risk): %v", err)
	}
	if signed == nil {
		t.Fatal("expected non-nil SignedOrder for neg_risk")
	}
}

func TestBuilder_CreateOrder_WithFunder(t *testing.T) {
	pk, _ := crypto.HexToECDSA(testPrivateKeyHex)
	funder := "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	b := NewBuilder(pk, 80002, model.PolyProxy, funder)

	args := model.OrderArgs{
		TokenID: "1234",
		Price:   0.50,
		Size:    100,
		Side:    "BUY",
	}
	opts := model.CreateOrderOptions{TickSize: model.TickSize001}

	signed, err := b.CreateOrder(args, opts)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	if signed.Maker.Hex() != funder {
		t.Errorf("maker = %s, want %s (funder)", signed.Maker.Hex(), funder)
	}
	// Signer should be the private key's address, not the funder
	signerAddr := crypto.PubkeyToAddress(pk.PublicKey)
	if signed.Signer != signerAddr {
		t.Errorf("signer = %s, want %s", signed.Signer.Hex(), signerAddr.Hex())
	}
}

func TestBuilder_CreateMarketOrder(t *testing.T) {
	pk, _ := crypto.HexToECDSA(testPrivateKeyHex)
	b := NewBuilder(pk, 80002, model.EOA, "")

	args := model.MarketOrderArgs{
		TokenID: "1234",
		Amount:  50,
		Side:    "BUY",
		Price:   0.50,
	}
	opts := model.CreateOrderOptions{TickSize: model.TickSize001}

	signed, err := b.CreateMarketOrder(args, opts)
	if err != nil {
		t.Fatalf("CreateMarketOrder: %v", err)
	}
	if signed == nil {
		t.Fatal("expected non-nil SignedOrder")
	}

	// Expiration should be 0 for market orders
	if signed.Expiration.Int64() != 0 {
		t.Errorf("expiration = %d, want 0", signed.Expiration.Int64())
	}
}

func TestCalculateBuyMarketPrice(t *testing.T) {
	asks := []model.OrderLevel{
		{Price: "0.60", Size: "100"}, // best ask
		{Price: "0.55", Size: "100"},
		{Price: "0.50", Size: "100"},
	}

	// $30 should be fillable at 0.50 (50 * 0.50 = $25, need more → 100 * 0.55 = $55 cumulative)
	price, err := CalculateBuyMarketPrice(asks, 30, model.OrderTypeGTC)
	if err != nil {
		t.Fatalf("CalculateBuyMarketPrice: %v", err)
	}
	t.Logf("Buy market price for $30: %.2f", price)
	if price <= 0 {
		t.Error("expected positive price")
	}
}

func TestCalculateBuyMarketPrice_EmptyBook(t *testing.T) {
	_, err := CalculateBuyMarketPrice(nil, 100, model.OrderTypeFOK)
	if err == nil {
		t.Error("expected error for empty asks")
	}
}

func TestCalculateSellMarketPrice(t *testing.T) {
	bids := []model.OrderLevel{
		{Price: "0.50", Size: "100"}, // best bid
		{Price: "0.45", Size: "200"},
		{Price: "0.40", Size: "300"},
	}

	price, err := CalculateSellMarketPrice(bids, 150, model.OrderTypeGTC)
	if err != nil {
		t.Fatalf("CalculateSellMarketPrice: %v", err)
	}
	t.Logf("Sell market price for 150 shares: %.2f", price)
	if price <= 0 {
		t.Error("expected positive price")
	}
}

func TestCalculateSellMarketPrice_FOK_InsufficientLiquidity(t *testing.T) {
	bids := []model.OrderLevel{
		{Price: "0.50", Size: "10"},
	}

	_, err := CalculateSellMarketPrice(bids, 1000, model.OrderTypeFOK)
	if err == nil {
		t.Error("expected error for FOK with insufficient liquidity")
	}
}
