// Package transport provides the HTTP layer for the Polymarket CLOB API.
package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

// Doer is the minimal HTTP interface needed by the CLOB client.
type Doer interface {
	Do(ctx context.Context, method, url string, headers http.Header, body []byte) ([]byte, error)
}

// HTTPClient wraps *http.Client and implements Doer.
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates an HTTPClient with the given options.
func NewHTTPClient(client *http.Client) *HTTPClient {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &HTTPClient{client: client}
}

// Do executes an HTTP request and returns the response body.
func (c *HTTPClient) Do(ctx context.Context, method, url string, headers http.Header, body []byte) ([]byte, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "go-poly-clob-client")
	req.Header.Set("Accept", "*/*")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	// Merge caller-provided headers (auth headers, etc.)
	for k, vals := range headers {
		for _, v := range vals {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &model.PolyAPIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
			Endpoint:   method + " " + url,
		}
	}

	return respBody, nil
}
