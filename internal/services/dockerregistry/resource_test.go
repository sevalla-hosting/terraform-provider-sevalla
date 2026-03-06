package dockerregistry_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccDockerRegistry_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccDockerRegistryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_docker_registry.test", "id"),
					resource.TestCheckResourceAttr("sevalla_docker_registry.test", "name", rName),
					resource.TestCheckResourceAttr("sevalla_docker_registry.test", "registry", "dockerHub"),
					resource.TestCheckResourceAttrSet("sevalla_docker_registry.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:      "sevalla_docker_registry.test",
				ImportState:       true,
				ImportStateVerify: true,
				// secret is write-only and not returned by the API
				ImportStateVerifyIgnore: []string{"secret"},
			},
			// Update name
			{
				Config: testAccDockerRegistryConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_docker_registry.test", "name", rName+"-updated"),
				),
			},
		},
	})
}

func testAccDockerRegistryConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "sevalla_docker_registry" "test" {
  name     = %[1]q
  registry = "dockerHub"
  username = "testuser"
  secret   = "testpassword"
}
`, name)
}

func testAccDockerRegistryConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "sevalla_docker_registry" "test" {
  name     = "%[1]s-updated"
  registry = "dockerHub"
  username = "testuser"
  secret   = "testpassword"
}
`, name)
}
