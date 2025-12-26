package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUsageLimitsPolicyDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-ds")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitsPolicyDataSourceConfig(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_usage_limits_policy.test", "id"),
					resource.TestCheckResourceAttr("data.portkey_usage_limits_policy.test", "name", rName),
					resource.TestCheckResourceAttr("data.portkey_usage_limits_policy.test", "type", "cost"),
					resource.TestCheckResourceAttr("data.portkey_usage_limits_policy.test", "status", "active"),
				),
			},
		},
	})
}

func testAccUsageLimitsPolicyDataSourceConfig(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_usage_limits_policy" "test" {
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
  type           = "cost"
  credit_limit   = 1000.0
  periodic_reset = "monthly"
}

data "portkey_usage_limits_policy" "test" {
  id = portkey_usage_limits_policy.test.id
}
`, name, workspaceID)
}
