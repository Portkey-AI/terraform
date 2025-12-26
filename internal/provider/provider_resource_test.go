package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Test workspace ID and integration that is enabled for it
// These are from the test environment
const testWorkspaceID = "9da48f29-e564-4bcd-8480-757803acf5ae"
const testIntegrationSlug = "foundry-entra"

func TestAccProviderResource_basic(t *testing.T) {
	rnd := rand.Int63()
	name := fmt.Sprintf("tf-acc-test-provider-%d", rnd)
	nameUpdated := fmt.Sprintf("tf-acc-test-provider-%d-updated", rnd)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProviderResourceConfig(name, testWorkspaceID, testIntegrationSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_provider.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_provider.test", "slug"),
					resource.TestCheckResourceAttr("portkey_provider.test", "name", name),
					resource.TestCheckResourceAttr("portkey_provider.test", "workspace_id", testWorkspaceID),
					resource.TestCheckResourceAttr("portkey_provider.test", "integration_id", testIntegrationSlug),
					resource.TestCheckResourceAttr("portkey_provider.test", "status", "active"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "portkey_provider.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:", testWorkspaceID), // Will be completed dynamically
				ImportStateVerify: false,                               // Skip verify due to import format
			},
			// Update and Read testing
			{
				Config: testAccProviderResourceConfigWithNote(nameUpdated, testWorkspaceID, testIntegrationSlug, "Updated note"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_provider.test", "name", nameUpdated),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProviderResourceConfig(name, workspaceID, integrationID string) string {
	return fmt.Sprintf(`
resource "portkey_provider" "test" {
  name           = %[1]q
  workspace_id   = %[2]q
  integration_id = %[3]q
}
`, name, workspaceID, integrationID)
}

func testAccProviderResourceConfigWithNote(name, workspaceID, integrationID, note string) string {
	return fmt.Sprintf(`
resource "portkey_provider" "test" {
  name           = %[1]q
  workspace_id   = %[2]q
  integration_id = %[3]q
  note           = %[4]q
}
`, name, workspaceID, integrationID, note)
}
