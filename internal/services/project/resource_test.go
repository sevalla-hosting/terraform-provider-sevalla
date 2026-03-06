package project_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccProject_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_project.test", "id"),
					resource.TestCheckResourceAttr("sevalla_project.test", "display_name", rName),
					resource.TestCheckResourceAttrSet("sevalla_project.test", "name"),
					resource.TestCheckResourceAttrSet("sevalla_project.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:      "sevalla_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update display_name
			{
				Config: testAccProjectConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_project.test", "display_name", rName+"-updated"),
				),
			},
		},
	})
}

func testAccProjectConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "sevalla_project" "test" {
  display_name = %[1]q
}
`, name)
}

func testAccProjectConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "sevalla_project" "test" {
  display_name = "%[1]s-updated"
}
`, name)
}
