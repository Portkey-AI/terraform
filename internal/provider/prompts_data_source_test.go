package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_prompts.all", "prompts.#"),
				),
			},
		},
	})
}

func TestAccPromptsDataSource_withCollection(t *testing.T) {
	collectionID := "a0d6b8c5-dfc4-11f0-84d4-024c88f9cbd3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptsDataSourceConfigWithCollection(collectionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Just check that the attribute is set (could be 0 prompts)
					resource.TestCheckResourceAttrSet("data.portkey_prompts.collection", "collection_id"),
				),
			},
		},
	})
}

func testAccPromptsDataSourceConfig() string {
	return `
provider "portkey" {}

data "portkey_prompts" "all" {}
`
}

func testAccPromptsDataSourceConfigWithCollection(collectionID string) string {
	return `
provider "portkey" {}

data "portkey_prompts" "collection" {
  collection_id = "` + collectionID + `"
}
`
}
