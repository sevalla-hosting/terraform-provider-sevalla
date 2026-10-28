package webhook_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla/terraform-provider-sevalla/internal/testutil"
)

// fakeWebhookAPI enables new webhooks and reorders allowed_events, so the
// toggle and the order-preserving read are both exercised.
func fakeWebhookAPI(t *testing.T, toggles *int) http.HandlerFunc {
	var mu sync.Mutex
	var webhook map[string]any

	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/webhooks":
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Errorf("decoding create request: %v", err)
			}
			events := in["allowed_events"].([]any)
			webhook = map[string]any{
				"id": "wh-1", "company_id": "c-1", "is_enabled": true, "endpoint": in["endpoint"],
				"allowed_events": []any{events[len(events)-1], events[0]}, "secret": "whsec_1",
				"old_secret": nil, "old_secret_expired_at": nil, "description": in["description"],
				"created_by": "key-1", "updated_by": nil,
				"created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(webhook)
		case webhook != nil && r.URL.Path == "/webhooks/wh-1/toggle" && r.Method == http.MethodPost:
			*toggles++
			webhook["is_enabled"] = !webhook["is_enabled"].(bool)
			json.NewEncoder(w).Encode(webhook)
		case webhook != nil && r.URL.Path == "/webhooks/wh-1":
			switch r.Method {
			case http.MethodGet:
				json.NewEncoder(w).Encode(webhook)
			case http.MethodDelete:
				webhook = nil
				json.NewEncoder(w).Encode(map[string]any{"message": "deleted"})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"message": "Not found", "status": 404})
		}
	}
}

func webhookConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "sevalla_webhook" "test" {
  endpoint       = "https://hooks.example.com"
  allowed_events = ["APP_DEPLOY", "STATIC_SITE_DEPLOY"]
  is_enabled     = %t
}
`, enabled)
}

func TestWebhook_FakeAPIToggle(t *testing.T) {
	toggles := 0
	testutil.FakeAPI(t, fakeWebhookAPI(t, &toggles))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: webhookConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_webhook.test", "is_enabled", "false"),
					resource.TestCheckResourceAttr("sevalla_webhook.test", "allowed_events.0", "APP_DEPLOY"),
				),
			},
			{
				Config: webhookConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_webhook.test", "is_enabled", "true"),
					func(_ *terraform.State) error {
						if toggles != 2 {
							return fmt.Errorf("expected 2 toggles, got %d", toggles)
						}
						return nil
					},
				),
			},
		},
	})
}
