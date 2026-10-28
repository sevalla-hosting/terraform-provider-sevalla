package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL = "https://api.sevalla.com/v3"
	DefaultTimeout = 30 * time.Second
)

type SevallaClient struct {
	BaseURL    string
	apiKey     string
	HTTPClient *http.Client
	UserAgent  string
}

type Option func(*SevallaClient)

func WithBaseURL(url string) Option {
	return func(c *SevallaClient) {
		c.BaseURL = url
	}
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *SevallaClient) {
		c.HTTPClient = hc
	}
}

func WithUserAgent(ua string) Option {
	return func(c *SevallaClient) {
		c.UserAgent = ua
	}
}

func NewClient(apiKey string, opts ...Option) *SevallaClient {
	c := &SevallaClient{
		BaseURL:   DefaultBaseURL,
		apiKey:    apiKey,
		UserAgent: "terraform-provider-sevalla/dev",
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.HTTPClient == nil {
		baseURL := c.BaseURL
		c.HTTPClient = &http.Client{
			Timeout:   DefaultTimeout,
			Transport: NewTransport(c.apiKey, c.UserAgent),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				// The transport sets Authorization on every request, so never
				// follow a redirect to another host: it would leak the API key.
				if len(via) > 0 && req.URL.Host != via[0].URL.Host {
					return http.ErrUseLastResponse
				}
				// Prevent redirects away from the configured base URL scheme.
				if req.URL.Scheme != "https" && baseURL == DefaultBaseURL {
					return http.ErrUseLastResponse
				}
				return nil
			},
		}
	}

	return c
}

// path formats an API path, escaping every argument as a single path segment.
func path(format string, ids ...string) string {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = url.PathEscape(id)
	}
	return fmt.Sprintf(format, args...)
}

// do sends a JSON request to the API and decodes a JSON response into out.
// A nil in sends no body; a nil out discards the response body. Any 2xx status
// is a success.
func (c *SevallaClient) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshaling %s %s request: %w", method, path, err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return parseErrorResponse(resp)
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding %s %s response: %w", method, path, err)
	}
	return nil
}
