package database_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// TestAccDatabaseInternalConnection_basic tests creating and importing an internal
// connection between a database and an application on the same cluster.
func TestAccDatabaseInternalConnection_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create database, application, and internal connection
			{
				Config: testAccDatabaseInternalConnectionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_database_internal_connection.test", "id"),
					resource.TestCheckResourceAttrPair(
						"sevalla_database_internal_connection.test", "database_id",
						"sevalla_database.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"sevalla_database_internal_connection.test", "target_id",
						"sevalla_application.test", "id",
					),
					resource.TestCheckResourceAttr("sevalla_database_internal_connection.test", "target_type", "app"),
					resource.TestCheckResourceAttrSet("sevalla_database_internal_connection.test", "source_type"),
					resource.TestCheckResourceAttrSet("sevalla_database_internal_connection.test", "source_display_name"),
					resource.TestCheckResourceAttrSet("sevalla_database_internal_connection.test", "target_display_name"),
				),
			},
			// Step 2: Import by database_id/connection_id
			{
				ResourceName:            "sevalla_database_internal_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"target_type"},
			},
		},
	})
}

func testAccDatabaseInternalConnectionConfig(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}
data "sevalla_database_resource_types" "all" {}

resource "sevalla_application" "test" {
  display_name = "%[1]s-app"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/kotapeter/pack-demo"
}

resource "sevalla_database" "test" {
  display_name     = "%[1]s-db"
  type             = "postgresql"
  version          = "16"
  cluster_id       = data.sevalla_clusters.all.clusters[0].id
  resource_type_id = data.sevalla_database_resource_types.all.database_resource_types[0].id
  db_name          = "testdb"
  db_password      = "TestPass1234"
}

resource "sevalla_database_internal_connection" "test" {
  database_id = sevalla_database.test.id
  target_id   = sevalla_application.test.id
  target_type = "app"
}
`, name)
}

// TestAccDatabaseIPRestriction_basic tests creating, importing, and updating
// IP restrictions on a database.
func TestAccDatabaseIPRestriction_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with a single IP
			{
				Config: testAccDatabaseIPRestrictionConfig(rName, true, `"1.2.3.4/32"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"sevalla_database_ip_restriction.test", "database_id",
						"sevalla_database.test", "id",
					),
					resource.TestCheckResourceAttr("sevalla_database_ip_restriction.test", "type", "allow"),
					resource.TestCheckResourceAttr("sevalla_database_ip_restriction.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("sevalla_database_ip_restriction.test", "ip_list.#", "1"),
					resource.TestCheckResourceAttr("sevalla_database_ip_restriction.test", "ip_list.0", "1.2.3.4/32"),
				),
			},
			// Step 2: Import by database_id
			{
				ResourceName:                         "sevalla_database_ip_restriction.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "database_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sevalla_database_ip_restriction.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["database_id"], nil
				},
			},
			// Step 3: Update to add a second IP address
			{
				Config: testAccDatabaseIPRestrictionConfig(rName, true, `"1.2.3.4/32", "10.0.0.0/8"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_database_ip_restriction.test", "type", "allow"),
					resource.TestCheckResourceAttr("sevalla_database_ip_restriction.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("sevalla_database_ip_restriction.test", "ip_list.#", "2"),
				),
			},
		},
	})
}

func testAccDatabaseIPRestrictionConfig(name string, isEnabled bool, ipListEntries string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}
data "sevalla_database_resource_types" "all" {}

resource "sevalla_database" "test" {
  display_name     = "%[1]s"
  type             = "postgresql"
  version          = "16"
  cluster_id       = data.sevalla_clusters.all.clusters[0].id
  resource_type_id = data.sevalla_database_resource_types.all.database_resource_types[0].id
  db_name          = "testdb"
  db_password      = "TestPass1234"
}

resource "sevalla_database_ip_restriction" "test" {
  database_id = sevalla_database.test.id
  type        = "allow"
  is_enabled  = %[2]t
  ip_list     = [%[3]s]
}
`, name, isEnabled, ipListEntries)
}

// TestAccDatabaseDataSource_basic tests reading a single database via its data source.
func TestAccDatabaseDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.sevalla_database.test", "id",
						"sevalla_database.test", "id",
					),
					resource.TestCheckResourceAttr("data.sevalla_database.test", "display_name", rName),
					resource.TestCheckResourceAttr("data.sevalla_database.test", "type", "postgresql"),
					resource.TestCheckResourceAttr("data.sevalla_database.test", "version", "16"),
					resource.TestCheckResourceAttr("data.sevalla_database.test", "db_name", "testdb"),
					resource.TestCheckResourceAttrSet("data.sevalla_database.test", "cluster_id"),
					resource.TestCheckResourceAttrSet("data.sevalla_database.test", "resource_type_id"),
				),
			},
		},
	})
}

func testAccDatabaseDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}
data "sevalla_database_resource_types" "all" {}

resource "sevalla_database" "test" {
  display_name     = "%[1]s"
  type             = "postgresql"
  version          = "16"
  cluster_id       = data.sevalla_clusters.all.clusters[0].id
  resource_type_id = data.sevalla_database_resource_types.all.database_resource_types[0].id
  db_name          = "testdb"
  db_password      = "TestPass1234"
}

data "sevalla_database" "test" {
  id = sevalla_database.test.id
}
`, name)
}

// TestAccDatabasesDataSource_basic tests listing all databases via the plural data source.
func TestAccDatabasesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sevalla_databases.all", "databases.#"),
				),
			},
		},
	})
}

func testAccDatabasesDataSourceConfig() string {
	return `
data "sevalla_databases" "all" {}
`
}
