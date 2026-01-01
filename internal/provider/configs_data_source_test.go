package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
	rName := acctest.RandomWithPrefix("tf-acc-ds-list")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a config first, then list - ensures list has at least 1 item
				Config: testAccConfigsDataSourceConfigWithWorkspaceAndConfig(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.portkey_configs.workspace", "workspace_id", workspaceID),
					// Verify at least 1 config exists
					resource.TestCheckResourceAttr("data.portkey_configs.workspace", "configs.#", "1"),
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

func testAccConfigsDataSourceConfigWithWorkspaceAndConfig(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_config" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  config       = jsonencode({
    retry = { attempts = 3 }
  })
}

data "portkey_configs" "workspace" {
  workspace_id = %[2]q
  depends_on   = [portkey_config.test]
}
`, name, workspaceID)
}
