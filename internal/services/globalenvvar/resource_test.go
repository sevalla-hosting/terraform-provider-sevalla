package globalenvvar_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccGlobalEnvironmentVariable_basic(t *testing.T) {
	rSuffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	envKey := fmt.Sprintf("TF_ACC_TEST_%s", rSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccGlobalEnvVarConfig_basic(envKey, "initial-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_global_environment_variable.test", "id"),
					resource.TestCheckResourceAttr("sevalla_global_environment_variable.test", "key", envKey),
					resource.TestCheckResourceAttr("sevalla_global_environment_variable.test", "value", "initial-value"),
				),
			},
			// Import
			{
				ResourceName:      "sevalla_global_environment_variable.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update value
			{
				Config: testAccGlobalEnvVarConfig_basic(envKey, "updated-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_global_environment_variable.test", "key", envKey),
					resource.TestCheckResourceAttr("sevalla_global_environment_variable.test", "value", "updated-value"),
				),
			},
		},
	})
}

func testAccGlobalEnvVarConfig_basic(key, value string) string {
	return fmt.Sprintf(`
resource "sevalla_global_environment_variable" "test" {
  key   = %[1]q
  value = %[2]q
}
`, key, value)
}
