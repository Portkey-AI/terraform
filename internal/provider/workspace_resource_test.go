package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkspaceResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWorkspaceResourceConfig(rName, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_workspace.test", "id"),
					resource.TestCheckResourceAttr("portkey_workspace.test", "name", rName),
					resource.TestCheckResourceAttr("portkey_workspace.test", "description", "Initial description"),
					resource.TestCheckResourceAttrSet("portkey_workspace.test", "created_at"),
					resource.TestCheckResourceAttrSet("portkey_workspace.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "portkey_workspace.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			// Update testing
			{
				Config: testAccWorkspaceResourceConfig(rName+"-updated", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace.test", "name", rName+"-updated"),
					resource.TestCheckResourceAttr("portkey_workspace.test", "description", "Updated description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccWorkspaceResource_minimal tests workspace creation without optional fields.
func TestAccWorkspaceResource_minimal(t *testing.T) {

	rName := acctest.RandomWithPrefix("tf-acc-minimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfigMinimal(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_workspace.test", "id"),
					resource.TestCheckResourceAttr("portkey_workspace.test", "name", rName),
					resource.TestCheckResourceAttrSet("portkey_workspace.test", "created_at"),
				),
			},
		},
	})
}

func TestAccWorkspaceResource_updateName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-rename")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfig(rName, "Initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace.test", "name", rName),
				),
			},
			{
				Config: testAccWorkspaceResourceConfig(rName+"-renamed", "Initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace.test", "name", rName+"-renamed"),
				),
			},
		},
	})
}

func TestAccWorkspaceResource_updateDescription(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfig(rName, "Original description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace.test", "description", "Original description"),
				),
			},
			{
				Config: testAccWorkspaceResourceConfig(rName, "Modified description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace.test", "description", "Modified description"),
				),
			},
			{
				Config: testAccWorkspaceResourceConfig(rName, "Final description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace.test", "description", "Final description"),
				),
			},
		},
	})
}

func testAccWorkspaceResourceConfig(name, description string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccWorkspaceResourceConfigMinimal(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name = %[1]q
}
`, name)
}
