package referencedata_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/testutil"
)

func TestAccClustersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "sevalla_clusters" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sevalla_clusters.all", "clusters.#"),
					resource.TestCheckResourceAttrSet("data.sevalla_clusters.all", "clusters.0.id"),
					resource.TestCheckResourceAttrSet("data.sevalla_clusters.all", "clusters.0.name"),
					resource.TestCheckResourceAttrSet("data.sevalla_clusters.all", "clusters.0.location"),
				),
			},
		},
	})
}

func TestAccProcessResourceTypesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "sevalla_process_resource_types" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sevalla_process_resource_types.all", "process_resource_types.#"),
					resource.TestCheckResourceAttrSet("data.sevalla_process_resource_types.all", "process_resource_types.0.id"),
					resource.TestCheckResourceAttrSet("data.sevalla_process_resource_types.all", "process_resource_types.0.name"),
					resource.TestCheckResourceAttrSet("data.sevalla_process_resource_types.all", "process_resource_types.0.cpu_limit"),
					resource.TestCheckResourceAttrSet("data.sevalla_process_resource_types.all", "process_resource_types.0.memory_limit"),
				),
			},
		},
	})
}

func TestAccDatabaseResourceTypesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "sevalla_database_resource_types" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sevalla_database_resource_types.all", "database_resource_types.#"),
					resource.TestCheckResourceAttrSet("data.sevalla_database_resource_types.all", "database_resource_types.0.id"),
					resource.TestCheckResourceAttrSet("data.sevalla_database_resource_types.all", "database_resource_types.0.name"),
					resource.TestCheckResourceAttrSet("data.sevalla_database_resource_types.all", "database_resource_types.0.cpu_limit"),
					resource.TestCheckResourceAttrSet("data.sevalla_database_resource_types.all", "database_resource_types.0.memory_limit"),
					resource.TestCheckResourceAttrSet("data.sevalla_database_resource_types.all", "database_resource_types.0.storage_limit"),
				),
			},
		},
	})
}

func TestAccUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "sevalla_users" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sevalla_users.all", "users.#"),
					resource.TestCheckResourceAttrSet("data.sevalla_users.all", "users.0.id"),
					resource.TestCheckResourceAttrSet("data.sevalla_users.all", "users.0.email"),
				),
			},
		},
	})
}
