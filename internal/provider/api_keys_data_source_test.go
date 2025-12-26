package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeysDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all API keys
			{
				Config: testAccAPIKeysDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_api_keys.test", "api_keys.#"),
				),
			},
		},
	})
}

func testAccAPIKeysDataSourceConfig() string {
	return `
data "portkey_api_keys" "test" {}
`
}

