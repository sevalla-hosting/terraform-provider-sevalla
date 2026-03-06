package application_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccApplication_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_application.test", "id"),
					resource.TestCheckResourceAttr("sevalla_application.test", "display_name", rName),
					resource.TestCheckResourceAttr("sevalla_application.test", "source", "publicGit"),
					resource.TestCheckResourceAttrSet("sevalla_application.test", "name"),
					resource.TestCheckResourceAttrSet("sevalla_application.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:      "sevalla_application.test",
				ImportState:       true,
				ImportStateVerify: true,
				// cluster_id is not returned by the API on import
				ImportStateVerifyIgnore: []string{"cluster_id"},
			},
			// Update
			{
				Config: testAccApplicationConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_application.test", "display_name", rName+"-updated"),
					resource.TestCheckResourceAttr("sevalla_application.test", "auto_deploy", "false"),
				),
			},
		},
	})
}

func testAccApplicationConfig_basic(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}

resource "sevalla_application" "test" {
  display_name = %[1]q
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/kotapeter/pack-demo"
}
`, name)
}

func testAccApplicationConfig_updated(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}

resource "sevalla_application" "test" {
  display_name = "%[1]s-updated"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/kotapeter/pack-demo"
  auto_deploy  = false
}
`, name)
}
