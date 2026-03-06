package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccLoadBalancer_basic(t *testing.T) {
	displayName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	updatedDisplayName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_load_balancer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create and verify
			{
				Config: testAccLoadBalancerConfig(displayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
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
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			// Step 3: Update display_name
			{
				Config: testAccLoadBalancerConfig(updatedDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", updatedDisplayName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "company_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccLoadBalancerDestination_basic(t *testing.T) {
	lbDisplayName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_load_balancer_destination.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create LB + external destination
			{
				Config: testAccLoadBalancerDestinationConfig(lbDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_id", "sevalla_load_balancer.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "EXTERNAL"),
					resource.TestCheckResourceAttr(resourceName, "url", "https://example.com"),
					resource.TestCheckResourceAttrSet(resourceName, "is_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Step 2: Import with composite ID
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["load_balancer_id"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func TestAccLoadBalancerDataSource_basic(t *testing.T) {
	displayName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "sevalla_load_balancer.test"
	dataSourceName := "data.sevalla_load_balancer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerDataSourceConfig(displayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "company_id", resourceName, "company_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_at", resourceName, "created_at"),
				),
			},
		},
	})
}

func TestAccLoadBalancersDataSource_basic(t *testing.T) {
	displayName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	dataSourceName := "data.sevalla_load_balancers.all"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancersDataSourceConfig(displayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "load_balancers.#"),
				),
			},
		},
	})
}

func testAccLoadBalancerConfig(displayName string) string {
	return fmt.Sprintf(`
resource "sevalla_load_balancer" "test" {
  display_name = %[1]q
}
`, displayName)
}

func testAccLoadBalancerDestinationConfig(lbDisplayName string) string {
	return fmt.Sprintf(`
resource "sevalla_load_balancer" "test" {
  display_name = %[1]q
}

resource "sevalla_load_balancer_destination" "test" {
  load_balancer_id = sevalla_load_balancer.test.id
  service_type     = "EXTERNAL"
  url              = "https://example.com"
}
`, lbDisplayName)
}

func testAccLoadBalancerDataSourceConfig(displayName string) string {
	return fmt.Sprintf(`
resource "sevalla_load_balancer" "test" {
  display_name = %[1]q
}

data "sevalla_load_balancer" "test" {
  id = sevalla_load_balancer.test.id
}
`, displayName)
}

func testAccLoadBalancersDataSourceConfig(displayName string) string {
	return fmt.Sprintf(`
resource "sevalla_load_balancer" "test" {
  display_name = %[1]q
}

data "sevalla_load_balancers" "all" {
  depends_on = [sevalla_load_balancer.test]
}
`, displayName)
}
