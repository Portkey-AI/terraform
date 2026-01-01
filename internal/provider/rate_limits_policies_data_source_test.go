package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRateLimitsPoliciesDataSource_basic(t *testing.T) {
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRateLimitsPoliciesDataSourceConfig(workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_rate_limits_policies.all", "workspace_id"),
				),
			},
		},
	})
}

func testAccRateLimitsPoliciesDataSourceConfig(workspaceID string) string {
	return `
provider "portkey" {}

data "portkey_rate_limits_policies" "all" {
  workspace_id = "` + workspaceID + `"
}
`
}
