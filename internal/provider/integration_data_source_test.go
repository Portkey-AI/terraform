package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIntegrationDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.portkey_integration.test", "id",
						"portkey_integration.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.portkey_integration.test", "name",
						"portkey_integration.test", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.portkey_integration.test", "ai_provider_id",
						"portkey_integration.test", "ai_provider_id",
					),
					resource.TestCheckResourceAttr("data.portkey_integration.test", "status", "active"),
				),
			},
		},
	})
}

func testAccIntegrationDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_integration" "test" {
  name           = %[1]q
  ai_provider_id = "openai"
  description    = "Test integration for data source"
  key            = "sk-test-fake-key-12345"
}

data "portkey_integration" "test" {
  slug = portkey_integration.test.slug
}
`, name)
}
