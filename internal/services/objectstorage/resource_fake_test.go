package objectstorage_test

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// fakeObjectStorageAPI returns the keys only on create and deletes
// asynchronously, like the real API.
func fakeObjectStorageAPI(t *testing.T) http.HandlerFunc {
	var mu sync.Mutex
	var bucket map[string]any

	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/object-storage":
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Errorf("decoding create request: %v", err)
			}
			bucket = map[string]any{
				"id": "os-1", "company_id": "c-1", "project_id": nil, "name": "media", "display_name": in["display_name"],
				"location": "enam", "jurisdiction": "default", "status": "active", "bucket_name": "media-1",
				"domain": nil, "endpoint": "https://s3.example.com",
				"created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
			}
			out := map[string]any{"access_key": "AK", "secret_key": "SK"}
			for k, v := range bucket {
				out[k] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(out)
		case bucket != nil && r.URL.Path == "/object-storage/os-1":
			switch r.Method {
			case http.MethodGet:
				json.NewEncoder(w).Encode(bucket)
			case http.MethodPatch:
				var in map[string]any
				if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
					t.Errorf("decoding update request: %v", err)
				}
				for k, v := range in {
					bucket[k] = v
				}
				json.NewEncoder(w).Encode(bucket)
			case http.MethodDelete:
				bucket = nil
				w.WriteHeader(http.StatusAccepted)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"message": "Not found", "status": 404})
		}
	}
}

func TestObjectStorageResource_FakeAPILifecycle(t *testing.T) {
	testutil.FakeAPI(t, fakeObjectStorageAPI(t))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "sevalla_object_storage" "test" { display_name = "media" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_object_storage.test", "location", "enam"),
					resource.TestCheckResourceAttr("sevalla_object_storage.test", "access_key", "AK"),
				),
			},
			{
				// Renaming must update in place and keep the create-only keys.
				Config: `resource "sevalla_object_storage" "test" { display_name = "media-renamed" }`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sevalla_object_storage.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_object_storage.test", "display_name", "media-renamed"),
					resource.TestCheckResourceAttr("sevalla_object_storage.test", "access_key", "AK"),
					resource.TestCheckResourceAttr("sevalla_object_storage.test", "secret_key", "SK"),
				),
			},
		},
	})
}
