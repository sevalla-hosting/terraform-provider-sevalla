package project_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// ---------------------------------------------------------------------------
// Project Service
// ---------------------------------------------------------------------------

func TestAccProjectService_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_project_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccProjectServiceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", "sevalla_project.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "service_id", "sevalla_pipeline.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pipeline"),
				),
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_type"},
				ImportStateIdFunc:       testAccProjectServiceImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccProjectServiceConfig(name string) string {
	return fmt.Sprintf(`
resource "sevalla_project" "test" {
  display_name = %[1]q
}

resource "sevalla_pipeline" "test" {
  display_name = "%[1]s-pipeline"
  type         = "trunk"
}

resource "sevalla_project_service" "test" {
  project_id   = sevalla_project.test.id
  service_id   = sevalla_pipeline.test.id
  service_type = "pipeline"
}
`, name)
}

// ---------------------------------------------------------------------------
// Data Source
// ---------------------------------------------------------------------------

func TestAccProjectDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.sevalla_project.test", "id",
						"sevalla_project.test", "id",
					),
					resource.TestCheckResourceAttr("data.sevalla_project.test", "display_name", rName),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "sevalla_project" "test" {
  display_name = %[1]q
}

data "sevalla_project" "test" {
  id = sevalla_project.test.id
}
`, name)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testAccProjectServiceImportStateIdFunc builds a composite import ID
// in the format "project_id/service_id" from the resource state.
func testAccProjectServiceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		projectID := rs.Primary.Attributes["project_id"]
		if projectID == "" {
			return "", fmt.Errorf("project_id not set on %s", resourceName)
		}
		serviceID := rs.Primary.Attributes["service_id"]
		if serviceID == "" {
			return "", fmt.Errorf("service_id not set on %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", projectID, serviceID), nil
	}
}
