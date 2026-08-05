package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/portkey-ai/terraform-provider-portkey/internal/client"
)

func TestMarshalGuardrailsForUpdate(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name    string
		cfg     types.List
		want    string
		wantNil bool
	}{
		{
			name:    "unknown cfg omits field",
			cfg:     types.ListUnknown(types.StringType),
			wantNil: true,
		},
		{
			// An omitted attribute means "not managed by Terraform", so the
			// field is omitted regardless of what prior state held. Returning
			// [] here would destroy guardrails attached out-of-band.
			name:    "null cfg omits field",
			cfg:     types.ListNull(types.StringType),
			wantNil: true,
		},
		{
			name: "empty cfg clears with []",
			cfg:  types.ListValueMust(types.StringType, []attr.Value{}),
			want: "[]",
		},
		{
			name: "populated cfg marshals array",
			cfg: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("g-1"),
				types.StringValue("g-2"),
			}),
			want: `["g-1","g-2"]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := marshalGuardrailsForUpdate(ctx, tc.cfg)
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %s", string(got))
				}
				return
			}
			if string(got) != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, string(got))
			}
		})
	}
}

func TestGuardrailsFromAPIToList(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
		null bool
	}{
		{
			name: "nil input returns null list",
			in:   nil,
			null: true,
		},
		{
			name: "empty input returns null list",
			in:   []string{},
			null: true,
		},
		{
			name: "empty strings are dropped",
			in:   []string{"", ""},
			null: true,
		},
		{
			name: "populated passes through",
			in:   []string{"pg-one", "pg-two"},
			want: []string{"pg-one", "pg-two"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := guardrailsFromAPIToList(tc.in)
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			if tc.null {
				if !got.IsNull() {
					t.Fatalf("expected null list, got %v", got)
				}
				return
			}
			if got.IsNull() {
				t.Fatalf("expected populated list, got null")
			}
			var ids []string
			diags = got.ElementsAs(context.Background(), &ids, false)
			if diags.HasError() {
				t.Fatalf("ElementsAs failed: %v", diags)
			}
			if len(ids) != len(tc.want) {
				t.Fatalf("expected %d ids, got %d (%v)", len(tc.want), len(ids), ids)
			}
			for i := range ids {
				if ids[i] != tc.want[i] {
					t.Fatalf("id[%d] = %q, want %q", i, ids[i], tc.want[i])
				}
			}
		})
	}
}

func TestOrganisationGuardrailRefsToList(t *testing.T) {
	cases := []struct {
		name string
		in   []client.OrganisationGuardrailRef
		want []string
		null bool
	}{
		{
			name: "nil input returns null list",
			in:   nil,
			null: true,
		},
		{
			name: "empty input returns null list",
			in:   []client.OrganisationGuardrailRef{},
			null: true,
		},
		{
			name: "slug is preferred over id",
			in: []client.OrganisationGuardrailRef{
				{ID: "11111111-1111-1111-1111-111111111111", Slug: "gr_one"},
				{ID: "22222222-2222-2222-2222-222222222222", Slug: "gr_two"},
			},
			want: []string{"gr_one", "gr_two"},
		},
		{
			name: "falls back to id when slug is missing",
			in: []client.OrganisationGuardrailRef{
				{ID: "11111111-1111-1111-1111-111111111111"},
				{ID: "22222222-2222-2222-2222-222222222222", Slug: "gr_two"},
			},
			want: []string{"11111111-1111-1111-1111-111111111111", "gr_two"},
		},
		{
			name: "entries with neither id nor slug are dropped",
			in:   []client.OrganisationGuardrailRef{{}, {}},
			null: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := organisationGuardrailRefsToList(tc.in)
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			if tc.null {
				if !got.IsNull() {
					t.Fatalf("expected null list, got %v", got)
				}
				return
			}
			if got.IsNull() {
				t.Fatalf("expected populated list, got null")
			}
			var entries []string
			diags = got.ElementsAs(context.Background(), &entries, false)
			if diags.HasError() {
				t.Fatalf("ElementsAs failed: %v", diags)
			}
			if len(entries) != len(tc.want) {
				t.Fatalf("expected %d entries, got %d (%v)", len(tc.want), len(entries), entries)
			}
			for i := range entries {
				if entries[i] != tc.want[i] {
					t.Fatalf("entry[%d] = %q, want %q", i, entries[i], tc.want[i])
				}
			}
		})
	}
}

func TestReconcileListAfterRead(t *testing.T) {
	nullList := types.ListNull(types.StringType)
	empty := types.ListValueMust(types.StringType, []attr.Value{})
	populated := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("pg-1")})

	tests := []struct {
		name    string
		state   types.List
		apiList types.List
		want    types.List
	}{
		{name: "both null → state (null)", state: nullList, apiList: nullList, want: nullList},
		{name: "state [] + api null → state []", state: empty, apiList: nullList, want: empty},
		{name: "state null + api [] → state null", state: nullList, apiList: empty, want: nullList},
		{name: "state [] + api [] → state []", state: empty, apiList: empty, want: empty},
		{name: "state populated + api populated → api", state: populated, apiList: populated, want: populated},
		{name: "state populated + api null → api null (drift)", state: populated, apiList: nullList, want: nullList},
		{name: "state null + api populated → api (drift)", state: nullList, apiList: populated, want: populated},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := reconcileListAfterRead(tc.state, tc.apiList)
			if !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestCoalesceListForConfig(t *testing.T) {
	nullList := types.ListNull(types.StringType)
	empty := types.ListValueMust(types.StringType, []attr.Value{})
	populated := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("g-1")})
	apiPopulated := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("g-2")})

	tests := []struct {
		name          string
		cfg           types.List
		apiList       types.List
		wantEqual     types.List
		expectEmpty   bool
		expectSameAs  string // "cfg" or "api"
		expectNullOut bool
	}{
		{name: "null cfg + null api → null", cfg: nullList, apiList: nullList, expectNullOut: true},
		{name: "null cfg + populated api → api", cfg: nullList, apiList: apiPopulated, expectSameAs: "api"},
		{name: "empty cfg → empty (user clear intent)", cfg: empty, apiList: nullList, expectEmpty: true},
		{name: "populated cfg + null api → cfg (lag fallback)", cfg: populated, apiList: nullList, expectSameAs: "cfg"},
		{name: "populated cfg + populated api → api", cfg: populated, apiList: apiPopulated, expectSameAs: "api"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coalesceListForConfig(tc.cfg, tc.apiList)
			if tc.expectNullOut {
				if !got.IsNull() {
					t.Fatalf("expected null, got %v", got)
				}
				return
			}
			if tc.expectEmpty {
				if got.IsNull() || len(got.Elements()) != 0 {
					t.Fatalf("expected empty non-null list, got %v", got)
				}
				return
			}
			switch tc.expectSameAs {
			case "cfg":
				if !got.Equal(tc.cfg) {
					t.Fatalf("expected cfg %v, got %v", tc.cfg, got)
				}
			case "api":
				if !got.Equal(tc.apiList) {
					t.Fatalf("expected api %v, got %v", tc.apiList, got)
				}
			}
		})
	}
}
