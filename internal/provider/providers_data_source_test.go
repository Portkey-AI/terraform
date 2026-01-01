package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProvidersDataSource_basic(t *testing.T) {
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all providers for a workspace
			{
				Config: testAccProvidersDataSourceConfig(workspaceID),
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
