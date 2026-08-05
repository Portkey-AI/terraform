package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/portkey-ai/terraform-provider-portkey/internal/client"
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

// TestAccWorkspaceDefaultsResource_adoptsOmittedList mirrors
// TestAccOrganisationDefaultsResource_adoptsOmittedList for workspace scope:
// manage only input_guardrails while the workspace already has an output
// guardrail attached out-of-band.
//
// Regression test. While the guardrail lists were Optional without Computed,
// Create wrote the preserved API value into state against a planned null and
// the apply failed with "Provider produced inconsistent result after apply".
func TestAccWorkspaceDefaultsResource_adoptsOmittedList(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-wsdef-adopt")

	var workspaceID, outputSlug string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: workspace and guardrails only, with nothing attached to
			// the workspace defaults yet.
			{
				Config: testAccWorkspaceDefaultsBase(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureResourceAttr("portkey_workspace.test", "id", &workspaceID),
					captureResourceAttr("portkey_guardrail.output", "slug", &outputSlug),
				),
			},
			// Step 2: attach the output guardrail outside Terraform, then let
			// Terraform create workspace_defaults managing input only.
			{
				PreConfig: func() {
					testAccAttachWorkspaceOutputGuardrail(t, workspaceID, outputSlug)
				},
				Config: testAccWorkspaceDefaultsConfigInputOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_workspace_defaults.test", "input_guardrails.#", "1"),
					resource.TestCheckResourceAttrPair(
						"portkey_workspace_defaults.test", "input_guardrails.0",
						"portkey_guardrail.input", "slug",
					),
					resource.TestCheckResourceAttr("portkey_workspace_defaults.test", "output_guardrails.#", "1"),
					resource.TestCheckResourceAttrPair(
						"portkey_workspace_defaults.test", "output_guardrails.0",
						"portkey_guardrail.output", "slug",
					),
				),
			},
			// Step 3: the adopted list must be stable — an omitted attribute
			// must not plan to clear what it preserved.
			{
				Config:   testAccWorkspaceDefaultsConfigInputOnly(rName),
				PlanOnly: true,
			},
		},
	})
}

// testAccAttachWorkspaceOutputGuardrail attaches an existing guardrail to the
// workspace's output defaults without going through Terraform, simulating a
// guardrail attached through the Portkey UI.
func testAccAttachWorkspaceOutputGuardrail(t *testing.T, workspaceID, guardrailSlug string) {
	t.Helper()

	if workspaceID == "" || guardrailSlug == "" {
		t.Fatalf("workspace ID (%q) and guardrail slug (%q) must both be captured from the previous step", workspaceID, guardrailSlug)
	}

	c, err := newTestClient()
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}

	encoded, err := json.Marshal([]string{guardrailSlug})
	if err != nil {
		t.Fatalf("marshal guardrail slug: %v", err)
	}
	if _, err := c.UpdateWorkspace(context.Background(), workspaceID, client.UpdateWorkspaceRequest{
		Defaults: &client.UpdateWorkspaceDefaults{OutputGuardrails: encoded},
	}); err != nil {
		t.Fatalf("UpdateWorkspace: %v", err)
	}
}

// testAccWorkspaceDefaultsBase declares the workspace and both guardrails
// without any portkey_workspace_defaults block.
func testAccWorkspaceDefaultsBase(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = "workspace_defaults adoption test"
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
`, name)
}

// testAccWorkspaceDefaultsConfigInputOnly manages input_guardrails only,
// leaving output_guardrails to whatever is attached out-of-band.
//
// depends_on is required for teardown: with output_guardrails absent from the
// config there is no reference to the output guardrail, so Terraform would
// otherwise delete it in parallel with clearing the defaults and hit
// `AB01 Guardrail is being used in workspace defaults`.
func testAccWorkspaceDefaultsConfigInputOnly(name string) string {
	return testAccWorkspaceDefaultsBase(name) + `
resource "portkey_workspace_defaults" "test" {
  workspace_id     = portkey_workspace.test.id
  input_guardrails = [portkey_guardrail.input.slug]

  depends_on = [portkey_guardrail.output]
}
`
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
