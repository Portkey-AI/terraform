package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGuardrailsDataSource_basic(t *testing.T) {
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailsDataSourceConfig(workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_guardrails.all", "workspace_id"),
				),
			},
		},
	})
}

func testAccGuardrailsDataSourceConfig(workspaceID string) string {
	return `
provider "portkey" {}

data "portkey_guardrails" "all" {
  workspace_id = "` + workspaceID + `"
}
`
}
