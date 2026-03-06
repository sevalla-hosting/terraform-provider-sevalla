package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPaginate_SinglePage(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PaginatedResponse[ApplicationListItem]{
			Data: []ApplicationListItem{
				{ID: "app-1", Name: "app-1", DisplayName: "App 1", Source: "publicGit", Type: "app", CreatedAt: "2025-01-01T00:00:00.000Z", UpdatedAt: "2025-01-01T00:00:00.000Z"},
			},
			Total:  1,
			Offset: 0,
			Limit:  100,
		})
	})
	defer server.Close()

	items, err := Paginate[ApplicationListItem](context.Background(), c, "/applications")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestPaginate_MultiplePages(t *testing.T) {
	var requestCount int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		offsetStr := r.URL.Query().Get("offset")
		offset, _ := strconv.Atoi(offsetStr)

		w.Header().Set("Content-Type", "application/json")

		if offset == 0 {
			json.NewEncoder(w).Encode(PaginatedResponse[ApplicationListItem]{
				Data: []ApplicationListItem{
					{ID: "app-1", Name: "app-1", DisplayName: "App 1", Source: "publicGit", Type: "app", CreatedAt: "2025-01-01T00:00:00.000Z", UpdatedAt: "2025-01-01T00:00:00.000Z"},
				},
				Total:  2,
				Offset: 0,
				Limit:  1,
			})
		} else {
			json.NewEncoder(w).Encode(PaginatedResponse[ApplicationListItem]{
				Data: []ApplicationListItem{
					{ID: "app-2", Name: "app-2", DisplayName: "App 2", Source: "publicGit", Type: "app", CreatedAt: "2025-01-01T00:00:00.000Z", UpdatedAt: "2025-01-01T00:00:00.000Z"},
				},
				Total:  2,
				Offset: 1,
				Limit:  1,
			})
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	c := NewClient("test-api-key",
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)

	// Note: Paginate uses limit=100 internally, so we simulate the API returning
	// total > limit to trigger multiple pages. We use total=200 with limit=100.
	// However, the test server above returns total=2 with limit=1, and the actual
	// Paginate function uses its own limit=100. So let's make a proper test:
	requestCount = 0
	items, err := Paginate[ApplicationListItem](context.Background(), c, "/applications")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The paginator sends limit=100. Our server returns total=2 with the first
	// offset=0 request, so offset(0)+limit(100) >= total(2), meaning it stops
	// after one page.
	if len(items) != 1 {
		t.Errorf("expected 1 item from single request, got %d", len(items))
	}

	// Now test with a server that actually forces multiple pages
	requestCount = 0
	multiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		offsetStr := r.URL.Query().Get("offset")
		offset, _ := strconv.Atoi(offsetStr)
		limitStr := r.URL.Query().Get("limit")
		limit, _ := strconv.Atoi(limitStr)

		w.Header().Set("Content-Type", "application/json")

		// Total is 150, limit from paginator is 100
		total := 150
		var data []ApplicationListItem
		for i := offset; i < offset+limit && i < total; i++ {
			data = append(data, ApplicationListItem{
				ID:          "app-" + strconv.Itoa(i),
				Name:        "app-" + strconv.Itoa(i),
				DisplayName: "App " + strconv.Itoa(i),
				Source:      "publicGit",
				Type:        "app",
				CreatedAt:   "2025-01-01T00:00:00.000Z",
				UpdatedAt:   "2025-01-01T00:00:00.000Z",
			})
		}

		json.NewEncoder(w).Encode(PaginatedResponse[ApplicationListItem]{
			Data:   data,
			Total:  total,
			Offset: offset,
			Limit:  limit,
		})
	}))
	defer multiServer.Close()

	c2 := NewClient("test-api-key",
		WithBaseURL(multiServer.URL),
		WithHTTPClient(multiServer.Client()),
	)

	items, err = Paginate[ApplicationListItem](context.Background(), c2, "/applications")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 150 {
		t.Errorf("expected 150 items, got %d", len(items))
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests (offset=0, offset=100), got %d", requestCount)
	}
}

func TestPaginate_EmptyResult(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PaginatedResponse[ApplicationListItem]{
			Data:   []ApplicationListItem{},
			Total:  0,
			Offset: 0,
			Limit:  100,
		})
	})
	defer server.Close()

	items, err := Paginate[ApplicationListItem](context.Background(), c, "/applications")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestPaginate_APIError(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Internal server error",
			"status":  500,
		})
	})
	defer server.Close()

	_, err := Paginate[ApplicationListItem](context.Background(), c, "/applications")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
	}
}
