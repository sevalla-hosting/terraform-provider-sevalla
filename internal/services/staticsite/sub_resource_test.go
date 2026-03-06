package staticsite_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// baseStaticSiteConfig returns the shared HCL that creates a sevalla_static_site
// used by all sub-resource tests.
func baseStaticSiteConfig(name string) string {
	return fmt.Sprintf(`
resource "sevalla_static_site" "test" {
  display_name   = %[1]q
  source         = "publicGit"
  repo_url       = "https://github.com/kotapeter/pack-demo"
  default_branch = "main"
}
`, name)
}

// ---------------------------------------------------------------------------
// Environment Variable
// ---------------------------------------------------------------------------

func TestAccStaticSiteEnvironmentVariable_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-env-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_static_site_environment_variable.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccStaticSiteEnvironmentVariableConfig(rName, "TF_TEST_VAR", "initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "static_site_id", "sevalla_static_site.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "key", "TF_TEST_VAR"),
					resource.TestCheckResourceAttr(resourceName, "value", "initial"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
				ImportStateIdFunc:       testAccStaticSiteSubResourceImportStateIdFunc(resourceName, "static_site_id", "id"),
			},
			// Step 3: Update value
			{
				Config: testAccStaticSiteEnvironmentVariableConfig(rName, "TF_TEST_VAR", "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "TF_TEST_VAR"),
					resource.TestCheckResourceAttr(resourceName, "value", "updated"),
				),
			},
		},
	})
}

func testAccStaticSiteEnvironmentVariableConfig(name, key, value string) string {
	return baseStaticSiteConfig(name) + fmt.Sprintf(`
resource "sevalla_static_site_environment_variable" "test" {
  static_site_id = sevalla_static_site.test.id
  key            = %q
  value          = %q
}
`, key, value)
}

// ---------------------------------------------------------------------------
// Domain
// ---------------------------------------------------------------------------

func TestAccStaticSiteDomain_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-domain-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	domainName := fmt.Sprintf("tf-acc-%s.example.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_static_site_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccStaticSiteDomainConfig(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "static_site_id", "sevalla_static_site.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "type"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "custom_ssl_cert", "custom_ssl_key"},
				ImportStateIdFunc:       testAccStaticSiteSubResourceImportStateIdFunc(resourceName, "static_site_id", "id"),
			},
		},
	})
}

func testAccStaticSiteDomainConfig(name, domainName string) string {
	return baseStaticSiteConfig(name) + fmt.Sprintf(`
resource "sevalla_static_site_domain" "test" {
  static_site_id = sevalla_static_site.test.id
  name           = %q
}
`, domainName)
}

// ---------------------------------------------------------------------------
// Data Sources
// ---------------------------------------------------------------------------

func TestAccStaticSiteDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticSiteDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.sevalla_static_site.test", "id",
						"sevalla_static_site.test", "id",
					),
					resource.TestCheckResourceAttr("data.sevalla_static_site.test", "display_name", rName),
				),
			},
		},
	})
}

func testAccStaticSiteDataSourceConfig(name string) string {
	return baseStaticSiteConfig(name) + `
data "sevalla_static_site" "test" {
  id = sevalla_static_site.test.id
}
`
}

func TestAccStaticSitesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticSitesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sevalla_static_sites.all", "static_sites.#"),
				),
			},
		},
	})
}

func testAccStaticSitesDataSourceConfig() string {
	return `
data "sevalla_static_sites" "all" {}
`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testAccStaticSiteSubResourceImportStateIdFunc builds a composite import ID
// in the format "static_site_id/sub_resource_id" from the given attribute names.
func testAccStaticSiteSubResourceImportStateIdFunc(resourceName, parentIDAttr, subIDAttr string) resource.ImportStateIdFunc {
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
