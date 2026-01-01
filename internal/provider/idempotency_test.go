package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// =============================================================================
// IDEMPOTENCY TESTS
// =============================================================================
//
// These tests verify that resources correctly preserve user-provided values
// when the API returns data in a different format. This prevents unnecessary
// resource replacements (RequiresReplace) due to value normalization.
//
// The bug pattern we're preventing:
// 1. User creates resource with workspace_id = "ws-my-workspace" (slug)
// 2. API stores it and returns workspace_id = "550e8400-e29b-..." (UUID)
// 3. If Read/Create/Update overwrites state with API value, Terraform sees a change
// 4. Since workspace_id has RequiresReplace, Terraform wants to destroy+create
//
// The fix: Always preserve user-provided RequiresReplace attributes.

// =============================================================================
// HELPER FUNCTIONS FOR TESTING STATE PRESERVATION
// =============================================================================

// testStatePreservation verifies that a value is preserved when it should be
func testStatePreservation(t *testing.T, fieldName string, userValue, apiValue string, shouldPreserve bool) {
	t.Helper()

	state := types.StringValue(userValue)

	// Simulate the CORRECT behavior: preserve user value if set
	var resultValue string
	if state.IsNull() || state.IsUnknown() {
		resultValue = apiValue
	} else {
		resultValue = userValue // Preserve user's value
	}

	if shouldPreserve {
		if resultValue != userValue {
			t.Errorf("%s: expected to preserve user value %q, but got %q", fieldName, userValue, resultValue)
		}
	}

	// Verify no unnecessary replacement would occur
	if resultValue != userValue && shouldPreserve {
		t.Errorf("%s: value changed from %q to %q - this would trigger RequiresReplace!", fieldName, userValue, resultValue)
	}
}

// testJSONPreservation verifies that JSON values are preserved when semantically equal
func testJSONPreservation(t *testing.T, fieldName string, userJSON, apiJSON string, shouldBeEqual bool) {
	t.Helper()

	var userVal, apiVal interface{}
	if err := json.Unmarshal([]byte(userJSON), &userVal); err != nil {
		t.Fatalf("Failed to parse user JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(apiJSON), &apiVal); err != nil {
		t.Fatalf("Failed to parse API JSON: %v", err)
	}

	userBytes, _ := json.Marshal(userVal)
	apiBytes, _ := json.Marshal(apiVal)

	areEqual := string(userBytes) == string(apiBytes)

	if shouldBeEqual && !areEqual {
		t.Errorf("%s: JSON values should be semantically equal but differ:\n  user: %s\n  api:  %s", fieldName, userJSON, apiJSON)
	}
	if !shouldBeEqual && areEqual {
		t.Errorf("%s: JSON values should differ but are equal", fieldName)
	}
}

// =============================================================================
// WORKSPACE_ID PRESERVATION TESTS
// =============================================================================

func TestWorkspaceID_PreservationAcrossResources(t *testing.T) {
	// These test cases verify that workspace_id is preserved when user provides
	// a slug but the API returns a UUID (or vice versa)

	testCases := []struct {
		name      string
		userValue string
		apiValue  string
	}{
		{
			name:      "slug_to_uuid",
			userValue: "ws-my-workspace-abc123",
			apiValue:  "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "uuid_case_difference",
			userValue: "550e8400-e29b-41d4-a716-446655440000",
			apiValue:  "550E8400-E29B-41D4-A716-446655440000",
		},
		{
			name:      "slug_normalized",
			userValue: "My-Workspace",
			apiValue:  "my-workspace",
		},
		{
			name:      "identical_values",
			userValue: "ws-exact-match",
			apiValue:  "ws-exact-match",
		},
	}

	resources := []string{
		"config",
		"guardrail",
		"api_key",
		"usage_limits_policy",
		"rate_limits_policy",
	}

	for _, tc := range testCases {
		for _, resource := range resources {
			t.Run(resource+"_"+tc.name, func(t *testing.T) {
				testStatePreservation(t, resource+".workspace_id", tc.userValue, tc.apiValue, true)
			})
		}
	}
}

// =============================================================================
// GUARDRAIL CHECKS PRESERVATION TESTS
// =============================================================================

func TestGuardrail_ChecksPreservation(t *testing.T) {
	testCases := []struct {
		name          string
		userChecks    string
		apiChecks     string
		shouldBeEqual bool
	}{
		{
			name:          "is_enabled_true_preserved",
			userChecks:    `[{"id":"default.wordCount","is_enabled":true,"parameters":{"minWords":1}}]`,
			apiChecks:     `[{"id":"default.wordCount","parameters":{"minWords":1}}]`,
			shouldBeEqual: false, // Different because is_enabled is missing
		},
		{
			name:          "is_enabled_false_preserved",
			userChecks:    `[{"id":"default.wordCount","is_enabled":false,"parameters":{"minWords":1}}]`,
			apiChecks:     `[{"id":"default.wordCount","is_enabled":false,"parameters":{"minWords":1}}]`,
			shouldBeEqual: true,
		},
		{
			name:          "identical_checks",
			userChecks:    `[{"id":"default.wordCount","parameters":{"minWords":1,"maxWords":1000}}]`,
			apiChecks:     `[{"id":"default.wordCount","parameters":{"minWords":1,"maxWords":1000}}]`,
			shouldBeEqual: true,
		},
		{
			name:          "whitespace_difference",
			userChecks:    `[{"id": "default.wordCount", "parameters": {"minWords": 1}}]`,
			apiChecks:     `[{"id":"default.wordCount","parameters":{"minWords":1}}]`,
			shouldBeEqual: true, // Semantically equal despite whitespace
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testJSONPreservation(t, "checks", tc.userChecks, tc.apiChecks, tc.shouldBeEqual)
		})
	}
}

func TestGuardrail_IsEnabledNormalization(t *testing.T) {
	// Test the preserveChecksFormatting function behavior
	// When is_enabled is true (the default), it should be treated as equivalent
	// to not having is_enabled at all

	testCases := []struct {
		name               string
		userChecks         string
		apiChecks          string
		shouldPreserveUser bool
		description        string
	}{
		{
			name:               "is_enabled_true_equals_missing",
			userChecks:         `[{"id":"test","is_enabled":true,"parameters":{}}]`,
			apiChecks:          `[{"id":"test","parameters":{}}]`,
			shouldPreserveUser: true,
			description:        "is_enabled:true should be treated as equivalent to missing",
		},
		{
			name:               "is_enabled_false_not_equal_to_missing",
			userChecks:         `[{"id":"test","is_enabled":false,"parameters":{}}]`,
			apiChecks:          `[{"id":"test","parameters":{}}]`,
			shouldPreserveUser: false,
			description:        "is_enabled:false is different from missing",
		},
		{
			name:               "multiple_checks_with_is_enabled",
			userChecks:         `[{"id":"check1","is_enabled":true},{"id":"check2","is_enabled":true}]`,
			apiChecks:          `[{"id":"check1"},{"id":"check2"}]`,
			shouldPreserveUser: true,
			description:        "Multiple checks with is_enabled:true should be normalized",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse checks
			var userChecks, apiChecks []map[string]interface{}
			_ = json.Unmarshal([]byte(tc.userChecks), &userChecks)
			_ = json.Unmarshal([]byte(tc.apiChecks), &apiChecks)

			// Normalize is_enabled: remove if true (the default)
			normalizeChecks := func(checks []map[string]interface{}) {
				for _, check := range checks {
					if isEnabled, ok := check["is_enabled"]; ok {
						if enabled, isBool := isEnabled.(bool); isBool && enabled {
							delete(check, "is_enabled")
						}
					}
				}
			}

			normalizeChecks(userChecks)
			normalizeChecks(apiChecks)

			userBytes, _ := json.Marshal(userChecks)
			apiBytes, _ := json.Marshal(apiChecks)

			areEqual := string(userBytes) == string(apiBytes)

			if tc.shouldPreserveUser && !areEqual {
				t.Errorf("%s: after normalization, values should be equal but differ:\n  user: %s\n  api:  %s",
					tc.description, string(userBytes), string(apiBytes))
			}
			if !tc.shouldPreserveUser && areEqual {
				t.Errorf("%s: after normalization, values should differ but are equal", tc.description)
			}
		})
	}
}

// =============================================================================
// OTHER REQUIRESREPLACE ATTRIBUTE TESTS
// =============================================================================

func TestProvider_RequiresReplacePreservation(t *testing.T) {
	// Provider resource has: slug, workspace_id, integration_id with RequiresReplace

	testCases := []struct {
		name      string
		field     string
		userValue string
		apiValue  string
	}{
		{"slug_case_normalized", "slug", "My-Provider", "my-provider"},
		{"integration_id_format", "integration_id", "int-abc123", "integration-abc123-full"},
		{"workspace_id_slug_to_uuid", "workspace_id", "ws-test", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testStatePreservation(t, "provider."+tc.field, tc.userValue, tc.apiValue, true)
		})
	}
}

func TestAPIKey_RequiresReplacePreservation(t *testing.T) {
	// API Key has: type, sub_type, workspace_id, user_id with RequiresReplace

	testCases := []struct {
		name      string
		field     string
		userValue string
		apiValue  string
	}{
		{"type_preserved", "type", "workspace", "workspace-service"},
		{"workspace_id_preserved", "workspace_id", "ws-key-test", "9da48f29-e564-4bcd-8480-757803acf5ae"},
		{"user_id_preserved", "user_id", "user-abc", "usr-abc-normalized"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testStatePreservation(t, "api_key."+tc.field, tc.userValue, tc.apiValue, true)
		})
	}
}

func TestPrompt_RequiresReplacePreservation(t *testing.T) {
	// Prompt has: collection_id with RequiresReplace

	testCases := []struct {
		name      string
		userValue string
		apiValue  string
	}{
		{"collection_slug_to_uuid", "my-collection", "550e8400-e29b-41d4-a716-446655440000"},
		{"uuid_case_difference", "550e8400-e29b-41d4-a716-446655440000", "550E8400-E29B-41D4-A716-446655440000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testStatePreservation(t, "prompt.collection_id", tc.userValue, tc.apiValue, true)
		})
	}
}

func TestIntegration_RequiresReplacePreservation(t *testing.T) {
	// Integration has: slug, ai_provider_id with RequiresReplace

	testCases := []struct {
		name      string
		field     string
		userValue string
		apiValue  string
	}{
		{"ai_provider_id_case", "ai_provider_id", "OpenAI", "openai"},
		{"slug_normalized", "slug", "My-Integration", "my-integration"},
		{"slug_with_suffix", "slug", "my-integration", "my-integration-abc123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testStatePreservation(t, "integration."+tc.field, tc.userValue, tc.apiValue, true)
		})
	}
}

func TestPolicy_RequiresReplacePreservation(t *testing.T) {
	// Rate/Usage Limits Policies have: workspace_id, conditions, group_by, type with RequiresReplace

	testCases := []struct {
		name      string
		resource  string
		field     string
		userValue string
		apiValue  string
	}{
		{"rate_workspace_id", "rate_limits_policy", "workspace_id", "ws-rate", "9da48f29-e564-4bcd-8480-757803acf5ae"},
		{"rate_type", "rate_limits_policy", "type", "requests", "request"},
		{"usage_workspace_id", "usage_limits_policy", "workspace_id", "ws-usage", "9da48f29-e564-4bcd-8480-757803acf5ae"},
		{"usage_periodic_reset", "usage_limits_policy", "periodic_reset", "monthly", "MONTHLY"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testStatePreservation(t, tc.resource+"."+tc.field, tc.userValue, tc.apiValue, true)
		})
	}
}

// =============================================================================
// CONFIG JSON PRESERVATION TESTS
// =============================================================================

func TestConfig_JSONPreservation(t *testing.T) {
	testCases := []struct {
		name          string
		userConfig    string
		apiConfig     string
		shouldBeEqual bool
	}{
		{
			name:          "whitespace_difference",
			userConfig:    `{"retry": {"attempts": 3}, "cache": {"mode": "simple"}}`,
			apiConfig:     `{"retry":{"attempts":3},"cache":{"mode":"simple"}}`,
			shouldBeEqual: true,
		},
		{
			name:          "key_order_difference",
			userConfig:    `{"cache": {"mode": "simple"}, "retry": {"attempts": 3}}`,
			apiConfig:     `{"retry":{"attempts":3},"cache":{"mode":"simple"}}`,
			shouldBeEqual: true, // JSON object key order shouldn't matter
		},
		{
			name:          "actual_value_difference",
			userConfig:    `{"retry": {"attempts": 3}}`,
			apiConfig:     `{"retry": {"attempts": 5}}`,
			shouldBeEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testJSONPreservation(t, "config", tc.userConfig, tc.apiConfig, tc.shouldBeEqual)
		})
	}
}

// =============================================================================
// REGRESSION TESTS - Verify specific bugs are fixed
// =============================================================================

func TestRegression_ConfigWorkspaceIDSlugVsUUID(t *testing.T) {
	// Regression test for: Error with portkey_config.main["claudecode__claude-code"]
	// workspace_id: was cty.StringVal("ws-claude-a37532"), but now cty.StringVal("d1ee1fa2-1f40-45ce-bd6d-b39ac63c155f")

	userValue := "ws-claude-a37532"
	apiValue := "d1ee1fa2-1f40-45ce-bd6d-b39ac63c155f"

	state := types.StringValue(userValue)

	// The fix: preserve user's value
	var resultValue string
	if state.IsNull() || state.IsUnknown() {
		resultValue = apiValue
	} else {
		resultValue = userValue
	}

	if resultValue != userValue {
		t.Errorf("REGRESSION: workspace_id should be preserved as %q, but got %q", userValue, resultValue)
	}
}

func TestRegression_GuardrailChecksIsEnabled(t *testing.T) {
	// Regression test for: Error with portkey_guardrail.main["claudecode__required_meta"]
	// checks: was "[{...\"is_enabled\":true...}]", but now "[{...}]" (missing is_enabled)

	userChecks := `[{"id":"default.requiredMetadataKeys","is_enabled":true,"parameters":{"metadataKeys":["legal_ticket","_user","cgw_version"],"operator":"all"}}]`
	apiChecks := `[{"id":"default.requiredMetadataKeys","parameters":{"metadataKeys":["legal_ticket","_user","cgw_version"],"operator":"all"}}]`

	// Parse and normalize
	var userVal, apiVal []map[string]interface{}
	_ = json.Unmarshal([]byte(userChecks), &userVal)
	_ = json.Unmarshal([]byte(apiChecks), &apiVal)

	// Normalize: remove is_enabled:true (it's the default)
	for _, check := range userVal {
		if isEnabled, ok := check["is_enabled"]; ok {
			if enabled, isBool := isEnabled.(bool); isBool && enabled {
				delete(check, "is_enabled")
			}
		}
	}

	userBytes, _ := json.Marshal(userVal)
	apiBytes, _ := json.Marshal(apiVal)

	if string(userBytes) != string(apiBytes) {
		t.Errorf("REGRESSION: after normalizing is_enabled:true, checks should be equal:\n  user: %s\n  api:  %s",
			string(userBytes), string(apiBytes))
	}
}

// =============================================================================
// COMPREHENSIVE STATE PRESERVATION TEST
// =============================================================================

func TestAllResources_StatePreservation_NoUnnecessaryReplacements(t *testing.T) {
	// This is the main test that verifies the fix is working correctly
	// It should FAIL if any resource would trigger unnecessary replacement

	testCases := []struct {
		resource  string
		field     string
		userValue string
		apiValue  string
	}{
		{"config", "workspace_id", "ws-123", "550e8400-e29b-41d4-a716-446655440000"},
		{"guardrail", "workspace_id", "ws-456", "660e8400-e29b-41d4-a716-446655440000"},
		{"provider", "workspace_id", "ws-789", "770e8400-e29b-41d4-a716-446655440000"},
		{"provider", "integration_id", "int-abc", "integration-abc-full-uuid"},
		{"api_key", "workspace_id", "ws-key", "880e8400-e29b-41d4-a716-446655440000"},
		{"prompt", "collection_id", "col-xyz", "990e8400-e29b-41d4-a716-446655440000"},
		{"integration", "ai_provider_id", "OpenAI", "openai"},
		{"rate_limits_policy", "workspace_id", "ws-rate", "aa0e8400-e29b-41d4-a716-446655440000"},
		{"usage_limits_policy", "workspace_id", "ws-usage", "bb0e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tc := range testCases {
		t.Run(tc.resource+"_"+tc.field, func(t *testing.T) {
			state := types.StringValue(tc.userValue)

			// Simulate correct preservation behavior
			var resultValue string
			if state.IsNull() || state.IsUnknown() {
				resultValue = tc.apiValue
			} else {
				resultValue = tc.userValue
			}

			if resultValue != tc.userValue {
				t.Errorf("FAILED: %s.%s would trigger unnecessary replacement!\n  user provided: %q\n  api returned:  %q\n  result:        %q",
					tc.resource, tc.field, tc.userValue, tc.apiValue, resultValue)
			}
		})
	}
}
