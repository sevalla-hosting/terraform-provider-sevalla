package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetStaticSite_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static-sites/ss-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StaticSite{
			ID:          "ss-123",
			Name:        "my-site",
			DisplayName: "My Site",
			Source:      "publicGit",
			AutoDeploy:  true,
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	site, err := c.GetStaticSite(context.Background(), "ss-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if site.ID != "ss-123" {
		t.Errorf("expected ID %q, got %q", "ss-123", site.ID)
	}
	if site.DisplayName != "My Site" {
		t.Errorf("expected display name %q, got %q", "My Site", site.DisplayName)
	}
	if !site.AutoDeploy {
		t.Error("expected AutoDeploy to be true")
	}
}

func TestGetStaticSite_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetStaticSite(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateStaticSite_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/static-sites" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateStaticSiteRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "New Site" {
			t.Errorf("expected display name %q, got %q", "New Site", req.DisplayName)
		}

		source := "publicGit"
		if req.Source != nil {
			source = *req.Source
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(StaticSite{
			ID:          "ss-new",
			Name:        "new-site",
			DisplayName: req.DisplayName,
			Source:      source,
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	source := "publicGit"
	site, err := c.CreateStaticSite(context.Background(), &CreateStaticSiteRequest{
		DisplayName:   "New Site",
		RepoURL:       "https://github.com/example/repo",
		DefaultBranch: "main",
		Source:        &source,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if site.ID != "ss-new" {
		t.Errorf("expected ID %q, got %q", "ss-new", site.ID)
	}
	if site.DisplayName != "New Site" {
		t.Errorf("expected display name %q, got %q", "New Site", site.DisplayName)
	}
}

func TestUpdateStaticSite_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/static-sites/ss-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StaticSite{
			ID:          "ss-123",
			Name:        "my-site",
			DisplayName: "Updated Site",
			Source:      "publicGit",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-02T00:00:00.000Z",
		})
	})
	defer server.Close()

	newName := "Updated Site"
	site, err := c.UpdateStaticSite(context.Background(), "ss-123", &UpdateStaticSiteRequest{
		DisplayName: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if site.DisplayName != "Updated Site" {
		t.Errorf("expected display name %q, got %q", "Updated Site", site.DisplayName)
	}
}

func TestDeleteStaticSite_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/static-sites/ss-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteStaticSite(context.Background(), "ss-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
