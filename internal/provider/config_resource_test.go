package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("portkey_config.test", "id"),
					resource.TestCheckResourceAttrSet("portkey_config.test", "slug"),
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
					resource.TestCheckResourceAttr("portkey_config.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("portkey_config.test", "status", "active"),
					resource.TestCheckResourceAttrSet("portkey_config.test", "version_id"),
					resource.TestCheckResourceAttrSet("portkey_config.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "portkey_config.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			// Update testing - change config values
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":5},"cache":{"mode":"simple"}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccConfigResource_updateName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-rename")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
				),
			},
			{
				Config: testAccConfigResourceConfig(rName+"-renamed", workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName+"-renamed"),
				),
			},
		},
	})
}

// TestAccConfigResource_workspaceSlug tests that workspace_id is preserved when
// user provides a slug but API returns UUID (regression test for the bug)
func TestAccConfigResource_workspaceSlug(t *testing.T) {
	workspaceSlug := getTestWorkspaceSlug()
	if workspaceSlug == "" {
		t.Skip("TEST_WORKSPACE_SLUG must be set to test slug preservation")
	}

	rName := acctest.RandomWithPrefix("tf-acc-slug")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigResourceConfig(rName, workspaceSlug, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
					// Verify workspace_id is preserved as the slug, not converted to UUID
					resource.TestCheckResourceAttr("portkey_config.test", "workspace_id", workspaceSlug),
				),
			},
			// Apply again without changes - should be idempotent
			{
				Config: testAccConfigResourceConfig(rName, workspaceSlug, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Workspace ID should still be the slug
					resource.TestCheckResourceAttr("portkey_config.test", "workspace_id", workspaceSlug),
				),
			},
		},
	})
}

// TestAccConfigResource_jsonWhitespace tests that config JSON with different
// whitespace doesn't trigger unnecessary updates
func TestAccConfigResource_jsonWhitespace(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-json")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with compact JSON
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
				),
			},
			// Apply with formatted JSON - should not cause update
			{
				Config: testAccConfigResourceConfigFormatted(rName, workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "name", rName),
				),
			},
		},
	})
}

// TestAccConfigResource_isDefaultReadOnly is a regression test for the
// "produced an unexpected new value: .is_default" apply-time bug. The
// Portkey /configs endpoints silently drop is_default on both POST and
// PUT under either camelCase (isDefault, the historical typo in this
// client) or snake_case (is_default). Because the create response then
// reports is_default: 0 while the plan carried true, the framework
// rejected the apply.
//
// The attribute is now Computed-only: any HCL that tries to write it is
// caught at plan time with the framework's canonical
// "Can't configure a value for" message instead of blowing up mid-apply.
func TestAccConfigResource_isDefaultReadOnly(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-isdefault")
	workspaceID := getTestWorkspaceID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Writing is_default = true is rejected at plan time.
			{
				Config:      testAccConfigResourceConfigWithIsDefault(rName, workspaceID, `{"retry":{"attempts":3}}`, true),
				ExpectError: regexp.MustCompile(`(?s)is_default.*read-only`),
			},
			// Omitting the attribute succeeds and state reflects the API's
			// value (false, the DB default: configs.is_default is never
			// written by any HTTP endpoint).
			{
				Config: testAccConfigResourceConfig(rName, workspaceID, `{"retry":{"attempts":3}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("portkey_config.test", "is_default", "false"),
				),
			},
		},
	})
}

func testAccConfigResourceConfigWithIsDefault(name, workspaceID, config string, isDefault bool) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_config" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  config       = %[3]q
  is_default   = %[4]t
}
`, name, workspaceID, config, isDefault)
}

func testAccConfigResourceConfig(name, workspaceID, config string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_config" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  config       = %[3]q
}
`, name, workspaceID, config)
}

func testAccConfigResourceConfigFormatted(name, workspaceID string) string {
	return fmt.Sprintf(`
provider "portkey" {}

resource "portkey_config" "test" {
  name         = %[1]q
  workspace_id = %[2]q
  config       = jsonencode({
    retry = {
      attempts = 3
    }
  })
}
`, name, workspaceID)
}
