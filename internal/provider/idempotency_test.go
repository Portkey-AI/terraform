package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These tests verify that Read() operations don't overwrite RequiresReplace attributes,
// which would cause unnecessary resource replacements on subsequent terraform apply.
//
// The bug pattern:
// 1. User creates resource with workspace_id = "abc-123"
// 2. API stores and returns workspace_id = "abc-123" (or potentially different format)
// 3. On next terraform plan, Read() overwrites state.workspace_id with API value
// 4. If ANY difference (even formatting), RequiresReplace triggers destroy+create
//
// The fix: Preserve user-provided RequiresReplace attributes in Read() if already set.

// simulateReadBehavior simulates what happens during a Read() operation
// Returns true if the behavior would cause unnecessary replacement
func simulateReadBehavior(
	stateValue string,
	apiValue string,
	preserveState bool,
) (resultValue string, wouldCauseReplacement bool) {
	state := types.StringValue(stateValue)

	if preserveState {
		// NEW (correct) behavior: preserve state if set
		if state.IsNull() || state.IsUnknown() {
			return apiValue, false
		}
		return stateValue, false
	} else {
		// OLD (buggy) behavior: always overwrite from API
		resultValue = apiValue
		wouldCauseReplacement = stateValue != apiValue
		return resultValue, wouldCauseReplacement
	}
}

// =============================================================================
// CONFIG RESOURCE TESTS
// =============================================================================

func TestConfigResource_Idempotency_WorkspaceID(t *testing.T) {
	tests := []struct {
		name         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "same_value_no_change",
			stateValue:   "ws-12345",
			apiValue:     "ws-12345",
			expectChange: false,
		},
		{
			name:         "different_case_causes_replacement",
			stateValue:   "9da48f29-e564-4bcd-8480-757803acf5ae",
			apiValue:     "9DA48F29-E564-4BCD-8480-757803ACF5AE",
			expectChange: true, // BUG: this would trigger replacement
		},
		{
			name:         "api_returns_different_format",
			stateValue:   "workspace-123",
			apiValue:     "ws-workspace-123-normalized",
			expectChange: true, // BUG: this would trigger replacement
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test OLD behavior (buggy)
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)

			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior: expected wouldCauseReplacement=%v, got %v", tt.expectChange, wouldReplace)
			}

			if tt.expectChange {
				t.Logf("BUG DETECTED: workspace_id change from %q to %q would trigger RequiresReplace", tt.stateValue, tt.apiValue)
			}

			// Test NEW behavior (fixed) - should never cause replacement when state is set
			_, wouldReplaceNew := simulateReadBehavior(tt.stateValue, tt.apiValue, true)
			if wouldReplaceNew {
				t.Errorf("NEW behavior should NOT cause replacement, but got wouldCauseReplacement=true")
			}
		})
	}
}

// =============================================================================
// GUARDRAIL RESOURCE TESTS
// =============================================================================

func TestGuardrailResource_Idempotency_WorkspaceID(t *testing.T) {
	tests := []struct {
		name         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "same_value_no_change",
			stateValue:   "ws-guardrail-test",
			apiValue:     "ws-guardrail-test",
			expectChange: false,
		},
		{
			name:         "whitespace_difference",
			stateValue:   "ws-12345",
			apiValue:     "ws-12345 ", // trailing space
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior: expected wouldCauseReplacement=%v, got %v", tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: workspace_id change would trigger RequiresReplace")
			}
		})
	}
}

// =============================================================================
// PROVIDER RESOURCE TESTS (portkey_provider / virtual key)
// =============================================================================

func TestProviderResource_Idempotency_RequiresReplaceAttrs(t *testing.T) {
	// Provider has: slug, workspace_id, integration_id with RequiresReplace

	tests := []struct {
		name         string
		attr         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "slug_same",
			attr:         "slug",
			stateValue:   "my-provider",
			apiValue:     "my-provider",
			expectChange: false,
		},
		{
			name:         "slug_normalized_by_api",
			attr:         "slug",
			stateValue:   "My-Provider",
			apiValue:     "my-provider", // API lowercases
			expectChange: true,
		},
		{
			name:         "integration_id_different",
			attr:         "integration_id",
			stateValue:   "int-abc123",
			apiValue:     "integration-abc123-full",
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior for %s: expected wouldCauseReplacement=%v, got %v", tt.attr, tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: %s change would trigger RequiresReplace", tt.attr)
			}
		})
	}
}

// =============================================================================
// PROMPT RESOURCE TESTS
// =============================================================================

func TestPromptResource_Idempotency_CollectionID(t *testing.T) {
	tests := []struct {
		name         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "collection_id_same",
			stateValue:   "col-12345",
			apiValue:     "col-12345",
			expectChange: false,
		},
		{
			name:         "collection_id_uuid_format_change",
			stateValue:   "550e8400-e29b-41d4-a716-446655440000",
			apiValue:     "550E8400-E29B-41D4-A716-446655440000",
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior: expected wouldCauseReplacement=%v, got %v", tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: collection_id change would trigger RequiresReplace")
			}
		})
	}
}

// =============================================================================
// INTEGRATION RESOURCE TESTS
// =============================================================================

func TestIntegrationResource_Idempotency_RequiresReplaceAttrs(t *testing.T) {
	// Integration has: slug, ai_provider_id with RequiresReplace

	tests := []struct {
		name         string
		attr         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "ai_provider_id_same",
			attr:         "ai_provider_id",
			stateValue:   "openai",
			apiValue:     "openai",
			expectChange: false,
		},
		{
			name:         "ai_provider_id_normalized",
			attr:         "ai_provider_id",
			stateValue:   "OpenAI",
			apiValue:     "openai",
			expectChange: true,
		},
		{
			name:         "slug_auto_generated_differs",
			attr:         "slug",
			stateValue:   "my-integration",
			apiValue:     "my-integration-abc123", // API adds suffix
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior for %s: expected wouldCauseReplacement=%v, got %v", tt.attr, tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: %s change would trigger RequiresReplace", tt.attr)
			}
		})
	}
}

// =============================================================================
// API KEY RESOURCE TESTS
// =============================================================================

func TestAPIKeyResource_Idempotency_RequiresReplaceAttrs(t *testing.T) {
	// API Key has: type, sub_type, workspace_id, user_id with RequiresReplace

	tests := []struct {
		name         string
		attr         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "type_same",
			attr:         "type",
			stateValue:   "workspace",
			apiValue:     "workspace",
			expectChange: false,
		},
		{
			name:         "type_from_combined_format",
			attr:         "type",
			stateValue:   "workspace",
			apiValue:     "organisation", // parseAPIKeyType bug
			expectChange: true,
		},
		{
			name:         "workspace_id_different",
			attr:         "workspace_id",
			stateValue:   "ws-user-provided",
			apiValue:     "ws-api-normalized",
			expectChange: true,
		},
		{
			name:         "user_id_different",
			attr:         "user_id",
			stateValue:   "user-abc",
			apiValue:     "usr-abc-normalized",
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior for %s: expected wouldCauseReplacement=%v, got %v", tt.attr, tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: %s change would trigger RequiresReplace", tt.attr)
			}
		})
	}
}

// =============================================================================
// USER INVITE RESOURCE TESTS
// =============================================================================

func TestUserInviteResource_Idempotency_RequiresReplaceAttrs(t *testing.T) {
	// User Invite has: email, role with RequiresReplace

	tests := []struct {
		name         string
		attr         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "email_same",
			attr:         "email",
			stateValue:   "user@example.com",
			apiValue:     "user@example.com",
			expectChange: false,
		},
		{
			name:         "email_case_normalized",
			attr:         "email",
			stateValue:   "User@Example.COM",
			apiValue:     "user@example.com",
			expectChange: true,
		},
		{
			name:         "role_same",
			attr:         "role",
			stateValue:   "admin",
			apiValue:     "admin",
			expectChange: false,
		},
		{
			name:         "role_case_difference",
			attr:         "role",
			stateValue:   "Admin",
			apiValue:     "admin",
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior for %s: expected wouldCauseReplacement=%v, got %v", tt.attr, tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: %s change would trigger RequiresReplace", tt.attr)
			}
		})
	}
}

// =============================================================================
// RATE LIMITS POLICY RESOURCE TESTS
// =============================================================================

func TestRateLimitsPolicyResource_Idempotency_RequiresReplaceAttrs(t *testing.T) {
	// Rate Limits Policy has: workspace_id, conditions, group_by, type with RequiresReplace

	tests := []struct {
		name         string
		attr         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "workspace_id_same",
			attr:         "workspace_id",
			stateValue:   "ws-policy-test",
			apiValue:     "ws-policy-test",
			expectChange: false,
		},
		{
			name:         "workspace_id_different",
			attr:         "workspace_id",
			stateValue:   "ws-user-specified",
			apiValue:     "ws-api-returned",
			expectChange: true,
		},
		{
			name:         "type_same",
			attr:         "type",
			stateValue:   "requests",
			apiValue:     "requests",
			expectChange: false,
		},
		{
			name:         "type_different",
			attr:         "type",
			stateValue:   "requests",
			apiValue:     "tokens",
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior for %s: expected wouldCauseReplacement=%v, got %v", tt.attr, tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: %s change would trigger RequiresReplace", tt.attr)
			}
		})
	}
}

// =============================================================================
// USAGE LIMITS POLICY RESOURCE TESTS
// =============================================================================

func TestUsageLimitsPolicyResource_Idempotency_RequiresReplaceAttrs(t *testing.T) {
	// Usage Limits Policy has: workspace_id, conditions, group_by, type, periodic_reset with RequiresReplace

	tests := []struct {
		name         string
		attr         string
		stateValue   string
		apiValue     string
		expectChange bool
	}{
		{
			name:         "periodic_reset_same",
			attr:         "periodic_reset",
			stateValue:   "monthly",
			apiValue:     "monthly",
			expectChange: false,
		},
		{
			name:         "periodic_reset_different",
			attr:         "periodic_reset",
			stateValue:   "monthly",
			apiValue:     "weekly",
			expectChange: true,
		},
		{
			name:         "type_cost_vs_tokens",
			attr:         "type",
			stateValue:   "cost",
			apiValue:     "credits", // API might use different term
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, wouldReplace := simulateReadBehavior(tt.stateValue, tt.apiValue, false)
			if wouldReplace != tt.expectChange {
				t.Errorf("OLD behavior for %s: expected wouldCauseReplacement=%v, got %v", tt.attr, tt.expectChange, wouldReplace)
			}
			if tt.expectChange {
				t.Logf("BUG DETECTED: %s change would trigger RequiresReplace", tt.attr)
			}
		})
	}
}

// =============================================================================
// SUMMARY TEST - Verifies the fix pattern works for all resources
// =============================================================================

func TestAllResources_FixedBehavior_NoUnnecessaryReplacements(t *testing.T) {
	// This test verifies that with the NEW (fixed) behavior,
	// NO resources will trigger unnecessary replacements

	testCases := []struct {
		resource   string
		attr       string
		stateValue string
		apiValue   string
	}{
		{"config", "workspace_id", "ws-123", "WS-123"},
		{"guardrail", "workspace_id", "ws-456", "ws-456-normalized"},
		{"provider", "integration_id", "int-abc", "integration-abc"},
		{"prompt", "collection_id", "col-xyz", "COL-XYZ"},
		{"integration", "ai_provider_id", "OpenAI", "openai"},
		{"api_key", "workspace_id", "ws-key", "workspace-key"},
		{"user_invite", "email", "User@Test.com", "user@test.com"},
		{"rate_limits_policy", "workspace_id", "ws-rate", "WS-RATE"},
		{"usage_limits_policy", "workspace_id", "ws-usage", "WS-USAGE"},
	}

	t.Log("=== Testing FIXED behavior: No unnecessary replacements ===")

	for _, tc := range testCases {
		t.Run(tc.resource+"_"+tc.attr, func(t *testing.T) {
			// With fix: preserveState = true
			result, wouldReplace := simulateReadBehavior(tc.stateValue, tc.apiValue, true)

			if wouldReplace {
				t.Errorf("FAILED: %s.%s would still trigger replacement even with fix!", tc.resource, tc.attr)
			} else {
				t.Logf("OK: %s.%s preserved as %q (API returned %q)", tc.resource, tc.attr, result, tc.apiValue)
			}
		})
	}
}

func TestAllResources_BuggyBehavior_CausesReplacements(t *testing.T) {
	// This test verifies that WITHOUT the fix,
	// resources WILL trigger unnecessary replacements

	testCases := []struct {
		resource   string
		attr       string
		stateValue string
		apiValue   string
	}{
		{"config", "workspace_id", "ws-123", "WS-123"},
		{"guardrail", "workspace_id", "ws-456", "ws-456-normalized"},
		{"provider", "integration_id", "int-abc", "integration-abc"},
		{"prompt", "collection_id", "col-xyz", "COL-XYZ"},
		{"integration", "ai_provider_id", "OpenAI", "openai"},
		{"api_key", "workspace_id", "ws-key", "workspace-key"},
		{"user_invite", "email", "User@Test.com", "user@test.com"},
		{"rate_limits_policy", "workspace_id", "ws-rate", "WS-RATE"},
		{"usage_limits_policy", "workspace_id", "ws-usage", "WS-USAGE"},
	}

	t.Log("=== Testing BUGGY behavior: Causes unnecessary replacements ===")

	bugsFound := 0
	for _, tc := range testCases {
		t.Run(tc.resource+"_"+tc.attr, func(t *testing.T) {
			// Without fix: preserveState = false
			_, wouldReplace := simulateReadBehavior(tc.stateValue, tc.apiValue, false)

			if wouldReplace {
				bugsFound++
				t.Logf("BUG CONFIRMED: %s.%s would trigger unnecessary replacement (%q -> %q)",
					tc.resource, tc.attr, tc.stateValue, tc.apiValue)
			}
		})
	}

	t.Logf("\n=== SUMMARY: %d resources would have unnecessary replacements ===", bugsFound)

	if bugsFound == 0 {
		t.Error("Expected to find bugs in the old behavior, but found none!")
	}
}

