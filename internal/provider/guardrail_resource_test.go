package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGuardrailResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	workspaceID := getTestWorkspaceID()

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
	workspaceID := getTestWorkspaceID()

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

// TestAccGuardrailResource_withIsEnabled tests that is_enabled:true in checks
// is preserved and doesn't cause inconsistency (regression test for the bug)
func TestAccGuardrailResource_withIsEnabled(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-enabled")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with is_enabled: true
			{
				Config: testAccGuardrailResourceConfigWithIsEnabled(rName, workspaceID, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "checks"),
				),
			},
			// Apply again - should be idempotent (no changes)
			{
				Config: testAccGuardrailResourceConfigWithIsEnabled(rName, workspaceID, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
				),
			},
		},
	})
}

// TestAccGuardrailResource_workspaceSlug tests that workspace_id is preserved
// when user provides a slug but API returns UUID
func TestAccGuardrailResource_workspaceSlug(t *testing.T) {
	workspaceSlug := getTestWorkspaceSlug()
	if workspaceSlug == "" {
		t.Skip("TEST_WORKSPACE_SLUG must be set to test slug preservation")
	}

	rName := acctest.RandomWithPrefix("tf-acc-slug")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourceConfig(rName, workspaceSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
					// Verify workspace_id is preserved as the slug
					resource.TestCheckResourceAttr("portkey_guardrail.test", "workspace_id", workspaceSlug),
				),
			},
			// Apply again - should preserve the slug
			{
				Config: testAccGuardrailResourceConfig(rName, workspaceSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "workspace_id", workspaceSlug),
				),
			},
		},
	})
}

// TestAccGuardrailResource_multipleChecks tests guardrails with multiple checks
func TestAccGuardrailResource_multipleChecks(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-multi")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourceConfigMultipleChecks(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "checks"),
				),
			},
			// Apply again - should be idempotent
			{
				Config: testAccGuardrailResourceConfigMultipleChecks(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
				),
			},
		},
	})
}

// testAccPreCheckOrganisationGuardrails gates tests that create
// organisation-scoped guardrails.
//
// Portkey scopes guardrail access per verb, and organisation-level guardrails
// use their own `organisation_guardrails.*` family rather than the
// `guardrails.*` one every workspace-scoped test in this file relies on. An org
// admin key that passes all of those can still lack `organisation_guardrails`
// scopes entirely and get `403 AB03 You do not have enough permissions to
// execute this request`. Without this opt-in the weekly acceptance-test workflow
// would go red on a key permission gap rather than a provider defect.
func testAccPreCheckOrganisationGuardrails(t *testing.T) {
	if os.Getenv("PORTKEY_TEST_ORG_GUARDRAILS") == "" {
		t.Skip("PORTKEY_TEST_ORG_GUARDRAILS must be set to run tests that create organisation-scoped guardrails; the key needs organisation_guardrails.create, .read, .update and .delete")
	}
	testAccPreCheck(t)
}

// TestAccGuardrailResource_organisationScoped tests the full CRUD lifecycle of
// an organisation-scoped guardrail, created by omitting workspace_id. The API
// derives the organisation from the Admin API key, and workspace_id must stay
// null in state (rather than "") so a config that omits it shows no drift.
func TestAccGuardrailResource_organisationScoped(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-orgguard")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckOrganisationGuardrails(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccGuardrailResourceConfigOrgScoped(rName, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_guardrail.test", "slug"),
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
					resource.TestCheckResourceAttr("portkey_guardrail.test", "status", "active"),
					resource.TestCheckNoResourceAttr("portkey_guardrail.test", "workspace_id"),
				),
			},
			// Re-apply must be a no-op: workspace_id must not drift to "".
			{
				Config:   testAccGuardrailResourceConfigOrgScoped(rName, 1000),
				PlanOnly: true,
			},
			// Import
			{
				ResourceName:            "portkey_guardrail.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			// Update in place (checks change), still org-scoped. Needs
			// `organisation_guardrails.update`.
			{
				Config: testAccGuardrailResourceConfigOrgScoped(rName, 2000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_guardrail.test", "name", rName),
					resource.TestCheckNoResourceAttr("portkey_guardrail.test", "workspace_id"),
					// Confirm the PUT actually landed rather than just returning 200.
					resource.TestMatchResourceAttr("portkey_guardrail.test", "checks", regexp.MustCompile(`"maxWords":\s*2000`)),
				),
			},
			// Delete happens automatically in TestCase.
		},
	})
}

func testAccGuardrailResourceConfigOrgScoped(name string, maxWords int) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_guardrail" "test" {
  name   = %[1]q
  checks = jsonencode([
    {
      id = "default.wordCount"
      parameters = {
        minWords = 1
        maxWords = %[2]d
      }
    }
  ])
  actions = jsonencode({
    onFail  = "log"
    message = "Word count check failed"
  })
}
`, name, maxWords)
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

func testAccGuardrailResourceConfigWithIsEnabled(name, workspaceID string, isEnabled bool) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_guardrail" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  checks       = jsonencode([
    {
      id         = "default.wordCount"
      is_enabled = %[3]t
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
`, name, workspaceID, isEnabled)
}

func testAccGuardrailResourceConfigMultipleChecks(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_guardrail" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  checks       = jsonencode([
    {
      id         = "default.wordCount"
      is_enabled = true
      parameters = {
        minWords = 1
        maxWords = 1000
      }
    },
    {
      id         = "default.sentenceCount"
      is_enabled = true
      parameters = {
        minSentences = 1
        maxSentences = 50
      }
    }
  ])
  actions = jsonencode({
    onFail  = "log"
    message = "Guardrail check failed"
  })
}
`, name, workspaceID)
}
