package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRateLimitsPolicyResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRateLimitsPolicyResourceConfig(rName, workspaceID, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_rate_limits_policy.test", "id"),
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "name", rName),
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "type", "requests"),
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "unit", "rpm"),
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "value", "100"),
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "status", "active"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "portkey_rate_limits_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			// Update value testing
			{
				Config: testAccRateLimitsPolicyResourceConfig(rName, workspaceID, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "value", "200"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRateLimitsPolicyResource_updateName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-rename")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRateLimitsPolicyResourceConfig(rName, workspaceID, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "name", rName),
				),
			},
			{
				Config: testAccRateLimitsPolicyResourceConfig(rName+"-updated", workspaceID, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_rate_limits_policy.test", "name", rName+"-updated"),
				),
			},
		},
	})
}

func testAccRateLimitsPolicyResourceConfig(name, workspaceID string, value int) string {
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
  value = %[3]d
}
`, name, workspaceID, value)
}
