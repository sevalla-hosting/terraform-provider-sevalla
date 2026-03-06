package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestTransport_AuthHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key-123" {
			t.Errorf("expected Authorization %q, got %q", "Bearer test-key-123", got)
		}
		// Content-Type only set when request has a body
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Errorf("expected no Content-Type for GET request, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("expected Accept %q, got %q", "application/json", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-agent/1.0" {
			t.Errorf("expected User-Agent %q, got %q", "test-agent/1.0", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tr := NewTransport("test-key-123", "test-agent/1.0")
	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestTransport_RetryOn429(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tr := NewTransport("key", "agent")
	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Errorf("expected at least 2 attempts, got %d", got)
	}
}

func TestTransport_RetryOn500(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tr := NewTransport("key", "agent")
	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Errorf("expected at least 2 attempts, got %d", got)
	}
}

func TestTransport_NoRetryOnClientError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	tr := NewTransport("key", "agent")
	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected exactly 1 attempt, got %d", got)
	}
}

func TestTransport_NoRetryPostOn500(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tr := NewTransport("key", "agent")
	req, err := http.NewRequest(http.MethodPost, server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected exactly 1 attempt for POST on 500, got %d", got)
	}
}

func TestTransport_RespectsRetryAfter(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tr := NewTransport("key", "agent")
	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Errorf("expected at least 2 attempts, got %d", got)
	}
}

func TestTransport_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	tr := NewTransport("key", "agent")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	_, err = tr.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}
