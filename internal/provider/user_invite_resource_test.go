package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserInviteResource_basic(t *testing.T) {
	rEmail := fmt.Sprintf("tf-acc-test-%s@example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserInviteResourceConfig(rEmail, "member"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_user_invite.test", "id"),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "email", rEmail),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "role", "member"),
					resource.TestCheckResourceAttrSet("portkey_user_invite.test", "created_at"),
					resource.TestCheckResourceAttrSet("portkey_user_invite.test", "expires_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "portkey_user_invite.test",
				ImportState:             true,
				ImportStateVerify:       true,
				// Ignored fields explained:
				// - scopes: API doesn't return scopes on GET invite
				// - status: computed by API, not from config
				// - timestamps: not part of import state
				ImportStateVerifyIgnore: []string{"scopes", "status", "created_at", "expires_at"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccUserInviteResource_admin(t *testing.T) {
	rEmail := fmt.Sprintf("tf-acc-admin-%s@example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserInviteResourceConfig(rEmail, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_user_invite.test", "email", rEmail),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "role", "admin"),
				),
			},
		},
	})
}

func TestAccUserInviteResource_withWorkspace(t *testing.T) {
	rEmail := fmt.Sprintf("tf-acc-ws-%s@example.com", acctest.RandString(8))
	rWorkspaceName := acctest.RandomWithPrefix("tf-acc-invite-ws")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserInviteResourceConfigWithWorkspace(rEmail, rWorkspaceName, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_user_invite.test", "email", rEmail),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "role", "member"),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "workspaces.#", "1"),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "workspaces.0.role", "admin"),
				),
			},
		},
	})
}

func TestAccUserInviteResource_withScopes(t *testing.T) {
	rEmail := fmt.Sprintf("tf-acc-scopes-%s@example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserInviteResourceConfigWithScopes(rEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_user_invite.test", "email", rEmail),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "scopes.#", "3"),
				),
			},
		},
	})
}

func TestAccUserInviteResource_fullConfig(t *testing.T) {
	rEmail := fmt.Sprintf("tf-acc-full-%s@example.com", acctest.RandString(8))
	rWorkspaceName := acctest.RandomWithPrefix("tf-acc-full-ws")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserInviteResourceConfigFull(rEmail, rWorkspaceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_user_invite.test", "email", rEmail),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "role", "member"),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "workspaces.#", "1"),
					resource.TestCheckResourceAttr("portkey_user_invite.test", "scopes.#", "4"),
				),
			},
		},
	})
}

func testAccUserInviteResourceConfig(email, role string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_user_invite" "test" {
  email = %[1]q
  role  = %[2]q
}
`, email, role)
}

func testAccUserInviteResourceConfigWithWorkspace(email, workspaceName, workspaceRole string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[2]q
  description = "Test workspace for invite"
}

resource "portkey_user_invite" "test" {
  email = %[1]q
  role  = "member"

  workspaces = [
    {
      id   = portkey_workspace.test.id
      role = %[3]q
    }
  ]
}
`, email, workspaceName, workspaceRole)
}

func testAccUserInviteResourceConfigWithScopes(email string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_user_invite" "test" {
  email = %[1]q
  role  = "member"

  scopes = [
    "logs.list",
    "logs.view",
    "configs.read"
  ]
}
`, email)
}

func testAccUserInviteResourceConfigFull(email, workspaceName string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[2]q
  description = "Full config test workspace"
}

resource "portkey_user_invite" "test" {
  email = %[1]q
  role  = "member"

  workspaces = [
    {
      id   = portkey_workspace.test.id
      role = "member"
    }
  ]

  scopes = [
    "logs.list",
    "logs.view",
    "configs.read",
    "configs.list"
  ]
}
`, email, workspaceName)
}
