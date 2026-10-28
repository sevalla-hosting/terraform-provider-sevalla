package client

import (
	"context"
	"net/http"
	"testing"
)

func TestDo_AcceptsAny2xxAndEscapesIDs(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/object-storage/a%2F..%2Fb" {
			t.Errorf("unexpected path: %s", r.URL.EscapedPath())
		}
		w.WriteHeader(http.StatusAccepted)
	})
	defer server.Close()

	if err := c.DeleteObjectStorage(context.Background(), "a/../b"); err != nil {
		t.Fatalf("expected 202 to be a success, got %v", err)
	}
}
