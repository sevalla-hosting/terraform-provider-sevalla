package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetWebhook_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/wh-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		secret := "whsec_abc123"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Webhook{
			ID:            "wh-123",
			Endpoint:      "https://example.com/webhook",
			AllowedEvents: []string{"deployment.created", "deployment.completed"},
			IsEnabled:     true,
			Secret:        &secret,
			CreatedAt:     "2025-01-01T00:00:00.000Z",
			UpdatedAt:     "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	webhook, err := c.GetWebhook(context.Background(), "wh-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if webhook.ID != "wh-123" {
		t.Errorf("expected ID %q, got %q", "wh-123", webhook.ID)
	}
	if webhook.Endpoint != "https://example.com/webhook" {
		t.Errorf("expected Endpoint %q, got %q", "https://example.com/webhook", webhook.Endpoint)
	}
	if len(webhook.AllowedEvents) != 2 {
		t.Errorf("expected 2 allowed_events, got %d", len(webhook.AllowedEvents))
	}
	if !webhook.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}
	if webhook.Secret == nil || *webhook.Secret != "whsec_abc123" {
		t.Errorf("expected secret %q", "whsec_abc123")
	}
}

func TestGetWebhook_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetWebhook(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateWebhook_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/webhooks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateWebhookRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Endpoint != "https://example.com/hooks" {
			t.Errorf("expected Endpoint %q, got %q", "https://example.com/hooks", req.Endpoint)
		}

		secret := "whsec_new"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Webhook{
			ID:            "wh-new",
			Endpoint:      req.Endpoint,
			AllowedEvents: req.AllowedEvents,
			IsEnabled:     true,
			Secret:        &secret,
			CreatedAt:     "2025-01-01T00:00:00.000Z",
			UpdatedAt:     "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	webhook, err := c.CreateWebhook(context.Background(), &CreateWebhookRequest{
		Endpoint:      "https://example.com/hooks",
		AllowedEvents: []string{"deployment.created"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if webhook.ID != "wh-new" {
		t.Errorf("expected ID %q, got %q", "wh-new", webhook.ID)
	}
	if webhook.Endpoint != "https://example.com/hooks" {
		t.Errorf("expected Endpoint %q, got %q", "https://example.com/hooks", webhook.Endpoint)
	}
}

func TestDeleteWebhook_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/webhooks/wh-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteWebhook(context.Background(), "wh-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToggleWebhook_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/webhooks/wh-123/toggle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := c.ToggleWebhook(context.Background(), "wh-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
