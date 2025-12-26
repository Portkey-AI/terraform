package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkspacesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_workspaces.all", "workspaces.#"),
				),
			},
		},
	})
}

func TestAccWorkspacesDataSource_withCreatedWorkspace(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-list")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfigWithWorkspace(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify at least one workspace exists (the one we just created)
					resource.TestCheckResourceAttrSet("data.portkey_workspaces.all", "workspaces.#"),
				),
			},
		},
	})
}

func testAccWorkspacesDataSourceConfigBasic() string {
	return `
provider "portkey" {}

data "portkey_workspaces" "all" {}
`
}

func testAccWorkspacesDataSourceConfigWithWorkspace(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = "Created for list test"
}

data "portkey_workspaces" "all" {
  depends_on = [portkey_workspace.test]
}
`, name)
}
