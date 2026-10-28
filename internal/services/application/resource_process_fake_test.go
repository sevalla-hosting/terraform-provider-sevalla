package application_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// fakeProcessAPI mimics the process endpoints closely enough to run a full
// Terraform lifecycle: it fills in server-side defaults and rejects the
// empty values that unknown plan attributes used to be sent as.
func fakeProcessAPI(t *testing.T) http.HandlerFunc {
	var mu sync.Mutex
	var process map[string]any

	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/applications/app-1/processes":
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Errorf("decoding create request: %v", err)
			}
			if _, ok := in["is_ingress_enabled"]; ok {
				t.Error("create request must not include is_ingress_enabled")
			}
			probe := in["liveness_probe"].(map[string]any)
			probe["timeoutSeconds"] = 120 // API default
			process = map[string]any{
				"id": "proc-1", "app_id": "app-1", "key": "web", "type": in["type"],
				"display_name": in["display_name"], "entrypoint": "", "port": in["port"],
				"is_ingress_enabled": true, "ingress_protocol": "http",
				"scaling_strategy": in["scaling_strategy"], "resource_type_id": in["resource_type_id"],
				"resource_type_name": "S1", "cpu_limit": 500, "memory_limit": 512,
				"schedule": nil, "time_zone": nil, "concurrency_policy": nil, "active_deadline_seconds": nil,
				"job_start_policy": nil, "liveness_probe": probe, "readiness_probe": nil,
				"internal_hostname": "web.internal", "created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
			}
			json.NewEncoder(w).Encode(process)
		case process != nil && r.URL.Path == "/applications/app-1/processes/proc-1":
			switch r.Method {
			case http.MethodGet:
				json.NewEncoder(w).Encode(process)
			case http.MethodPatch:
				var in map[string]any
				if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
					t.Errorf("decoding update request: %v", err)
				}
				if _, ok := in["liveness_probe"]; ok {
					t.Error("update request must not resend an unchanged probe")
				}
				for k, v := range in {
					if v == "" || v == float64(0) {
						w.WriteHeader(http.StatusBadRequest)
						json.NewEncoder(w).Encode(map[string]any{"message": "invalid " + k, "status": 400})
						return
					}
					process[k] = v
				}
				json.NewEncoder(w).Encode(process)
			case http.MethodDelete:
				process = nil
				json.NewEncoder(w).Encode(map[string]any{"message": "deleted"})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"message": "Not found", "status": 404})
		}
	}
}

func TestProcessResource_FakeAPILifecycle(t *testing.T) {
	testutil.FakeAPI(t, fakeProcessAPI(t))

	config := func(name string) string {
		return strings.ReplaceAll(`
resource "sevalla_application_process" "web" {
  application_id   = "app-1"
  display_name     = "NAME"
  type             = "web"
  port               = 8080
  resource_type_id   = "rt-1"
  is_ingress_enabled = false

  scaling_strategy = {
    type           = "manual"
    instance_count = 1
  }

  liveness_probe = {
    http_get = {
      path = "/healthz"
    }
  }
}
`, "NAME", name)
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("web"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_application_process.web", "id", "proc-1"),
					// The create endpoint ignores ingress settings, so they are applied with an update.
					resource.TestCheckResourceAttr("sevalla_application_process.web", "is_ingress_enabled", "false"),
					resource.TestCheckResourceAttr("sevalla_application_process.web", "liveness_probe.timeout_seconds", "120"),
				),
			},
			{
				Config: config("web-renamed"),
				Check:  resource.TestCheckResourceAttr("sevalla_application_process.web", "display_name", "web-renamed"),
			},
			{
				ResourceName:      "sevalla_application_process.web",
				ImportState:       true,
				ImportStateId:     "app-1/proc-1",
				ImportStateVerify: true,
			},
		},
	})
}
