package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyDataSource_basic(t *testing.T) {
	rnd := rand.Int63()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create an API key first, then read it with data source
			{
				Config: testAccAPIKeyDataSourceConfigRnd(rnd),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("data.portkey_api_key.test", "name"),
					resource.TestCheckResourceAttrSet("data.portkey_api_key.test", "type"),
					resource.TestCheckResourceAttrSet("data.portkey_api_key.test", "sub_type"),
					resource.TestCheckResourceAttr("data.portkey_api_key.test", "status", "active"),
				),
			},
		},
	})
}

func testAccAPIKeyDataSourceConfigRnd(rnd int64) string {
	return fmt.Sprintf(`
resource "portkey_api_key" "test" {
  name     = "tf-acc-test-api-key-ds-%d"
  type     = "organisation"
  sub_type = "service"
  scopes   = ["providers.list"]
}

data "portkey_api_key" "test" {
  id = portkey_api_key.test.id
}
`, rnd)
}

