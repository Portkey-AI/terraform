package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkspaceDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig(rName, "Test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the data source returns the correct workspace
					resource.TestCheckResourceAttr("data.portkey_workspace.test", "name", rName),
					resource.TestCheckResourceAttr("data.portkey_workspace.test", "description", "Test description"),
					resource.TestCheckResourceAttrSet("data.portkey_workspace.test", "id"),
					resource.TestCheckResourceAttrSet("data.portkey_workspace.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.portkey_workspace.test", "updated_at"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig(name, description string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = %[2]q
}

data "portkey_workspace" "test" {
  id = portkey_workspace.test.id
}
`, name, description)
}
