package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGuardrailResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGuardrailResourceConfig(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "slug"),
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
					resource.TestCheckResourceAttr("portkey_guardrail.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("portkey_guardrail.test", "status", "active"),
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "version_id"),
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "portkey_guardrail.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			// Update name testing
			{
				Config: testAccGuardrailResourceConfigUpdated(rName+"-renamed", workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName+"-renamed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccGuardrailResource_updateName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-rename")
	workspaceID := "9da48f29-e564-4bcd-8480-757803acf5ae"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourceConfig(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
				),
			},
			{
				Config: testAccGuardrailResourceConfigUpdated(rName+"-updated", workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName+"-updated"),
				),
			},
		},
	})
}

func testAccGuardrailResourceConfig(name, workspaceID string) string {
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
`, name, workspaceID)
}

func testAccGuardrailResourceConfigUpdated(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_guardrail" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  checks       = jsonencode([
    {
      id = "default.wordCount"
      parameters = {
        minWords = 5
        maxWords = 2000
      }
    }
  ])
  actions = jsonencode({
    onFail  = "block"
    message = "Word count check failed - updated"
  })
}
`, name, workspaceID)
}

