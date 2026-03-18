package order

import "testing"

func TestRoundDown(t *testing.T) {
	tests := []struct {
		x    float64
		n    int
		want float64
	}{
		{1.2345, 2, 1.23},
		{1.999, 2, 1.99},
		{1.001, 2, 1.00},
		{100.0, 2, 100.0},
		{0.12345, 4, 0.1234},
		{0.5, 0, 0},
		{1.9, 0, 1},
	}
	for _, tt := range tests {
		got := RoundDown(tt.x, tt.n)
		if got != tt.want {
			t.Errorf("RoundDown(%f, %d) = %f, want %f", tt.x, tt.n, got, tt.want)
		}
	}
}

func TestRoundNormal(t *testing.T) {
	tests := []struct {
		x    float64
		n    int
		want float64
	}{
		{1.2345, 2, 1.23},
		{1.235, 2, 1.24}, // rounds up at .5
		{1.999, 2, 2.00},
		{0.50, 2, 0.50},
		{0.005, 2, 0.01}, // 0.5 rounds up
	}
	for _, tt := range tests {
		got := RoundNormal(tt.x, tt.n)
		if got != tt.want {
			t.Errorf("RoundNormal(%f, %d) = %f, want %f", tt.x, tt.n, got, tt.want)
		}
	}
}

func TestRoundUp(t *testing.T) {
	tests := []struct {
		x    float64
		n    int
		want float64
	}{
		{1.2301, 2, 1.24},
		{1.20, 2, 1.20},
		{1.001, 2, 1.01},
		{0.0001, 2, 0.01},
	}
	for _, tt := range tests {
		got := RoundUp(tt.x, tt.n)
		if got != tt.want {
			t.Errorf("RoundUp(%f, %d) = %f, want %f", tt.x, tt.n, got, tt.want)
		}
	}
}

func TestToTokenDecimals(t *testing.T) {
	tests := []struct {
		x    float64
		want int
	}{
		{1.0, 1000000},
		{0.5, 500000},
		{100.0, 100000000},
		{0.000001, 1},
		{1.5, 1500000},
		{0.01, 10000},
	}
	for _, tt := range tests {
		got := ToTokenDecimals(tt.x)
		if got != tt.want {
			t.Errorf("ToTokenDecimals(%f) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

func TestDecimalPlaces(t *testing.T) {
	tests := []struct {
		x    float64
		want int
	}{
		{100.0, 0},
		{1.5, 1},
		{1.23, 2},
		{0.001, 3},
		{0.0001, 4},
		{949.9970999999999, 13},
		{1.0, 0},
	}
	for _, tt := range tests {
		got := DecimalPlaces(tt.x)
		if got != tt.want {
			t.Errorf("DecimalPlaces(%v) = %d, want %d", tt.x, got, tt.want)
		}
	}
}
