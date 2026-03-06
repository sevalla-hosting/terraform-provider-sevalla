package apikey_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccAPIKey_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-apikey")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-apikey-updated")
	resourceName := "sevalla_api_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with name only
			{
				Config: testAccAPIKeyConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "company_id"),
					resource.TestCheckResourceAttrSet(resourceName, "source"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "role_ids"},
			},
			// Step 3: Update name
			{
				Config: testAccAPIKeyConfig(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "company_id"),
					resource.TestCheckResourceAttrSet(resourceName, "source"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccAPIKeyDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-apikey-ds")
	resourceName := "sevalla_api_key.test"
	dataSourceName := "data.sevalla_api_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enabled"),
					resource.TestCheckResourceAttrSet(dataSourceName, "company_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "source"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccAPIKeyPermissionsDataSource_basic(t *testing.T) {
	dataSourceName := "data.sevalla_api_key_permissions.all"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyPermissionsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "permissions.#"),
					testCheckAttrGreaterThan(dataSourceName, "permissions.#", 0),
					resource.TestCheckResourceAttrSet(dataSourceName, "permissions.0.name"),
				),
			},
		},
	})
}

func TestAccAPIKeyRolesDataSource_basic(t *testing.T) {
	dataSourceName := "data.sevalla_api_key_roles.all"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyRolesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "roles.#"),
					testCheckAttrGreaterThan(dataSourceName, "roles.#", 0),
					resource.TestCheckResourceAttrSet(dataSourceName, "roles.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "roles.0.name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "roles.0.description"),
				),
			},
		},
	})
}

func testAccAPIKeyConfig(name string) string {
	return fmt.Sprintf(`
data "sevalla_api_key_roles" "all" {}

resource "sevalla_api_key" "test" {
  name     = %q
  role_ids = [data.sevalla_api_key_roles.all.roles[0].id]
}
`, name)
}

func testAccAPIKeyDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "sevalla_api_key_roles" "all" {}

resource "sevalla_api_key" "test" {
  name     = %q
  role_ids = [data.sevalla_api_key_roles.all.roles[0].id]
}

data "sevalla_api_key" "test" {
  id = sevalla_api_key.test.id
}
`, name)
}

func testAccAPIKeyPermissionsDataSourceConfig() string {
	return `
data "sevalla_api_key_permissions" "all" {}
`
}

func testAccAPIKeyRolesDataSourceConfig() string {
	return `
data "sevalla_api_key_roles" "all" {}
`
}

// testCheckAttrGreaterThan returns a TestCheckFunc that verifies the given
// attribute's integer value is strictly greater than min.
func testCheckAttrGreaterThan(name, key string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		val, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("attribute %q not found on resource %s", key, name)
		}

		intVal, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("attribute %q value %q is not an integer: %s", key, val, err)
		}

		if intVal <= min {
			return fmt.Errorf("expected attribute %q to be greater than %d, got %d", key, min, intVal)
		}

		return nil
	}
}
