package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// FakeAPI starts an HTTPS server with handler and points the provider at it
// for the rest of the test.
func FakeAPI(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	// The provider requires https, so trust the test server's certificate.
	transport := http.DefaultTransport.(*http.Transport)
	previous := transport.TLSClientConfig
	transport.TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig
	t.Cleanup(func() { transport.TLSClientConfig = previous })

	t.Setenv("SEVALLA_API_KEY", "test")
	t.Setenv("SEVALLA_API_URL", server.URL)
	return server
}
