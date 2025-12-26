package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae" // Test workspace

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_config.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_config.test", "slug"),
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
					resource.TestCheckResourceAttr("portkey_config.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("portkey_config.test", "status", "active"),
					resource.TestCheckResourceAttrSet("portkey_config.test", "version_id"),
					resource.TestCheckResourceAttrSet("portkey_config.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "portkey_config.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			// Update testing - change config values
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":5},"cache":{"mode":"simple"}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccConfigResource_updateName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-rename")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
				),
			},
			{
				Config: testAccConfigResourceConfig(rName+"-renamed", workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName+"-renamed"),
				),
			},
		},
	})
}

func testAccConfigResourceConfig(name, workspaceID, config string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_config" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  config       = %[3]q
}
`, name, workspaceID, config)
}

