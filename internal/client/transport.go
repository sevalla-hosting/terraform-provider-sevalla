package client

import (
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const maxRetryAfter = 120 // seconds

type transport struct {
	apiKey    string
	userAgent string
	base      http.RoundTripper
}

func NewTransport(apiKey, userAgent string) http.RoundTripper {
	return &transport{
		apiKey:    apiKey,
		userAgent: userAgent,
		base:      http.DefaultTransport,
	}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	if req.Body != nil && req.Body != http.NoBody {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", t.userAgent)

	// Only retry on safe (idempotent) methods for 5xx errors.
	// 429 (rate limit) is always safe to retry since the server didn't process the request.
	safeMethod := req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodOptions

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			jitter := time.Duration(rand.Int63n(int64(time.Second)))

			if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
						if seconds > maxRetryAfter {
							seconds = maxRetryAfter
						}
						backoff = time.Duration(seconds) * time.Second
					}
				}
			}

			// Re-read request body for retry if available.
			if req.GetBody != nil {
				req.Body, _ = req.GetBody()
			}

			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(backoff + jitter):
			}
		}

		resp, err = t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		shouldRetry := resp.StatusCode == http.StatusTooManyRequests ||
			(resp.StatusCode >= 500 && safeMethod)
		if !shouldRetry {
			break
		}

		// Drain and close body before retrying to avoid resource leak.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	return resp, nil
}
