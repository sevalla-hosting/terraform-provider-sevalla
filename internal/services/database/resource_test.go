package database_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccDatabase_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_database.test", "id"),
					resource.TestCheckResourceAttr("sevalla_database.test", "display_name", rName),
					resource.TestCheckResourceAttr("sevalla_database.test", "type", "postgresql"),
					resource.TestCheckResourceAttrSet("sevalla_database.test", "name"),
					resource.TestCheckResourceAttrSet("sevalla_database.test", "status"),
					resource.TestCheckResourceAttrSet("sevalla_database.test", "cluster_id"),
					resource.TestCheckResourceAttrSet("sevalla_database.test", "resource_type_id"),
					resource.TestCheckResourceAttrSet("sevalla_database.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:            "sevalla_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"db_password", "updated_at"},
			},
			// Update display_name
			{
				Config: testAccDatabaseConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_database.test", "display_name", rName+"-updated"),
				),
			},
		},
	})
}

func testAccDatabaseConfig_basic(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}

data "sevalla_database_resource_types" "all" {}

resource "sevalla_database" "test" {
  display_name     = %[1]q
  type             = "postgresql"
  version          = "16"
  cluster_id       = data.sevalla_clusters.all.clusters[0].id
  resource_type_id = data.sevalla_database_resource_types.all.database_resource_types[0].id
  db_name          = "testdb"
  db_password      = "TestPass1234"
}
`, name)
}

func testAccDatabaseConfig_updated(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}

data "sevalla_database_resource_types" "all" {}

resource "sevalla_database" "test" {
  display_name     = "%[1]s-updated"
  type             = "postgresql"
  version          = "16"
  cluster_id       = data.sevalla_clusters.all.clusters[0].id
  resource_type_id = data.sevalla_database_resource_types.all.database_resource_types[0].id
  db_name          = "testdb"
  db_password      = "TestPass1234"
}
`, name)
}
