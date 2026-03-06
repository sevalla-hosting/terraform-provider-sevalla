package objectstorage_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccObjectStorage_basic(t *testing.T) {
	displayName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))
	updatedDisplayName := fmt.Sprintf("tf-acc-%s-updated", acctest.RandString(8))
	resourceName := "sevalla_object_storage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccObjectStorageConfig(displayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(resourceName, "location", "enam"),
					resource.TestCheckResourceAttr(resourceName, "jurisdiction", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "access_key"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					resource.TestCheckResourceAttrSet(resourceName, "bucket_name"),
					resource.TestCheckResourceAttrSet(resourceName, "company_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_key", "secret_key", "updated_at"},
			},
			// Step 3: Update display_name
			{
				Config: testAccObjectStorageConfig(updatedDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", updatedDisplayName),
					resource.TestCheckResourceAttr(resourceName, "location", "enam"),
					resource.TestCheckResourceAttr(resourceName, "jurisdiction", "default"),
				),
			},
		},
	})
}

func TestAccObjectStorageCORSPolicy_basic(t *testing.T) {
	displayName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))
	resourceName := "sevalla_object_storage_cors_policy.test"
	osResourceName := "sevalla_object_storage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccObjectStorageCORSPolicyConfig(displayName, `["https://example.com"]`, `["GET"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "object_storage_id", osResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_origins.0", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "allowed_methods.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_methods.0", "GET"),
				),
			},
			// Step 2: Import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					osID := rs.Primary.Attributes["object_storage_id"]
					policyID := rs.Primary.ID
					return fmt.Sprintf("%s/%s", osID, policyID), nil
				},
			},
			// Step 3: Update origins and methods
			{
				Config: testAccObjectStorageCORSPolicyConfig(displayName, `["https://example.com", "https://other.com"]`, `["GET", "PUT"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "allowed_origins.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_origins.0", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "allowed_origins.1", "https://other.com"),
					resource.TestCheckResourceAttr(resourceName, "allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_methods.0", "GET"),
					resource.TestCheckResourceAttr(resourceName, "allowed_methods.1", "PUT"),
				),
			},
		},
	})
}

func TestAccObjectStorageDataSource_basic(t *testing.T) {
	displayName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))
	resourceName := "sevalla_object_storage.test"
	dataSourceName := "data.sevalla_object_storage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectStorageDataSourceConfig(displayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "location", resourceName, "location"),
					resource.TestCheckResourceAttrPair(dataSourceName, "jurisdiction", resourceName, "jurisdiction"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint", resourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_name", resourceName, "bucket_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "company_id", resourceName, "company_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_at", resourceName, "created_at"),
				),
			},
		},
	})
}

func testAccObjectStorageConfig(displayName string) string {
	return fmt.Sprintf(`
resource "sevalla_object_storage" "test" {
  display_name = %q
}
`, displayName)
}

func testAccObjectStorageCORSPolicyConfig(displayName, allowedOrigins, allowedMethods string) string {
	return fmt.Sprintf(`
resource "sevalla_object_storage" "test" {
  display_name = %q
}

resource "sevalla_object_storage_cors_policy" "test" {
  object_storage_id = sevalla_object_storage.test.id
  allowed_origins   = %s
  allowed_methods   = %s
}
`, displayName, allowedOrigins, allowedMethods)
}

func testAccObjectStorageDataSourceConfig(displayName string) string {
	return fmt.Sprintf(`
resource "sevalla_object_storage" "test" {
  display_name = %q
}

data "sevalla_object_storage" "test" {
  id = sevalla_object_storage.test.id
}
`, displayName)
}
