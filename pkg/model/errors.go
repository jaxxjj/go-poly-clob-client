package model

import "fmt"

// PolyAPIError represents an HTTP error from the Polymarket CLOB API.
type PolyAPIError struct {
	StatusCode int
	Body       string
	Endpoint   string
}

func (e *PolyAPIError) Error() string {
	return fmt.Sprintf("polymarket API error: status=%d endpoint=%s body=%s", e.StatusCode, e.Endpoint, e.Body)
}

// ErrNotAuthenticated is returned when an L1/L2 method is called without credentials.
type ErrNotAuthenticated struct {
	Level string // "L1" or "L2"
}

func (e *ErrNotAuthenticated) Error() string {
	return fmt.Sprintf("client not authenticated at %s level", e.Level)
}
