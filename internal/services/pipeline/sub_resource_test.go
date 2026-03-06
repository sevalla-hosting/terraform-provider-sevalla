package pipeline_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// ---------------------------------------------------------------------------
// Pipeline Stage
// ---------------------------------------------------------------------------

func TestAccPipelineStage_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_pipeline_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccPipelineStageConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "pipeline_id", "sevalla_pipeline.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName+"-stage"),
					resource.TestCheckResourceAttrSet(resourceName, "type"),
					resource.TestCheckResourceAttrSet(resourceName, "order"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"insert_before", "updated_at"},
				ImportStateIdFunc:       testAccPipelineSubResourceImportStateIdFunc(resourceName, "pipeline_id", "id"),
			},
		},
	})
}

func testAccPipelineStageConfig(name string) string {
	return fmt.Sprintf(`
resource "sevalla_pipeline" "test" {
  display_name = %[1]q
  type         = "trunk"
}

resource "sevalla_pipeline_stage" "test" {
  pipeline_id   = sevalla_pipeline.test.id
  name          = "%[1]s-stage"
  insert_before = 1
}
`, name)
}

// ---------------------------------------------------------------------------
// Pipeline Stage Application
// ---------------------------------------------------------------------------

func TestAccPipelineStageApplication_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_pipeline_stage_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccPipelineStageApplicationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "pipeline_id", "sevalla_pipeline.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "stage_id", "sevalla_pipeline_stage.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "sevalla_application.test", "id"),
				),
			},
			// Step 2: Import
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "application_id",
				ImportStateIdFunc:                    testAccPipelineStageApplicationImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccPipelineStageApplicationConfig(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}

resource "sevalla_application" "test" {
  display_name = "%[1]s-app"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/kotapeter/pack-demo"
}

resource "sevalla_pipeline" "test" {
  display_name = "%[1]s-pipeline"
  type         = "trunk"
}

resource "sevalla_pipeline_stage" "test" {
  pipeline_id   = sevalla_pipeline.test.id
  name          = "%[1]s-stage"
  insert_before = 1
}

resource "sevalla_pipeline_stage_application" "test" {
  pipeline_id    = sevalla_pipeline.test.id
  stage_id       = sevalla_pipeline_stage.test.id
  application_id = sevalla_application.test.id
}
`, name)
}

// ---------------------------------------------------------------------------
// Data Source
// ---------------------------------------------------------------------------

func TestAccPipelineDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.sevalla_pipeline.test", "id",
						"sevalla_pipeline.test", "id",
					),
					resource.TestCheckResourceAttr("data.sevalla_pipeline.test", "display_name", rName),
				),
			},
		},
	})
}

func testAccPipelineDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "sevalla_pipeline" "test" {
  display_name = %[1]q
  type         = "trunk"
}

data "sevalla_pipeline" "test" {
  id = sevalla_pipeline.test.id
}
`, name)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testAccPipelineSubResourceImportStateIdFunc builds a composite import ID
// in the format "pipeline_id/sub_resource_id" from the given attribute names.
func testAccPipelineSubResourceImportStateIdFunc(resourceName, parentIDAttr, subIDAttr string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		parentID := rs.Primary.Attributes[parentIDAttr]
		if parentID == "" {
			return "", fmt.Errorf("%s not set on %s", parentIDAttr, resourceName)
		}
		subResourceID := rs.Primary.Attributes[subIDAttr]
		if subResourceID == "" {
			return "", fmt.Errorf("%s not set on %s", subIDAttr, resourceName)
		}
		return fmt.Sprintf("%s/%s", parentID, subResourceID), nil
	}
}

// testAccPipelineStageApplicationImportStateIdFunc builds a composite import ID
// in the format "pipeline_id/stage_id/application_id".
func testAccPipelineStageApplicationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		pipelineID := rs.Primary.Attributes["pipeline_id"]
		if pipelineID == "" {
			return "", fmt.Errorf("pipeline_id not set on %s", resourceName)
		}
		stageID := rs.Primary.Attributes["stage_id"]
		if stageID == "" {
			return "", fmt.Errorf("stage_id not set on %s", resourceName)
		}
		applicationID := rs.Primary.Attributes["application_id"]
		if applicationID == "" {
			return "", fmt.Errorf("application_id not set on %s", resourceName)
		}
		return fmt.Sprintf("%s/%s/%s", pipelineID, stageID, applicationID), nil
	}
}
