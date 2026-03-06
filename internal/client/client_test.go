package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) (*SevallaClient, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	c := NewClient("test-api-key",
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	return c, server
}

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("my-key")
	if c.BaseURL != DefaultBaseURL {
		t.Errorf("expected base URL %q, got %q", DefaultBaseURL, c.BaseURL)
	}
	if c.HTTPClient == nil {
		t.Fatal("expected non-nil HTTP client")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	c := NewClient("key", WithBaseURL("https://custom.api"), WithUserAgent("custom-agent"))
	if c.BaseURL != "https://custom.api" {
		t.Errorf("expected base URL %q, got %q", "https://custom.api", c.BaseURL)
	}
	if c.UserAgent != "custom-agent" {
		t.Errorf("expected user agent %q, got %q", "custom-agent", c.UserAgent)
	}
}

func TestGetApplication_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/test-id" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Application{
			ID:          "test-id",
			DisplayName: "Test App",
			Name:        "test-app",
			Source:      "publicGit",
			Type:        "app",
			BuildType:   "nixpacks",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	app, err := c.GetApplication(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != "test-id" {
		t.Errorf("expected ID %q, got %q", "test-id", app.ID)
	}
	if app.DisplayName != "Test App" {
		t.Errorf("expected display name %q, got %q", "Test App", app.DisplayName)
	}
}

func TestGetApplication_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetApplication(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true, got false")
	}
}

func TestCreateApplication_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateApplicationRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "New App" {
			t.Errorf("expected display name %q, got %q", "New App", req.DisplayName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Application{
			ID:          "new-id",
			DisplayName: req.DisplayName,
			Name:        "new-app",
			Source:      req.Source,
			Type:        "app",
			BuildType:   "nixpacks",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	app, err := c.CreateApplication(context.Background(), &CreateApplicationRequest{
		DisplayName: "New App",
		ClusterID:   "cluster-1",
		Source:      "publicGit",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != "new-id" {
		t.Errorf("expected ID %q, got %q", "new-id", app.ID)
	}
}

func TestUpdateApplication_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req UpdateApplicationRequest
		json.NewDecoder(r.Body).Decode(&req)

		name := "Updated App"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Application{
			ID:          "test-id",
			DisplayName: name,
			Name:        "test-app",
			Source:      "publicGit",
			Type:        "app",
			BuildType:   "nixpacks",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-02T00:00:00.000Z",
		})
	})
	defer server.Close()

	newName := "Updated App"
	app, err := c.UpdateApplication(context.Background(), "test-id", &UpdateApplicationRequest{
		DisplayName: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.DisplayName != "Updated App" {
		t.Errorf("expected display name %q, got %q", "Updated App", app.DisplayName)
	}
}

func TestDeleteApplication_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteApplication(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListApplications_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PaginatedResponse[ApplicationListItem]{
			Data: []ApplicationListItem{
				{ID: "app-1", Name: "app-1", DisplayName: "App 1", Source: "publicGit", Type: "app", CreatedAt: "2025-01-01T00:00:00.000Z", UpdatedAt: "2025-01-01T00:00:00.000Z"},
				{ID: "app-2", Name: "app-2", DisplayName: "App 2", Source: "publicGit", Type: "app", CreatedAt: "2025-01-01T00:00:00.000Z", UpdatedAt: "2025-01-01T00:00:00.000Z"},
			},
			Total:  2,
			Offset: 0,
			Limit:  100,
		})
	})
	defer server.Close()

	apps, err := c.ListApplications(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Errorf("expected 2 apps, got %d", len(apps))
	}
	if apps[0].ID != "app-1" {
		t.Errorf("expected first app ID %q, got %q", "app-1", apps[0].ID)
	}
}

func TestListClusters_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/resources/clusters" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		displayName := "Iowa, USA"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Cluster{
			{ID: "cluster-1", Name: "us-central1", DisplayName: &displayName, Location: "us-central1"},
		})
	})
	defer server.Close()

	clusters, err := c.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	if clusters[0].Name != "us-central1" {
		t.Errorf("expected name %q, got %q", "us-central1", clusters[0].Name)
	}
}

func TestAPIError_Format(t *testing.T) {
	err := &APIError{StatusCode: 400, Message: "Bad request"}
	expected := "sevalla API error (HTTP 400): Bad request"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}

	errWithCode := &APIError{
		StatusCode: 400,
		Message:    "Bad request",
		Data: &struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: "err-123", Message: "details"},
	}
	expected = "sevalla API error (HTTP 400): Bad request [err-123]"
	if errWithCode.Error() != expected {
		t.Errorf("expected %q, got %q", expected, errWithCode.Error())
	}
}

func TestIsNotFound(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("expected false for nil error")
	}
	if IsNotFound(&APIError{StatusCode: 400}) {
		t.Error("expected false for 400")
	}
	if !IsNotFound(&APIError{StatusCode: 404}) {
		t.Error("expected true for 404")
	}
}
