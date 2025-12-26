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
// Note: This test may be affected by API behavior where org admins have elevated workspace roles.
func TestAccWorkspaceMemberResource_basic(t *testing.T) {
	t.Skip("Skipping: API getMember endpoint has inconsistent role handling for org admins")

	rName := acctest.RandomWithPrefix("tf-acc-wsmember")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create workspace and add member using first user from org
			{
				Config: testAccWorkspaceMemberResourceConfigDynamic(rName, "member"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "workspace_id"),
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "user_id"),
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "role"),
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "created_at"),
				),
			},
			// Import testing
			{
				ResourceName:            "portkey_workspace_member.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "role"},
				ImportStateIdFunc:       workspaceMemberImportStateIdFunc("portkey_workspace_member.test"),
			},
		},
	})
}

// TestAccWorkspaceMemberResource_roleUpdate tests changing member roles.
// Note: This test may be affected by API behavior where org admins have elevated workspace roles.
func TestAccWorkspaceMemberResource_roleUpdate(t *testing.T) {
	t.Skip("Skipping: API getMember endpoint has inconsistent role handling for org admins")

	rName := acctest.RandomWithPrefix("tf-acc-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceMemberResourceConfigDynamic(rName, "member"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "role"),
				),
			},
			{
				Config: testAccWorkspaceMemberResourceConfigDynamic(rName, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_workspace_member.test", "role"),
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
		memberID := rs.Primary.ID
		return fmt.Sprintf("%s/%s", workspaceID, memberID), nil
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
