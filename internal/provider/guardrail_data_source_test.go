package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGuardrailDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-ds")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailDataSourceConfig(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_guardrail.test", "id"),
					resource.TestCheckResourceAttr("data.portkey_guardrail.test", "name", rName),
					resource.TestCheckResourceAttr("data.portkey_guardrail.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("data.portkey_guardrail.test", "status", "active"),
					resource.TestCheckResourceAttrSet("data.portkey_guardrail.test", "checks"),
					resource.TestCheckResourceAttrSet("data.portkey_guardrail.test", "actions"),
				),
			},
		},
	})
}

func testAccGuardrailDataSourceConfig(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_guardrail" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  checks       = jsonencode([
    {
      id = "default.wordCount"
      parameters = {
        minWords = 1
        maxWords = 1000
      }
    }
  ])
  actions = jsonencode({
    onFail  = "log"
    message = "Word count check failed"
  })
}

data "portkey_guardrail" "test" {
  slug = portkey_guardrail.test.slug
}
`, name, workspaceID)
}

