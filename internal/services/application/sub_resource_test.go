package application_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

// baseApplicationConfig returns the shared HCL that creates a sevalla_application
// used by all sub-resource tests.
func baseApplicationConfig(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}

resource "sevalla_application" "test" {
  display_name = %q
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/kotapeter/pack-demo"
}
`, name)
}

// ---------------------------------------------------------------------------
// Environment Variable
// ---------------------------------------------------------------------------

func TestAccApplicationEnvironmentVariable_basic(t *testing.T) {
	appName := fmt.Sprintf("tf-acc-env-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_application_environment_variable.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccApplicationEnvironmentVariableConfig(appName, "TF_ACC_TEST_KEY", "initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "sevalla_application.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "key", "TF_ACC_TEST_KEY"),
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
				ImportStateIdFunc:       testAccApplicationSubResourceImportStateIdFunc(resourceName, "application_id", "id"),
			},
			// Step 3: Update value
			{
				Config: testAccApplicationEnvironmentVariableConfig(appName, "TF_ACC_TEST_KEY", "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "TF_ACC_TEST_KEY"),
					resource.TestCheckResourceAttr(resourceName, "value", "updated"),
				),
			},
		},
	})
}

func testAccApplicationEnvironmentVariableConfig(appName, key, value string) string {
	return baseApplicationConfig(appName) + fmt.Sprintf(`
resource "sevalla_application_environment_variable" "test" {
  application_id = sevalla_application.test.id
  key            = %q
  value          = %q
}
`, key, value)
}

// ---------------------------------------------------------------------------
// Domain
// ---------------------------------------------------------------------------

func TestAccApplicationDomain_basic(t *testing.T) {
	appName := fmt.Sprintf("tf-acc-domain-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	domainName := fmt.Sprintf("tf-acc-%s.example.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_application_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccApplicationDomainConfig(appName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "sevalla_application.test", "id"),
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
				ImportStateIdFunc:       testAccApplicationSubResourceImportStateIdFunc(resourceName, "application_id", "id"),
			},
		},
	})
}

func testAccApplicationDomainConfig(appName, domainName string) string {
	return baseApplicationConfig(appName) + fmt.Sprintf(`
resource "sevalla_application_domain" "test" {
  application_id = sevalla_application.test.id
  name           = %q
}
`, domainName)
}

// ---------------------------------------------------------------------------
// Deployment Hook
// ---------------------------------------------------------------------------

func TestAccApplicationDeploymentHook_basic(t *testing.T) {
	appName := fmt.Sprintf("tf-acc-hook-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_application_deployment_hook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccApplicationDeploymentHookConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "sevalla_application.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
			// Step 2: Import
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"url"},
				ImportStateVerifyIdentifierAttribute: "application_id",
				ImportStateIdFunc:                    testAccApplicationDeploymentHookImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccApplicationDeploymentHookConfig(appName string) string {
	return baseApplicationConfig(appName) + `
resource "sevalla_application_deployment_hook" "test" {
  application_id = sevalla_application.test.id
}
`
}

// testAccApplicationDeploymentHookImportStateIdFunc returns the application_id
// as the import identifier since deployment hooks are imported by application_id alone.
func testAccApplicationDeploymentHookImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		applicationID := rs.Primary.Attributes["application_id"]
		if applicationID == "" {
			return "", fmt.Errorf("application_id not set on %s", resourceName)
		}
		return applicationID, nil
	}
}

// ---------------------------------------------------------------------------
// IP Restriction
// ---------------------------------------------------------------------------

func TestAccApplicationIPRestriction_basic(t *testing.T) {
	appName := fmt.Sprintf("tf-acc-iprestr-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_application_ip_restriction.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with one IP
			{
				Config: testAccApplicationIPRestrictionConfig_oneIP(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "sevalla_application.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "allow"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_list.0", "1.2.3.4/32"),
				),
			},
			// Step 2: Import
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "application_id",
				ImportStateIdFunc:                    testAccApplicationIPRestrictionImportStateIdFunc(resourceName),
			},
			// Step 3: Update to two IPs
			{
				Config: testAccApplicationIPRestrictionConfig_twoIPs(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "allow"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_list.#", "2"),
				),
			},
		},
	})
}

func testAccApplicationIPRestrictionConfig_oneIP(appName string) string {
	return baseApplicationConfig(appName) + `
resource "sevalla_application_ip_restriction" "test" {
  application_id = sevalla_application.test.id
  type           = "allow"
  is_enabled     = true
  ip_list        = ["1.2.3.4/32"]
}
`
}

func testAccApplicationIPRestrictionConfig_twoIPs(appName string) string {
	return baseApplicationConfig(appName) + `
resource "sevalla_application_ip_restriction" "test" {
  application_id = sevalla_application.test.id
  type           = "allow"
  is_enabled     = true
  ip_list        = ["1.2.3.4/32", "5.6.7.8/32"]
}
`
}

// testAccApplicationIPRestrictionImportStateIdFunc returns the application_id
// as the import identifier since IP restrictions use application_id as their identifier.
func testAccApplicationIPRestrictionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		applicationID := rs.Primary.Attributes["application_id"]
		if applicationID == "" {
			return "", fmt.Errorf("application_id not set on %s", resourceName)
		}
		return applicationID, nil
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testAccApplicationSubResourceImportStateIdFunc builds a composite import ID
// in the format "application_id/sub_resource_id" from the given attribute names.
func testAccApplicationSubResourceImportStateIdFunc(resourceName, appIDAttr, subIDAttr string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		applicationID := rs.Primary.Attributes[appIDAttr]
		if applicationID == "" {
			return "", fmt.Errorf("%s not set on %s", appIDAttr, resourceName)
		}
		subResourceID := rs.Primary.Attributes[subIDAttr]
		if subResourceID == "" {
			return "", fmt.Errorf("%s not set on %s", subIDAttr, resourceName)
		}
		return fmt.Sprintf("%s/%s", applicationID, subResourceID), nil
	}
}
