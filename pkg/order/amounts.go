package order

import (
	"fmt"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

// GetOrderAmounts calculates the maker and taker amounts for a limit order.
//
// For BUY: maker pays USDC (size × price), taker receives shares (size).
// For SELL: maker pays shares (size), taker receives USDC (size × price).
//
// Returns (side, makerAmount, takerAmount) in 6-decimal token units.
func GetOrderAmounts(side string, size, price float64, rc model.RoundConfig) (model.Side, int, int, error) {
	rawPrice := RoundNormal(price, rc.Price)

	switch side {
	case "BUY":
		rawTakerAmt := RoundDown(size, rc.Size)
		rawMakerAmt := rawTakerAmt * rawPrice

		rawMakerAmt = adjustPrecision(rawMakerAmt, rc.Amount)

		return model.Buy, ToTokenDecimals(rawMakerAmt), ToTokenDecimals(rawTakerAmt), nil

	case "SELL":
		rawMakerAmt := RoundDown(size, rc.Size)
		rawTakerAmt := rawMakerAmt * rawPrice

		rawTakerAmt = adjustPrecision(rawTakerAmt, rc.Amount)

		return model.Sell, ToTokenDecimals(rawMakerAmt), ToTokenDecimals(rawTakerAmt), nil

	default:
		return 0, 0, 0, fmt.Errorf("invalid side: %q, must be BUY or SELL", side)
	}
}

// GetMarketOrderAmounts calculates the maker and taker amounts for a market order.
//
// For BUY: amount is USDC to spend; taker receives shares (amount / price).
// For SELL: amount is shares to sell; taker receives USDC (amount × price).
func GetMarketOrderAmounts(side string, amount, price float64, rc model.RoundConfig) (model.Side, int, int, error) {
	rawPrice := RoundNormal(price, rc.Price)

	switch side {
	case "BUY":
		rawMakerAmt := RoundDown(amount, rc.Size)
		rawTakerAmt := rawMakerAmt / rawPrice

		rawTakerAmt = adjustPrecision(rawTakerAmt, rc.Amount)

		return model.Buy, ToTokenDecimals(rawMakerAmt), ToTokenDecimals(rawTakerAmt), nil

	case "SELL":
		rawMakerAmt := RoundDown(amount, rc.Size)
		rawTakerAmt := rawMakerAmt * rawPrice

		rawTakerAmt = adjustPrecision(rawTakerAmt, rc.Amount)

		return model.Sell, ToTokenDecimals(rawMakerAmt), ToTokenDecimals(rawTakerAmt), nil

	default:
		return 0, 0, 0, fmt.Errorf("invalid side: %q, must be BUY or SELL", side)
	}
}

// adjustPrecision applies the py-clob-client rounding heuristic:
// if too many decimal places, try round_up(amount+4), if still too many, round_down(amount).
func adjustPrecision(val float64, maxDecimals int) float64 {
	if DecimalPlaces(val) > maxDecimals {
		val = RoundUp(val, maxDecimals+4)
		if DecimalPlaces(val) > maxDecimals {
			val = RoundDown(val, maxDecimals)
		}
	}
	return val
}
