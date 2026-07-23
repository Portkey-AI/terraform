package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/portkey-ai/terraform-provider-portkey/internal/client"
)

func strPtr(s string) *string { return &s }

// TestSecretMappings_RoundTripWithValueFormat verifies that value_format
// survives a full state -> wire -> state round-trip, which is the whole point
// of exposing the field in the provider (Pylon #6704).
func TestSecretMappings_RoundTripWithValueFormat(t *testing.T) {
	ctx := context.Background()

	apiMappings := []client.SecretMapping{
		{
			TargetField:       "configurations.vertex_service_account_json",
			SecretReferenceID: "sr_vault_sa",
			SecretKey:         strPtr("vertex_service_account_json"),
			ValueFormat:       strPtr("json"),
		},
	}

	objType := types.ObjectType{AttrTypes: secretMappingAttrTypes}
	priorNull := types.SetNull(objType)

	setVal, diags := secretMappingsFromClient(apiMappings, priorNull)
	if diags.HasError() {
		t.Fatalf("secretMappingsFromClient returned errors: %v", diags)
	}

	roundTripped, diags := secretMappingsToClient(ctx, setVal)
	if diags.HasError() {
		t.Fatalf("secretMappingsToClient returned errors: %v", diags)
	}
	if roundTripped == nil {
		t.Fatal("secretMappingsToClient returned nil, expected slice")
	}
	if len(*roundTripped) != 1 {
		t.Fatalf("expected 1 mapping after round-trip, got %d", len(*roundTripped))
	}

	got := (*roundTripped)[0]
	if got.TargetField != "configurations.vertex_service_account_json" {
		t.Errorf("target_field lost in round-trip: got %q", got.TargetField)
	}
	if got.SecretReferenceID != "sr_vault_sa" {
		t.Errorf("secret_reference_id lost in round-trip: got %q", got.SecretReferenceID)
	}
	if got.SecretKey == nil || *got.SecretKey != "vertex_service_account_json" {
		t.Errorf("secret_key lost in round-trip: got %v", got.SecretKey)
	}
	if got.ValueFormat == nil {
		t.Fatal("value_format lost in round-trip: got nil pointer")
	}
	if *got.ValueFormat != "json" {
		t.Errorf("value_format lost in round-trip: got %q, want %q", *got.ValueFormat, "json")
	}
}

// TestSecretMapping_JSONMarshal_OmitEmpty verifies that the wire payload
// omits value_format when it is unset, so pre-existing callers keep the same
// on-the-wire request shape (albus defaults missing value_format to "string").
func TestSecretMapping_JSONMarshal_OmitEmpty(t *testing.T) {
	m := client.SecretMapping{
		TargetField:       "key",
		SecretReferenceID: "sr_abc",
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(b)
	want := `{"target_field":"key","secret_reference_id":"sr_abc"}`
	if got != want {
		t.Errorf("wire payload changed for unset value_format:\n got  %s\n want %s", got, want)
	}

	m.ValueFormat = strPtr("json")
	b, err = json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got = string(b)
	want = `{"target_field":"key","secret_reference_id":"sr_abc","value_format":"json"}`
	if got != want {
		t.Errorf("wire payload wrong when value_format=json:\n got  %s\n want %s", got, want)
	}
}

// TestSecretMappings_NullValueFormatRoundTrip guards against spurious drift
// on integrations that don't set value_format: null in state must stay null
// after a Read of an API response that omits the field.
func TestSecretMappings_NullValueFormatRoundTrip(t *testing.T) {
	apiMappings := []client.SecretMapping{
		{
			TargetField:       "key",
			SecretReferenceID: "sr_openai",
		},
	}

	objType := types.ObjectType{AttrTypes: secretMappingAttrTypes}
	priorNull := types.SetNull(objType)

	setVal, diags := secretMappingsFromClient(apiMappings, priorNull)
	if diags.HasError() {
		t.Fatalf("secretMappingsFromClient: %v", diags)
	}

	elems := setVal.Elements()
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	obj, ok := elems[0].(types.Object)
	if !ok {
		t.Fatalf("expected types.Object element, got %T", elems[0])
	}
	vf := obj.Attributes()["value_format"].(types.String)
	if !vf.IsNull() {
		t.Errorf("value_format should be null when API omits it, got %v", vf)
	}
}

// Compile-time check to ensure secretMappingAttrTypes remains
// authoritative for the shape used by state conversion helpers.
var _ = map[string]attr.Type(secretMappingAttrTypes)
