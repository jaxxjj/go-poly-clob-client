package order

import (
	"testing"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

func TestGetOrderAmounts_BUY(t *testing.T) {
	tests := []struct {
		name      string
		size      float64
		price     float64
		tickSize  model.TickSize
		wantMaker int
		wantTaker int
	}{
		{
			name:      "tick=0.01, price=0.50, size=100",
			size:      100,
			price:     0.50,
			tickSize:  model.TickSize001,
			wantMaker: 50000000,  // 50 USDC
			wantTaker: 100000000, // 100 shares
		},
		{
			name:      "tick=0.01, price=0.99, size=10",
			size:      10,
			price:     0.99,
			tickSize:  model.TickSize001,
			wantMaker: 9900000, // 9.9 USDC
			wantTaker: 10000000,
		},
		{
			name:      "tick=0.1, price=0.5, size=50",
			size:      50,
			price:     0.5,
			tickSize:  model.TickSize01,
			wantMaker: 25000000,
			wantTaker: 50000000,
		},
		{
			name:      "tick=0.001, price=0.123, size=100",
			size:      100,
			price:     0.123,
			tickSize:  model.TickSize0001,
			wantMaker: 12300000,
			wantTaker: 100000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := model.RoundConfigForTickSize(tt.tickSize)
			side, maker, taker, err := GetOrderAmounts("BUY", tt.size, tt.price, rc)
			if err != nil {
				t.Fatalf("GetOrderAmounts: %v", err)
			}
			if side != model.Buy {
				t.Errorf("side = %d, want Buy(%d)", side, model.Buy)
			}
			if maker != tt.wantMaker {
				t.Errorf("maker = %d, want %d", maker, tt.wantMaker)
			}
			if taker != tt.wantTaker {
				t.Errorf("taker = %d, want %d", taker, tt.wantTaker)
			}
		})
	}
}

func TestGetOrderAmounts_SELL(t *testing.T) {
	tests := []struct {
		name      string
		size      float64
		price     float64
		tickSize  model.TickSize
		wantMaker int
		wantTaker int
	}{
		{
			name:      "tick=0.01, price=0.50, size=100",
			size:      100,
			price:     0.50,
			tickSize:  model.TickSize001,
			wantMaker: 100000000, // 100 shares
			wantTaker: 50000000,  // 50 USDC
		},
		{
			name:      "tick=0.01, price=0.01, size=10",
			size:      10,
			price:     0.01,
			tickSize:  model.TickSize001,
			wantMaker: 10000000,
			wantTaker: 100000, // 0.1 USDC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := model.RoundConfigForTickSize(tt.tickSize)
			side, maker, taker, err := GetOrderAmounts("SELL", tt.size, tt.price, rc)
			if err != nil {
				t.Fatalf("GetOrderAmounts: %v", err)
			}
			if side != model.Sell {
				t.Errorf("side = %d, want Sell(%d)", side, model.Sell)
			}
			if maker != tt.wantMaker {
				t.Errorf("maker = %d, want %d", maker, tt.wantMaker)
			}
			if taker != tt.wantTaker {
				t.Errorf("taker = %d, want %d", taker, tt.wantTaker)
			}
		})
	}
}

func TestGetOrderAmounts_InvalidSide(t *testing.T) {
	rc := model.RoundConfigForTickSize(model.TickSize001)
	_, _, _, err := GetOrderAmounts("INVALID", 100, 0.5, rc)
	if err == nil {
		t.Error("expected error for invalid side")
	}
}

func TestGetMarketOrderAmounts_BUY(t *testing.T) {
	// BUY: amount = USDC to spend, taker = shares received
	rc := model.RoundConfigForTickSize(model.TickSize001)
	side, maker, taker, err := GetMarketOrderAmounts("BUY", 50, 0.50, rc)
	if err != nil {
		t.Fatalf("GetMarketOrderAmounts: %v", err)
	}
	if side != model.Buy {
		t.Errorf("side = %d, want Buy", side)
	}
	if maker != 50000000 { // 50 USDC
		t.Errorf("maker = %d, want 50000000", maker)
	}
	if taker != 100000000 { // 100 shares (50 / 0.5)
		t.Errorf("taker = %d, want 100000000", taker)
	}
}

func TestGetMarketOrderAmounts_SELL(t *testing.T) {
	// SELL: amount = shares to sell, taker = USDC received
	rc := model.RoundConfigForTickSize(model.TickSize001)
	side, maker, taker, err := GetMarketOrderAmounts("SELL", 100, 0.50, rc)
	if err != nil {
		t.Fatalf("GetMarketOrderAmounts: %v", err)
	}
	if side != model.Sell {
		t.Errorf("side = %d, want Sell", side)
	}
	if maker != 100000000 { // 100 shares
		t.Errorf("maker = %d, want 100000000", maker)
	}
	if taker != 50000000 { // 50 USDC (100 * 0.5)
		t.Errorf("taker = %d, want 50000000", taker)
	}
}

func TestAdjustPrecision(t *testing.T) {
	// Value that has too many decimal places should get rounded
	val := 949.9970999999999
	result := adjustPrecision(val, 4)
	if DecimalPlaces(result) > 4 {
		t.Errorf("adjustPrecision(%f, 4) = %f, still has %d decimal places", val, result, DecimalPlaces(result))
	}
}
