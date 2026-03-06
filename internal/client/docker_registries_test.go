package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetDockerRegistry_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docker-registries/reg-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		username := "docker-user"
		registry := "dockerHub"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DockerRegistry{
			ID:        "reg-123",
			Name:      "My Registry",
			Registry:  &registry,
			Username:  &username,
			CreatedAt: "2025-01-01T00:00:00.000Z",
			UpdatedAt: "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	registry, err := c.GetDockerRegistry(context.Background(), "reg-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if registry.ID != "reg-123" {
		t.Errorf("expected ID %q, got %q", "reg-123", registry.ID)
	}
	if registry.Name != "My Registry" {
		t.Errorf("expected name %q, got %q", "My Registry", registry.Name)
	}
	if registry.Registry == nil || *registry.Registry != "dockerHub" {
		t.Errorf("expected registry %q", "dockerHub")
	}
	if registry.Username == nil || *registry.Username != "docker-user" {
		t.Errorf("expected username %q", "docker-user")
	}
}

func TestGetDockerRegistry_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetDockerRegistry(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateDockerRegistry_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/docker-registries" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateDockerRegistryRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Name != "New Registry" {
			t.Errorf("expected name %q, got %q", "New Registry", req.Name)
		}
		if req.Secret != "mysecret" {
			t.Errorf("expected secret %q, got %q", "mysecret", req.Secret)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(DockerRegistry{
			ID:        "reg-new",
			Name:      req.Name,
			Registry:  req.Registry,
			Username:  &req.Username,
			CreatedAt: "2025-01-01T00:00:00.000Z",
			UpdatedAt: "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	registryType := "github"
	registry, err := c.CreateDockerRegistry(context.Background(), &CreateDockerRegistryRequest{
		Name:     "New Registry",
		Username: "ghcr-user",
		Secret:   "mysecret",
		Registry: &registryType,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if registry.ID != "reg-new" {
		t.Errorf("expected ID %q, got %q", "reg-new", registry.ID)
	}
	if registry.Registry == nil || *registry.Registry != "github" {
		t.Errorf("expected registry %q", "github")
	}
}

func TestDeleteDockerRegistry_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/docker-registries/reg-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteDockerRegistry(context.Background(), "reg-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
