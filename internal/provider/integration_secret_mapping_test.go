package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccIntegrationResource_secretMappingValueFormat exercises the
// value_format attribute end-to-end through Terraform. It complements the
// unit tests in secret_mapping_test.go, which cover only the Go conversion
// helpers, by driving a real API round-trip so the framework's implicit
// post-apply plan-idempotency check catches any server-side default leaking
// into state ("Provider produced inconsistent result after apply").
//
// secret_reference_id uses .slug (not .id) because the Portkey API accepts a
// UUID on write but echoes back the slug, which would leave state
// uncorrelated with the plan. That is a pre-existing gateway behaviour,
// unrelated to value_format.
func TestAccIntegrationResource_secretMappingValueFormat(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-vf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// No value_format at all -> must stay null, and the framework's
			// implicit post-apply plan check catches any server-side default
			// leaking into state.
			{
				Config: testAccIntegrationSecretMappingValueFormat(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_integration.test", "secret_mappings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("portkey_integration.test", "secret_mappings.*", map[string]string{
						"target_field": "key",
						"secret_key":   "openai_key",
					}),
					resource.TestCheckNoResourceAttr("portkey_integration.test", "secret_mappings.0.value_format"),
				),
			},
			// value_format = "json" -> accepted and echoed back.
			{
				Config: testAccIntegrationSecretMappingValueFormat(rName, "json"),
				Check: resource.TestCheckTypeSetElemNestedAttrs("portkey_integration.test", "secret_mappings.*", map[string]string{
					"target_field": "key",
					"value_format": "json",
				}),
			},
			// Explicit "string" -> also round-trips.
			{
				Config: testAccIntegrationSecretMappingValueFormat(rName, "string"),
				Check: resource.TestCheckTypeSetElemNestedAttrs("portkey_integration.test", "secret_mappings.*", map[string]string{
					"target_field": "key",
					"value_format": "string",
				}),
			},
		},
	})
}

func testAccIntegrationSecretMappingValueFormat(name, valueFormat string) string {
	vf := ""
	if valueFormat != "" {
		vf = fmt.Sprintf("\n      value_format        = %q", valueFormat)
	}
	return fmt.Sprintf(`
%s

resource "portkey_secret_reference" "test" {
  name         = "%s-sr"
  manager_type = "hashicorp_vault"
  secret_path  = "kv/data/test/path"

  vault_approle_auth = {
    vault_addr      = "https://vault.example.internal"
    vault_role_id   = "test-role-id"
    vault_secret_id = "test-secret-id"
  }

  allow_all_workspaces = true
}

resource "portkey_integration" "test" {
  name           = %q
  ai_provider_id = "openai"

  secret_mappings = [
    {
      target_field        = "key"
      secret_reference_id = portkey_secret_reference.test.slug
      secret_key          = "openai_key"%s
    },
  ]
}
`, providerConfig, name, name, vf)
}
