package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/portkey-ai/terraform-provider-portkey/internal/client"
)

// testAccPreCheckOrganisationDefaults gates tests that overwrite the
// organisation's real default guardrails.
//
// portkey_organisation_defaults is a singleton over shared organisation state —
// there is no per-test isolation the way workspace defaults get from a
// throwaway workspace — and destroy clears both lists. Without this opt-in the
// weekly acceptance-test workflow (.github/workflows/acc-tests.yml runs
// `make testacc` every Monday) would silently wipe the default guardrails of
// whatever organisation its API key belongs to. Avoid running these in parallel
// with anything else that touches organisation defaults.
func testAccPreCheckOrganisationDefaults(t *testing.T) {
	if os.Getenv("PORTKEY_TEST_ORG_DEFAULTS") == "" {
		t.Skip("PORTKEY_TEST_ORG_DEFAULTS must be set to run tests that overwrite the organisation's real default guardrails")
	}
	testAccPreCheck(t)
}

// TestAccOrganisationDefaultsResource_lifecycle exercises organisation-scoped
// guardrails → organisation_defaults.
func TestAccOrganisationDefaultsResource_lifecycle(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-orgdef")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckOrganisationDefaults(t) },
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

// TestAccOrganisationDefaultsResource_adoptsOmittedList covers the path
// _lifecycle cannot: managing only one list while the organisation already has
// guardrails of the other kind attached out-of-band (e.g. through the Portkey
// UI). This is the adoption scenario the resource documents as supported.
//
// Regression test. While the guardrail lists were Optional without Computed,
// Create wrote the preserved API value into state against a planned null and
// the apply failed with "Provider produced inconsistent result after apply".
func TestAccOrganisationDefaultsResource_adoptsOmittedList(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-orgdef-adopt")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckOrganisationDefaults(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create the guardrails only, so they exist in the
			// organisation before anything is attached to its defaults.
			{
				Config: testAccOrganisationDefaultsGuardrails(rName),
			},
			// Step 2: attach the output guardrail outside Terraform, then let
			// Terraform create organisation_defaults managing input only. The
			// PUT must omit output_guardrails, and state must adopt what the
			// API preserved rather than reset it to null.
			{
				PreConfig: func() {
					testAccAttachOrganisationOutputGuardrail(t, rName+"-out")
				},
				Config: testAccOrganisationDefaultsConfigInputOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_organisation_defaults.test", "input_guardrails.#", "1"),
					resource.TestCheckResourceAttrPair(
						"portkey_organisation_defaults.test", "input_guardrails.0",
						"portkey_guardrail.org_input", "slug",
					),
					resource.TestCheckResourceAttr("portkey_organisation_defaults.test", "output_guardrails.#", "1"),
					resource.TestCheckResourceAttrPair(
						"portkey_organisation_defaults.test", "output_guardrails.0",
						"portkey_guardrail.org_output", "slug",
					),
				),
			},
			// Step 3: the adopted list must be stable — an omitted attribute
			// must not plan to clear what it preserved.
			{
				Config:   testAccOrganisationDefaultsConfigInputOnly(rName),
				PlanOnly: true,
			},
		},
	})
}

// testAccAttachOrganisationOutputGuardrail attaches an existing guardrail to the
// organisation's output defaults without going through Terraform, simulating a
// guardrail attached through the Portkey UI.
func testAccAttachOrganisationOutputGuardrail(t *testing.T, guardrailName string) {
	t.Helper()

	c, err := newTestClient()
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}

	ctx := context.Background()
	guardrails, err := c.ListGuardrails(ctx, "")
	if err != nil {
		t.Fatalf("ListGuardrails: %v", err)
	}

	var slug string
	for _, g := range guardrails {
		if g.Name == guardrailName {
			slug = g.Slug
			break
		}
	}
	if slug == "" {
		t.Fatalf("guardrail %q not found; cannot simulate an out-of-band attachment", guardrailName)
	}

	encoded, err := json.Marshal([]string{slug})
	if err != nil {
		t.Fatalf("marshal guardrail slug: %v", err)
	}
	if _, err := c.UpdateOrganisationDefaults(ctx, client.UpdateOrganisationDefaultsRequest{
		OutputGuardrails: encoded,
	}); err != nil {
		t.Fatalf("UpdateOrganisationDefaults: %v", err)
	}
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

// testAccOrganisationDefaultsConfigInputOnly manages input_guardrails only,
// leaving output_guardrails to whatever is attached out-of-band.
//
// depends_on is required for teardown: with output_guardrails absent from the
// config there is no reference to org_output, so Terraform would otherwise
// delete that guardrail in parallel with clearing the defaults and hit
// `AB01 Guardrail is being used in organisation defaults`.
func testAccOrganisationDefaultsConfigInputOnly(name string) string {
	return testAccOrganisationDefaultsGuardrails(name) + `
resource "portkey_organisation_defaults" "test" {
  input_guardrails = [portkey_guardrail.org_input.slug]

  depends_on = [portkey_guardrail.org_output]
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
