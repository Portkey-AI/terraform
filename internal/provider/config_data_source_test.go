package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-ds")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigDataSourceConfig(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_config.test", "id"),
					resource.TestCheckResourceAttr("data.portkey_config.test", "name", rName),
					resource.TestCheckResourceAttr("data.portkey_config.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("data.portkey_config.test", "status", "active"),
					resource.TestCheckResourceAttrSet("data.portkey_config.test", "config"),
				),
			},
		},
	})
}

func testAccConfigDataSourceConfig(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_config" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  config       = "{\"retry\":{\"attempts\":3}}"
}

data "portkey_config" "test" {
  slug = portkey_config.test.slug
}
`, name, workspaceID)
}
