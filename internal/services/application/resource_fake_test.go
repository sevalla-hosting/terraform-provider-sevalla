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

// fakeApplicationAPI keeps the settings of other build types when build_type
// changes, like the real API, and omits fields from PATCH responses.
func fakeApplicationAPI(t *testing.T) http.HandlerFunc {
	var mu sync.Mutex
	var app map[string]any

	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		decode := func() map[string]any {
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Errorf("decoding %s request: %v", r.Method, err)
			}
			return in
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/applications":
			in := decode()
			app = map[string]any{
				"id": "app-1", "company_id": "c-1", "project_id": nil, "name": "web-1", "namespace": "web-1",
				"display_name": in["display_name"], "source": in["source"], "type": "app", "status": nil,
				"build_cache_enabled": true, "git_lfs_enabled": true, "hibernation_enabled": false,
				"hibernate_after_seconds": nil, "auto_deploy": false, "is_suspended": false,
				"is_cdn_enabled": false, "is_edge_cache_enabled": false, "git_type": nil,
				"repo_url": in["repo_url"], "default_branch": "main", "docker_image": nil,
				"build_type": "nixpacks", "build_path": nil, "pack_builder": nil, "nixpacks_version": nil,
				"dockerfile_path": nil, "docker_context": nil, "allow_deploy_paths": []any{}, "ignore_deploy_paths": []any{},
				"buildpacks": []any{}, "cluster_id": in["cluster_id"], "docker_registry_credential_id": nil,
				"created_by": nil, "created_at": "2025-01-01T00:00:00.000Z", "updated_at": "2025-01-01T00:00:00.000Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(app)
		case app != nil && r.URL.Path == "/applications/app-1":
			switch r.Method {
			case http.MethodGet:
				json.NewEncoder(w).Encode(app)
			case http.MethodPatch:
				for k, v := range decode() {
					app[k] = v
				}
				partial := map[string]any{}
				for k, v := range app {
					if k != "buildpacks" && k != "cluster_id" {
						partial[k] = v
					}
				}
				json.NewEncoder(w).Encode(partial)
			case http.MethodDelete:
				app = nil
				json.NewEncoder(w).Encode(map[string]any{"message": "deleted"})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"message": "Not found", "status": 404})
		}
	}
}

func TestApplicationResource_FakeAPIBuildTypeSwitch(t *testing.T) {
	testutil.FakeAPI(t, fakeApplicationAPI(t))

	config := func(buildSettings string) string {
		return strings.ReplaceAll(`
resource "sevalla_application" "web" {
  display_name = "web"
  cluster_id   = "cluster-1"
  source       = "publicGit"
  repo_url     = "https://github.com/acme/web"
  SETTINGS
}
`, "SETTINGS", buildSettings)
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(`
  build_type   = "pack"
  pack_builder = "heroku/builder:24"
  buildpacks   = [{ order = 0, source = "heroku/nodejs" }]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_application.web", "pack_builder", "heroku/builder:24"),
					resource.TestCheckResourceAttr("sevalla_application.web", "cluster_id", "cluster-1"),
				),
			},
			{
				// The API keeps the pack settings, which must not surface as a diff.
				Config: config(`build_type = "nixpacks"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("sevalla_application.web", "pack_builder"),
					resource.TestCheckNoResourceAttr("sevalla_application.web", "buildpacks.#"),
				),
			},
			{
				// Switching back reads the kept settings from the API.
				Config: config(`build_type = "pack"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_application.web", "pack_builder", "heroku/builder:24"),
					resource.TestCheckResourceAttr("sevalla_application.web", "buildpacks.#", "1"),
				),
			},
			{
				ResourceName:      "sevalla_application.web",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
