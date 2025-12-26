package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUserDataSource_basic tests fetching a single user by ID.
// Uses the first user from the organization's user list.
func TestAccUserDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfigDynamic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.portkey_user.test", "email"),
					resource.TestCheckResourceAttrSet("data.portkey_user.test", "role"),
					resource.TestCheckResourceAttrSet("data.portkey_user.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.portkey_user.test", "updated_at"),
				),
			},
		},
	})
}

// testAccUserDataSourceConfigDynamic fetches a user using the first user ID from the org
func testAccUserDataSourceConfigDynamic() string {
	return `
provider "portkey" {}

# Get existing users from the organization
data "portkey_users" "all" {}

# Fetch the first user's details
data "portkey_user" "test" {
  id = data.portkey_users.all.users[0].id
}
`
}
