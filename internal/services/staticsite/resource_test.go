package staticsite_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccStaticSite_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccStaticSiteConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_static_site.test", "id"),
					resource.TestCheckResourceAttr("sevalla_static_site.test", "display_name", rName),
					resource.TestCheckResourceAttr("sevalla_static_site.test", "source", "publicGit"),
					resource.TestCheckResourceAttrSet("sevalla_static_site.test", "name"),
					resource.TestCheckResourceAttrSet("sevalla_static_site.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:            "sevalla_static_site.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "hostname"},
			},
			// Update
			{
				Config: testAccStaticSiteConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_static_site.test", "display_name", rName+"-updated"),
					resource.TestCheckResourceAttr("sevalla_static_site.test", "auto_deploy", "false"),
				),
			},
		},
	})
}

func testAccStaticSiteConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "sevalla_static_site" "test" {
  display_name   = %[1]q
  source         = "publicGit"
  repo_url       = "https://github.com/kotapeter/pack-demo"
  default_branch = "main"
}
`, name)
}

func testAccStaticSiteConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "sevalla_static_site" "test" {
  display_name   = "%[1]s-updated"
  source         = "publicGit"
  repo_url       = "https://github.com/kotapeter/pack-demo"
  default_branch = "main"
  auto_deploy    = false
}
`, name)
}
