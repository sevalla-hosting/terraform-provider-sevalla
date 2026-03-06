package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListGlobalEnvVars_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/global-env-vars" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PaginatedResponse[GlobalEnvironmentVariable]{
			Data: []GlobalEnvironmentVariable{
				{ID: "genv-1", Key: "GLOBAL_API_URL", Value: "https://api.example.com"},
				{ID: "genv-2", Key: "GLOBAL_LOG_LEVEL", Value: "info"},
			},
			Total:  2,
			Offset: 0,
			Limit:  100,
		})
	})
	defer server.Close()

	envVars, err := c.ListGlobalEnvVars(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(envVars) != 2 {
		t.Fatalf("expected 2 global env vars, got %d", len(envVars))
	}
	if envVars[0].Key != "GLOBAL_API_URL" {
		t.Errorf("expected key %q, got %q", "GLOBAL_API_URL", envVars[0].Key)
	}
	if envVars[1].Value != "info" {
		t.Errorf("expected value %q, got %q", "info", envVars[1].Value)
	}
}

func TestListGlobalEnvVars_Error(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Internal server error",
			"status":  500,
		})
	})
	defer server.Close()

	_, err := c.ListGlobalEnvVars(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateGlobalEnvVar_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/global-env-vars" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateGlobalEnvVarRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Key != "NEW_GLOBAL_VAR" {
			t.Errorf("expected key %q, got %q", "NEW_GLOBAL_VAR", req.Key)
		}
		if req.Value != "global-value" {
			t.Errorf("expected value %q, got %q", "global-value", req.Value)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(GlobalEnvironmentVariable{
			ID:    "genv-new",
			Key:   req.Key,
			Value: req.Value,
		})
	})
	defer server.Close()

	envVar, err := c.CreateGlobalEnvVar(context.Background(), &CreateGlobalEnvVarRequest{
		Key:   "NEW_GLOBAL_VAR",
		Value: "global-value",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if envVar.ID != "genv-new" {
		t.Errorf("expected ID %q, got %q", "genv-new", envVar.ID)
	}
	if envVar.Key != "NEW_GLOBAL_VAR" {
		t.Errorf("expected key %q, got %q", "NEW_GLOBAL_VAR", envVar.Key)
	}
}

func TestUpdateGlobalEnvVar_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/global-env-vars/genv-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GlobalEnvironmentVariable{
			ID:    "genv-123",
			Key:   "UPDATED_KEY",
			Value: "updated-value",
		})
	})
	defer server.Close()

	newValue := "updated-value"
	envVar, err := c.UpdateGlobalEnvVar(context.Background(), "genv-123", &UpdateGlobalEnvVarRequest{
		Value: &newValue,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if envVar.Value != "updated-value" {
		t.Errorf("expected value %q, got %q", "updated-value", envVar.Value)
	}
}

func TestDeleteGlobalEnvVar_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/global-env-vars/genv-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteGlobalEnvVar(context.Background(), "genv-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
