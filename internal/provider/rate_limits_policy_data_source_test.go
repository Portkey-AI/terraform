package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRateLimitsPolicyDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-ds")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRateLimitsPolicyDataSourceConfig(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_rate_limits_policy.test", "id"),
					resource.TestCheckResourceAttr("data.portkey_rate_limits_policy.test", "name", rName),
					resource.TestCheckResourceAttr("data.portkey_rate_limits_policy.test", "type", "requests"),
					resource.TestCheckResourceAttr("data.portkey_rate_limits_policy.test", "unit", "rpm"),
					resource.TestCheckResourceAttr("data.portkey_rate_limits_policy.test", "status", "active"),
				),
			},
		},
	})
}

func testAccRateLimitsPolicyDataSourceConfig(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_rate_limits_policy" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  conditions   = jsonencode([
    {
      key   = "workspace_id"
      value = %[2]q
    }
  ])
  group_by = jsonencode([
    {
      key = "api_key"
    }
  ])
  type  = "requests"
  unit  = "rpm"
  value = 100
}

data "portkey_rate_limits_policy" "test" {
  id = portkey_rate_limits_policy.test.id
}
`, name, workspaceID)
}
