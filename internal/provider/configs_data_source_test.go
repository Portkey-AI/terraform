package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_configs.all", "configs.#"),
				),
			},
		},
	})
}

func TestAccConfigsDataSource_withWorkspace(t *testing.T) {
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigsDataSourceConfigWithWorkspace(workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_configs.workspace", "configs.#"),
				),
			},
		},
	})
}

func testAccConfigsDataSourceConfig() string {
	return `
provider "portkey" {}

data "portkey_configs" "all" {}
`
}

func testAccConfigsDataSourceConfigWithWorkspace(workspaceID string) string {
	return `
provider "portkey" {}

data "portkey_configs" "workspace" {
  workspace_id = "` + workspaceID + `"
}
`
}
