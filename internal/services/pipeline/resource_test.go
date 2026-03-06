package pipeline_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccPipeline_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccPipelineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_pipeline.test", "id"),
					resource.TestCheckResourceAttr("sevalla_pipeline.test", "display_name", rName),
					resource.TestCheckResourceAttr("sevalla_pipeline.test", "type", "trunk"),
					resource.TestCheckResourceAttrSet("sevalla_pipeline.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:      "sevalla_pipeline.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update display_name
			{
				Config: testAccPipelineConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_pipeline.test", "display_name", rName+"-updated"),
				),
			},
		},
	})
}

func testAccPipelineConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "sevalla_pipeline" "test" {
  display_name = %[1]q
  type         = "trunk"
}
`, name)
}

func testAccPipelineConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "sevalla_pipeline" "test" {
  display_name = "%[1]s-updated"
  type         = "trunk"
}
`, name)
}
