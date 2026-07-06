package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccWorkspaceDefaultsResource_lifecycle exercises the full DAG in a
// single config: workspace → guardrails → workspace_defaults. Because
// workspace_defaults references guardrail.id (and guardrails reference
// workspace.id), Terraform's DAG resolves cleanly with no cycle.
func TestAccWorkspaceDefaultsResource_lifecycle(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-wsdef")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create workspace, guardrails, and defaults (attached).
			{
				Config: testAccWorkspaceDefaultsConfigAttached(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_workspace_defaults.test", "id"),
					resource.TestCheckResourceAttr("portkey_workspace_defaults.test", "input_guardrails.#", "1"),
					resource.TestCheckResourceAttr("portkey_workspace_defaults.test", "output_guardrails.#", "1"),
				),
			},
			// Step 2: no-op re-apply must show no drift.
			{
				Config:   testAccWorkspaceDefaultsConfigAttached(rName),
				PlanOnly: true,
			},
			// Step 3: clear both with []; workspace still stands.
			{
				Config: testAccWorkspaceDefaultsConfigCleared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace_defaults.test", "input_guardrails.#", "0"),
					resource.TestCheckResourceAttr("portkey_workspace_defaults.test", "output_guardrails.#", "0"),
				),
			},
			// Step 4: import — verify state hydrates from the workspace by ID.
			{
				ResourceName:      "portkey_workspace_defaults.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspaceDefaultsConfigAttached(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = "workspace_defaults lifecycle test"
}

resource "portkey_guardrail" "input" {
  name         = "%[1]s-in"
  workspace_id = portkey_workspace.test.id
  checks = jsonencode([{
    id = "default.wordCount"
    parameters = {
      minWords = 1
      maxWords = 1000
    }
  }])
  actions = jsonencode({
    onFail  = "log"
    message = "input check"
  })
}

resource "portkey_guardrail" "output" {
  name         = "%[1]s-out"
  workspace_id = portkey_workspace.test.id
  checks = jsonencode([{
    id = "default.wordCount"
    parameters = {
      minWords = 1
      maxWords = 1000
    }
  }])
  actions = jsonencode({
    onFail  = "log"
    message = "output check"
  })
}

resource "portkey_workspace_defaults" "test" {
  workspace_id      = portkey_workspace.test.id
  input_guardrails  = [portkey_guardrail.input.slug]
  output_guardrails = [portkey_guardrail.output.slug]
}
`, name)
}

func testAccWorkspaceDefaultsConfigCleared(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = "workspace_defaults lifecycle test"
}

resource "portkey_guardrail" "input" {
  name         = "%[1]s-in"
  workspace_id = portkey_workspace.test.id
  checks = jsonencode([{
    id = "default.wordCount"
    parameters = {
      minWords = 1
      maxWords = 1000
    }
  }])
  actions = jsonencode({
    onFail  = "log"
    message = "input check"
  })
}

resource "portkey_guardrail" "output" {
  name         = "%[1]s-out"
  workspace_id = portkey_workspace.test.id
  checks = jsonencode([{
    id = "default.wordCount"
    parameters = {
      minWords = 1
      maxWords = 1000
    }
  }])
  actions = jsonencode({
    onFail  = "log"
    message = "output check"
  })
}

resource "portkey_workspace_defaults" "test" {
  workspace_id      = portkey_workspace.test.id
  input_guardrails  = []
  output_guardrails = []
}
`, name)
}
