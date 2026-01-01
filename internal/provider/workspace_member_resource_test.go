package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccWorkspaceMemberResource_basic tests the basic workspace member lifecycle.
// Uses the first user from the organization's user list.
// Note: Org owners always get "admin" role in workspaces regardless of the requested role.
func TestAccWorkspaceMemberResource_basic(t *testing.T) {

	rName := acctest.RandomWithPrefix("tf-acc-wsmember")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create workspace and add member using first user from org
			// Using "admin" role since org owners automatically get admin access
			{
				Config: testAccWorkspaceMemberResourceConfigDynamic(rName, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "workspace_id"),
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "user_id"),
					resource.TestCheckResourceAttr("portkey_workspace_member.test", "role", "admin"),
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "created_at"),
				),
			},
			// Import testing
			{
				ResourceName:            "portkey_workspace_member.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at"},
				ImportStateIdFunc:       workspaceMemberImportStateIdFunc("portkey_workspace_member.test"),
			},
		},
	})
}

// TestAccWorkspaceMemberResource_roleUpdate tests changing member roles.
// Note: This test is skipped because org owners always have "admin" role in workspaces.
// To properly test role updates, you need a non-owner user in the organization.
func TestAccWorkspaceMemberResource_roleUpdate(t *testing.T) {
	t.Skip("Skipping: org owners always have admin role in workspaces; need non-owner user to test role changes")

	rName := acctest.RandomWithPrefix("tf-acc-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceMemberResourceConfigDynamic(rName, "member"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace_member.test", "role", "member"),
				),
			},
			{
				Config: testAccWorkspaceMemberResourceConfigDynamic(rName, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace_member.test", "role", "admin"),
				),
			},
		},
	})
}

// workspaceMemberImportStateIdFunc returns a function that generates the import ID
func workspaceMemberImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		workspaceID := rs.Primary.Attributes["workspace_id"]
		userID := rs.Primary.Attributes["user_id"]
		return fmt.Sprintf("%s/%s", workspaceID, userID), nil
	}
}

// testAccWorkspaceMemberResourceConfigDynamic creates a workspace member using
// the first user from the organization's user list (fetched via data source)
func testAccWorkspaceMemberResourceConfigDynamic(workspaceName, role string) string {
	return fmt.Sprintf(`
provider "portkey" {}

# Get existing users from the organization
data "portkey_users" "all" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = "Test workspace for member testing"
}

resource "portkey_workspace_member" "test" {
  workspace_id = portkey_workspace.test.id
  user_id      = data.portkey_users.all.users[0].id
  role         = %[2]q
}
`, workspaceName, role)
}
