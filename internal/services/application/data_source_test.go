package application_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccApplicationDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sevalla_application.test", "id"),
					resource.TestCheckResourceAttr("data.sevalla_application.test", "display_name", rName),
					resource.TestCheckResourceAttr("data.sevalla_application.test", "source", "publicGit"),
					resource.TestCheckResourceAttrSet("data.sevalla_application.test", "name"),
					resource.TestCheckResourceAttrSet("data.sevalla_application.test", "created_at"),
				),
			},
		},
	})
}

func TestAccApplicationsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "sevalla_applications" "all" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.sevalla_applications.all", "applications.#"),
			},
		},
	})
}

func testAccApplicationDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "sevalla_clusters" "all" {}

resource "sevalla_application" "test" {
  display_name = %[1]q
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/kotapeter/pack-demo"
}

data "sevalla_application" "test" {
  id = sevalla_application.test.id
}
`, name)
}
