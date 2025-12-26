package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"portkey": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates the necessary test environment variables exist.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PORTKEY_API_KEY"); v == "" {
		t.Fatal("PORTKEY_API_KEY must be set for acceptance tests")
	}
}

// providerConfig is a shared configuration for all acceptance tests.
const providerConfig = `
provider "portkey" {}
`

// Ensure terraform.State is imported for use in import state functions
var _ = terraform.State{}

// TestProvider_HasChildResources verifies the provider has resources
func TestProvider_HasChildResources(t *testing.T) {
	expectedResources := []string{
		"portkey_workspace",
		"portkey_workspace_member",
		"portkey_user_invite",
		"portkey_integration",
		"portkey_api_key",
		"portkey_provider",
		"portkey_config",
		"portkey_prompt",
		"portkey_guardrail",
		"portkey_usage_limits_policy",
		"portkey_rate_limits_policy",
	}

	resources := New("test")().Resources(context.Background())

	if len(resources) != len(expectedResources) {
		t.Errorf("Expected %d resources, got %d", len(expectedResources), len(resources))
	}
}

// TestProvider_HasChildDataSources verifies the provider has data sources
func TestProvider_HasChildDataSources(t *testing.T) {
	expectedDataSources := []string{
		"portkey_workspace",
		"portkey_workspaces",
		"portkey_user",
		"portkey_users",
		"portkey_integration",
		"portkey_integrations",
		"portkey_api_key",
		"portkey_api_keys",
		"portkey_provider",
		"portkey_providers",
		"portkey_config",
		"portkey_configs",
		"portkey_prompt",
		"portkey_prompts",
		"portkey_guardrail",
		"portkey_guardrails",
		"portkey_usage_limits_policy",
		"portkey_usage_limits_policies",
		"portkey_rate_limits_policy",
		"portkey_rate_limits_policies",
	}

	dataSources := New("test")().DataSources(context.Background())

	if len(dataSources) != len(expectedDataSources) {
		t.Errorf("Expected %d data sources, got %d", len(expectedDataSources), len(dataSources))
	}
}

// TestAccProvider_Configure validates the provider can be configured
func TestAccProvider_Configure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "portkey_workspaces" "test" {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_workspaces.test", "workspaces.#"),
				),
			},
		},
	})
}
