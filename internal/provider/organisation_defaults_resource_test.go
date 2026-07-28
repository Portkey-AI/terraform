package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccOrganisationDefaultsResource_lifecycle exercises organisation-scoped
// guardrails → organisation_defaults.
//
// NOTE: portkey_organisation_defaults is a singleton over shared organisation
// state — there is no per-test isolation the way workspace defaults get from a
// throwaway workspace. This test therefore writes to the test organisation's
// real defaults, and the final destroy clears both lists. Do not run it against
// an organisation whose defaults matter, and avoid running it in parallel with
// anything else that touches organisation defaults.
func TestAccOrganisationDefaultsResource_lifecycle(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-orgdef")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create org-scoped guardrails and attach both lists.
			{
				Config: testAccOrganisationDefaultsConfigAttached(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_organisation_defaults.test", "id", "organisation_defaults"),
					resource.TestCheckResourceAttr("portkey_organisation_defaults.test", "input_guardrails.#", "1"),
					resource.TestCheckResourceAttr("portkey_organisation_defaults.test", "output_guardrails.#", "1"),
					// Guardrails created without workspace_id must stay org-scoped.
					resource.TestCheckNoResourceAttr("portkey_guardrail.org_input", "workspace_id"),
				),
			},
			// Step 2: no-op re-apply must show no drift. Confirms the API's
			// {id, slug} read shape round-trips to the slugs used in HCL.
			{
				Config:   testAccOrganisationDefaultsConfigAttached(rName),
				PlanOnly: true,
			},
			// Step 3: import while both lists are populated. The import ID is
			// ignored (the organisation comes from the API key), so state must
			// hydrate purely from the defaults endpoint.
			{
				ResourceName:      "portkey_organisation_defaults.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 4: clear both lists with [].
			{
				Config: testAccOrganisationDefaultsConfigCleared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_organisation_defaults.test", "input_guardrails.#", "0"),
					resource.TestCheckResourceAttr("portkey_organisation_defaults.test", "output_guardrails.#", "0"),
				),
			},
		},
	})
}

// TestAccOrganisationDefaultsResource_requiresOneList verifies the plan-time
// guard mirroring the API's "at least one of input_guardrails or
// output_guardrails" requirement.
func TestAccOrganisationDefaultsResource_requiresOneList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "portkey" {}

resource "portkey_organisation_defaults" "test" {}
`,
				// Loose regex: the framework's exact wording for AtLeastOneOf has
				// shifted across releases, so only assert it names both attributes.
				ExpectError: regexp.MustCompile(`(?is)input_guardrails.*output_guardrails`),
			},
		},
	})
}

// TestAccOrganisationDefaultsResource_rejectsWorkspaceGuardrail verifies the
// API-side scope check surfaces as a Terraform error when a workspace-scoped
// guardrail is used as an organisation default.
func TestAccOrganisationDefaultsResource_rejectsWorkspaceGuardrail(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-orgdef-ws")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrganisationDefaultsConfigWorkspaceScoped(rName),
				ExpectError: regexp.MustCompile(`(?is)workspace-scoped guardrails`),
			},
		},
	})
}

// testAccOrganisationDefaultsGuardrails declares two organisation-scoped
// guardrails. Omitting workspace_id is what makes them org-scoped, which the
// defaults endpoint requires.
func testAccOrganisationDefaultsGuardrails(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_guardrail" "org_input" {
  name = "%[1]s-in"
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

resource "portkey_guardrail" "org_output" {
  name = "%[1]s-out"
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

func testAccOrganisationDefaultsConfigAttached(name string) string {
	return testAccOrganisationDefaultsGuardrails(name) + `
resource "portkey_organisation_defaults" "test" {
  input_guardrails  = [portkey_guardrail.org_input.slug]
  output_guardrails = [portkey_guardrail.org_output.slug]
}
`
}

func testAccOrganisationDefaultsConfigCleared(name string) string {
	return testAccOrganisationDefaultsGuardrails(name) + `
resource "portkey_organisation_defaults" "test" {
  input_guardrails  = []
  output_guardrails = []
}
`
}

func testAccOrganisationDefaultsConfigWorkspaceScoped(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_workspace" "test" {
  name        = %[1]q
  description = "organisation_defaults scope-rejection test"
}

resource "portkey_guardrail" "workspace_scoped" {
  name         = "%[1]s-ws"
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
    message = "workspace check"
  })
}

resource "portkey_organisation_defaults" "test" {
  input_guardrails = [portkey_guardrail.workspace_scoped.slug]
}
`, name)
}
