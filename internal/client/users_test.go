package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListUsers_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(usersResponse{
			Data: []User{
				{
					ID:       "user-1",
					Email:    "john@example.com",
					FullName: "John Doe",
					Image:    "https://example.com/john.png",
				},
				{
					ID:       "user-2",
					Email:    "jane@example.com",
					FullName: "Jane Smith",
					Image:    "https://example.com/jane.png",
				},
			},
		})
	})
	defer server.Close()

	users, err := c.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Email != "john@example.com" {
		t.Errorf("expected email %q, got %q", "john@example.com", users[0].Email)
	}
	if users[0].FullName != "John Doe" {
		t.Errorf("expected full name %q, got %q", "John Doe", users[0].FullName)
	}
	if users[1].Email != "jane@example.com" {
		t.Errorf("expected email %q, got %q", "jane@example.com", users[1].Email)
	}
}

func TestListUsers_Error(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Forbidden",
			"status":  403,
		})
	})
	defer server.Close()

	_, err := c.ListUsers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("expected status code 403, got %d", apiErr.StatusCode)
	}
}
