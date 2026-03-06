package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccWebhook_basic(t *testing.T) {
	rSuffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	webhookURL := fmt.Sprintf("https://example.com/webhook/tf-acc-test-%s", rSuffix)
	webhookURLUpdated := fmt.Sprintf("https://example.com/webhook/tf-acc-test-%s-updated", rSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWebhookConfig_basic(webhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sevalla_webhook.test", "id"),
					resource.TestCheckResourceAttr("sevalla_webhook.test", "endpoint", webhookURL),
					resource.TestCheckResourceAttrSet("sevalla_webhook.test", "is_enabled"),
					resource.TestCheckResourceAttrSet("sevalla_webhook.test", "created_at"),
					resource.TestCheckResourceAttr("sevalla_webhook.test", "allowed_events.#", "1"),
				),
			},
			// Import
			{
				ResourceName:      "sevalla_webhook.test",
				ImportState:       true,
				ImportStateVerify: true,
				// secret may not be returned on subsequent reads
				ImportStateVerifyIgnore: []string{"secret"},
			},
			// Update URL
			{
				Config: testAccWebhookConfig_updated(webhookURLUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sevalla_webhook.test", "endpoint", webhookURLUpdated),
					resource.TestCheckResourceAttr("sevalla_webhook.test", "allowed_events.#", "2"),
				),
			},
		},
	})
}

func testAccWebhookConfig_basic(url string) string {
	return fmt.Sprintf(`
resource "sevalla_webhook" "test" {
  endpoint       = %[1]q
  allowed_events = ["APP_DEPLOY"]
}
`, url)
}

func testAccWebhookConfig_updated(url string) string {
	return fmt.Sprintf(`
resource "sevalla_webhook" "test" {
  endpoint       = %[1]q
  allowed_events = ["APP_DEPLOY", "APP_CREATE"]
}
`, url)
}
