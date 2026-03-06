package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccWebhookDataSource_basic(t *testing.T) {
	rSuffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	webhookURL := fmt.Sprintf("https://example.com/webhook/tf-acc-test-%s", rSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookDataSourceConfig(webhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.sevalla_webhook.test", "id",
						"sevalla_webhook.test", "id",
					),
					resource.TestCheckResourceAttr("data.sevalla_webhook.test", "endpoint", webhookURL),
				),
			},
		},
	})
}

func testAccWebhookDataSourceConfig(url string) string {
	return fmt.Sprintf(`
resource "sevalla_webhook" "test" {
  endpoint       = %[1]q
  allowed_events = ["APP_DEPLOY"]
}

data "sevalla_webhook" "test" {
  id = sevalla_webhook.test.id
}
`, url)
}
