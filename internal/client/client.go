package client

import (
	"fmt"
	"net/http"
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
				// Strip Authorization header on cross-origin redirects
				// to prevent leaking the API key.
				if len(via) > 0 && req.URL.Host != via[0].URL.Host {
					req.Header.Del("Authorization")
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

func (c *SevallaClient) url(path string, args ...interface{}) string {
	if len(args) > 0 {
		path = fmt.Sprintf(path, args...)
	}
	return c.BaseURL + path
}
