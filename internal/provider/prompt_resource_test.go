package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	// Using workspace 9da48f29-e564-4bcd-8480-757803acf5ae for testing
	collectionID := "a0d6b8c5-dfc4-11f0-84d4-024c88f9cbd3"
	virtualKey := "4d1848ab-3e0b-42be-b147-f08d9087f6ef"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPromptResourceConfig(rName, collectionID, virtualKey, "Hello {{name}}", "gpt-4o"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_prompt.test", "slug"),
					resource.TestCheckResourceAttr("portkey_prompt.test", "name", rName),
					resource.TestCheckResourceAttr("portkey_prompt.test", "collection_id", collectionID),
					resource.TestCheckResourceAttr("portkey_prompt.test", "template", "Hello {{name}}"),
					resource.TestCheckResourceAttr("portkey_prompt.test", "model", "gpt-4o"),
					resource.TestCheckResourceAttr("portkey_prompt.test", "status", "active"),
					resource.TestCheckResourceAttr("portkey_prompt.test", "prompt_version", "1"),
					resource.TestCheckResourceAttrSet("portkey_prompt.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "portkey_prompt.test",
				ImportState:       true,
				ImportStateVerify: true,
				// parameters: API adds "model" field; virtual_key: API returns slug instead of ID
				ImportStateVerifyIgnore: []string{"created_at", "updated_at", "version_description", "parameters", "virtual_key"},
			},
			// Update name testing
			{
				Config: testAccPromptResourceConfig(rName+"-renamed", collectionID, virtualKey, "Hello {{name}}", "gpt-4o"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_prompt.test", "name", rName+"-renamed"),
					// Version should not change for name-only updates
					resource.TestCheckResourceAttr("portkey_prompt.test", "prompt_version", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPromptResource_updateName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-rename")
	collectionID := "a0d6b8c5-dfc4-11f0-84d4-024c88f9cbd3"
	virtualKey := "4d1848ab-3e0b-42be-b147-f08d9087f6ef"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig(rName, collectionID, virtualKey, "Test prompt", "gpt-4o"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_prompt.test", "name", rName),
				),
			},
			{
				Config: testAccPromptResourceConfig(rName+"-updated", collectionID, virtualKey, "Test prompt", "gpt-4o"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_prompt.test", "name", rName+"-updated"),
				),
			},
		},
	})
}

func testAccPromptResourceConfig(name, collectionID, virtualKey, template, model string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_prompt" "test" {
  name          = %[1]q
  collection_id = %[2]q
  virtual_key   = %[3]q
  template      = %[4]q
  model         = %[5]q
  parameters    = "{}"
}
`, name, collectionID, virtualKey, template, model)
}
