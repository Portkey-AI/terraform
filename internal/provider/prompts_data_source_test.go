package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptsDataSource_basic(t *testing.T) {
	collectionID := getTestCollectionID()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if collectionID == "" {
				t.Skip("TEST_COLLECTION_ID must be set for prompts data source tests")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Use collection filter to ensure we get valid results
				Config: testAccPromptsDataSourceConfigWithCollection(collectionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.portkey_prompts.collection", "collection_id"),
				),
			},
		},
	})
}

func TestAccPromptsDataSource_withCollection(t *testing.T) {
	collectionID := getTestCollectionID()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if collectionID == "" {
				t.Skip("TEST_COLLECTION_ID must be set for prompts data source tests")
			}
		},
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

func testAccPromptsDataSourceConfigWithCollection(collectionID string) string {
	return `
provider "portkey" {}

data "portkey_prompts" "collection" {
  collection_id = "` + collectionID + `"
}
`
}
