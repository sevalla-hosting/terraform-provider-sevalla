package application_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// fakeDomainAPI serves application domains. It never returns the custom
// certificate, like the real API.
func fakeDomainAPI(t *testing.T, domains map[string]map[string]any) http.HandlerFunc {
	var mu sync.Mutex
	next := 1

	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		id := strings.TrimPrefix(r.URL.Path, "/applications/app-1/domains/")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/applications/app-1/domains":
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Errorf("decoding create request: %v", err)
			}
			id = fmt.Sprintf("dom-%d", next)
			next++
			domains[id] = map[string]any{
				"id": id, "name": in["domain_name"], "status": "VERIFICATION", "type": "CUSTOM",
				"is_enabled": true, "is_wildcard": false, "is_primary": false,
				"dns_records": []any{map[string]any{"type": "CNAME", "name": in["domain_name"], "value": "app.sevalla.app"}},
				"errors":      nil, "created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(domains[id])
		case domains[id] != nil && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(domains[id])
		case domains[id] != nil && r.Method == http.MethodDelete:
			delete(domains, id)
			json.NewEncoder(w).Encode(map[string]any{"message": "deleted"})
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"message": "Not found", "status": 404})
		}
	}
}

func domainConfig(withCert bool) string {
	cert := ""
	if withCert {
		cert = `
  custom_ssl_cert = "CERT"
  custom_ssl_key  = "KEY"`
	}
	return fmt.Sprintf(`
resource "sevalla_application_domain" "test" {
  application_id = "app-1"
  name           = "shop.example.com"%s
}
`, cert)
}

func TestApplicationDomain_FakeAPIAddingCertificateReplaces(t *testing.T) {
	testutil.FakeAPI(t, fakeDomainAPI(t, map[string]map[string]any{}))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: domainConfig(false),
				Check:  resource.TestCheckResourceAttrSet("sevalla_application_domain.test", "dns_records"),
			},
			{
				// The certificate can only be set when a domain is created.
				Config: domainConfig(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sevalla_application_domain.test", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestApplicationDomain_FakeAPIImportRecordsCertificate(t *testing.T) {
	testutil.FakeAPI(t, fakeDomainAPI(t, map[string]map[string]any{
		"dom-9": {
			"id": "dom-9", "name": "shop.example.com", "status": "ACTIVE", "type": "CUSTOM",
			"is_enabled": true, "is_wildcard": false, "is_primary": true, "dns_records": nil, "errors": nil,
			"created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
		},
	}))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             domainConfig(true),
				ResourceName:       "sevalla_application_domain.test",
				ImportState:        true,
				ImportStateId:      "app-1/dom-9",
				ImportStatePersist: true,
			},
			{
				// An imported domain must not be destroyed to record its certificate.
				Config: domainConfig(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sevalla_application_domain.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("sevalla_application_domain.test", "id", "dom-9"),
			},
		},
	})
}

// The API never reports a domain's certificate, so a certificate added after an
// import can only be recorded (with a warning) instead of replacing the domain.
func TestApplicationDomain_FakeAPICertificateAddedAfterImportIsRecorded(t *testing.T) {
	testutil.FakeAPI(t, fakeDomainAPI(t, map[string]map[string]any{
		"dom-9": {
			"id": "dom-9", "name": "shop.example.com", "status": "ACTIVE", "type": "CUSTOM",
			"is_enabled": true, "is_wildcard": false, "is_primary": true, "dns_records": nil, "errors": nil,
			"created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
		},
	}))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             domainConfig(false),
				ResourceName:       "sevalla_application_domain.test",
				ImportState:        true,
				ImportStateId:      "app-1/dom-9",
				ImportStatePersist: true,
			},
			{Config: domainConfig(false)},
			{
				Config: domainConfig(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sevalla_application_domain.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

// fakeTCPProxyAPI has no endpoint to get a single proxy, like the real API,
// and reports a new proxy as creating until it is listed once.
func fakeTCPProxyAPI(t *testing.T) http.HandlerFunc {
	var mu sync.Mutex
	var proxy map[string]any

	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/applications/app-1/tcp-proxies":
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Errorf("decoding create request: %v", err)
			}
			proxy = map[string]any{
				"id": "tcp-1", "app_id": "app-1", "process_id": in["process_id"], "port": in["port"],
				"external_port": 31000, "hostname": "tcp.sevalla.app", "status": "creating",
				"created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(proxy)
		case r.Method == http.MethodGet && r.URL.Path == "/applications/app-1/tcp-proxies":
			data := []any{}
			if proxy != nil {
				data = append(data, proxy)
				proxy = mapWith(proxy, "status", "ready")
			}
			json.NewEncoder(w).Encode(map[string]any{"data": data, "total": len(data), "offset": 0, "limit": 100})
		case proxy != nil && r.Method == http.MethodDelete && r.URL.Path == "/applications/app-1/tcp-proxies/tcp-1":
			proxy = nil
			json.NewEncoder(w).Encode(map[string]any{"message": "deleted"})
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"message": "Not found", "status": 404})
		}
	}
}

func mapWith(m map[string]any, key string, value any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	out[key] = value
	return out
}

func TestTCPProxy_FakeAPILifecycle(t *testing.T) {
	testutil.FakeAPI(t, fakeTCPProxyAPI(t))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sevalla_application_tcp_proxy" "test" {
  application_id = "app-1"
  process_id     = "proc-1"
  port           = 5432
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_application_tcp_proxy.test", "status", "ready"),
					resource.TestCheckResourceAttr("sevalla_application_tcp_proxy.test", "external_port", "31000"),
				),
			},
			{
				ResourceName:      "sevalla_application_tcp_proxy.test",
				ImportState:       true,
				ImportStateId:     "app-1/tcp-1",
				ImportStateVerify: true,
			},
		},
	})
}
