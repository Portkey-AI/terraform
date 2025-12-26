package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-ds")
	collectionID := "a0d6b8c5-dfc4-11f0-84d4-024c88f9cbd3"
	virtualKey := "4d1848ab-3e0b-42be-b147-f08d9087f6ef"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig(rName, collectionID, virtualKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_prompt.test", "id"),
					resource.TestCheckResourceAttr("data.portkey_prompt.test", "name", rName),
					resource.TestCheckResourceAttr("data.portkey_prompt.test", "collection_id", collectionID),
					resource.TestCheckResourceAttr("data.portkey_prompt.test", "model", "gpt-4o"),
					resource.TestCheckResourceAttr("data.portkey_prompt.test", "status", "active"),
					resource.TestCheckResourceAttrSet("data.portkey_prompt.test", "template"),
				),
			},
		},
	})
}

func testAccPromptDataSourceConfig(name, collectionID, virtualKey string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_prompt" "test" {
  name          = %[1]q
  collection_id = %[2]q
  virtual_key   = %[3]q
  template      = "Hello {{name}}"
  model         = "gpt-4o"
  parameters    = "{}"
}

data "portkey_prompt" "test" {
  slug = portkey_prompt.test.slug
}
`, name, collectionID, virtualKey)
}

