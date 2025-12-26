package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProvidersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all providers for a workspace
			{
				Config: testAccProvidersDataSourceConfig(testWorkspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_providers.test", "providers.#"),
				),
			},
		},
	})
}

func testAccProvidersDataSourceConfig(workspaceID string) string {
	return fmt.Sprintf(`
data "portkey_providers" "test" {
  workspace_id = %[1]q
}
`, workspaceID)
}

