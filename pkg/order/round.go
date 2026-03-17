// Package order implements order amount calculation and building for the Polymarket CLOB.
package order

import (
	"math"
	"strconv"
	"strings"
)

// RoundDown truncates x to n decimal places (floor).
func RoundDown(x float64, n int) float64 {
	mult := math.Pow(10, float64(n))
	return math.Floor(x*mult) / mult
}

// RoundNormal rounds x to n decimal places (half-up).
func RoundNormal(x float64, n int) float64 {
	mult := math.Pow(10, float64(n))
	return math.Round(x*mult) / mult
}

// RoundUp rounds x up to n decimal places (ceil).
func RoundUp(x float64, n int) float64 {
	mult := math.Pow(10, float64(n))
	return math.Ceil(x*mult) / mult
}

// ToTokenDecimals converts a human-readable amount to 6-decimal token units.
// e.g., 1.5 → 1500000
func ToTokenDecimals(x float64) int {
	f := x * 1e6
	if DecimalPlaces(f) > 0 {
		f = RoundNormal(f, 0)
	}
	return int(f)
}

// DecimalPlaces returns the number of decimal places in a float's string representation.
func DecimalPlaces(x float64) int {
	s := strconv.FormatFloat(x, 'f', -1, 64)
	idx := strings.IndexByte(s, '.')
	if idx == -1 {
		return 0
	}
	return len(s) - idx - 1
}
