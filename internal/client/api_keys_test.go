package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetAPIKey_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api-keys/key-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"key-123","name":"My API Key","is_enabled":true,"masked_token":"sevall...f456",
			"roles":[{"id":"role-1","name":"Admin","description":"Full access"}],
			"capabilities":[{"id":"cap-1","permission":"APP:READ","resource_id":"app-1"}],
			"created_at":"2025-01-01T00:00:00.000Z","updated_at":"2025-01-01T00:00:00.000Z"}`))
	})
	defer server.Close()

	apiKey, err := c.GetAPIKey(context.Background(), "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey.ID != "key-123" {
		t.Errorf("expected ID %q, got %q", "key-123", apiKey.ID)
	}
	if apiKey.Name != "My API Key" {
		t.Errorf("expected name %q, got %q", "My API Key", apiKey.Name)
	}
	if !apiKey.Enabled {
		t.Error("expected Enabled to be true")
	}
	if len(apiKey.Roles) != 1 || apiKey.Roles[0].ID != "role-1" {
		t.Errorf("unexpected roles: %+v", apiKey.Roles)
	}
	if len(apiKey.Capabilities) != 1 || apiKey.Capabilities[0].ResourceID == nil || *apiKey.Capabilities[0].ResourceID != "app-1" {
		t.Errorf("unexpected capabilities: %+v", apiKey.Capabilities)
	}
}

func TestGetAPIKey_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetAPIKey(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateAPIKey_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api-keys" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateAPIKeyRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Name != "Terraform Key" {
			t.Errorf("expected name %q, got %q", "Terraform Key", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreatedAPIKey{
			ID:    "key-new",
			Name:  req.Name,
			Token: "sevalla_new_token",
		})
	})
	defer server.Close()

	apiKey, err := c.CreateAPIKey(context.Background(), &CreateAPIKeyRequest{
		Name: "Terraform Key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey.ID != "key-new" {
		t.Errorf("expected ID %q, got %q", "key-new", apiKey.ID)
	}
	if apiKey.Name != "Terraform Key" {
		t.Errorf("expected name %q, got %q", "Terraform Key", apiKey.Name)
	}
	if apiKey.Token != "sevalla_new_token" {
		t.Errorf("expected token %q", "sevalla_new_token")
	}
}

func TestDeleteAPIKey_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api-keys/key-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteAPIKey(context.Background(), "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToggleAPIKey_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api-keys/key-123/toggle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := c.ToggleAPIKey(context.Background(), "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
