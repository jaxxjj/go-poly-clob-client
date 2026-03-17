package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

func TestHTTPClient_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "go-poly-clob-client" {
			t.Error("missing User-Agent header")
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := NewHTTPClient(nil)
	body, err := c.Do(context.Background(), "GET", server.URL+"/test", nil, nil)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q, want {\"ok\":true}", string(body))
	}
}

func TestHTTPClient_PostWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type for POST with body")
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	c := NewHTTPClient(nil)
	body, err := c.Do(context.Background(), "POST", server.URL+"/order", nil, []byte(`{"size":10}`))
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if string(body) != `{"id":"123"}` {
		t.Errorf("body = %q", string(body))
	}
}

func TestHTTPClient_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("POLY_API_KEY") != "my-key" {
			t.Errorf("POLY_API_KEY = %q, want my-key", r.Header.Get("POLY_API_KEY"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer server.Close()

	c := NewHTTPClient(nil)
	h := make(http.Header)
	h.Set("POLY_API_KEY", "my-key")
	_, err := c.Do(context.Background(), "GET", server.URL+"/", h, nil)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestHTTPClient_4xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	c := NewHTTPClient(nil)
	_, err := c.Do(context.Background(), "GET", server.URL+"/bad", nil, nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}

	apiErr, ok := err.(*model.PolyAPIError)
	if !ok {
		t.Fatalf("expected *model.PolyAPIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("status = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Body != `{"error":"bad request"}` {
		t.Errorf("body = %q", apiErr.Body)
	}
}

func TestHTTPClient_5xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`internal server error`))
	}))
	defer server.Close()

	c := NewHTTPClient(nil)
	_, err := c.Do(context.Background(), "GET", server.URL+"/fail", nil, nil)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	apiErr, ok := err.(*model.PolyAPIError)
	if !ok {
		t.Fatalf("expected *model.PolyAPIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("status = %d, want 500", apiErr.StatusCode)
	}
}

func TestHTTPClient_DeleteMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`deleted`))
	}))
	defer server.Close()

	c := NewHTTPClient(nil)
	body, err := c.Do(context.Background(), "DELETE", server.URL+"/order", nil, []byte(`{"orderID":"x"}`))
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if string(body) != "deleted" {
		t.Errorf("body = %q", string(body))
	}
}
