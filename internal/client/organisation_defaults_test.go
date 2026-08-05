package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOrganisationDefaultsURL verifies that the organisation defaults endpoint
// swaps the configured BaseURL's /v1 suffix for /v2, since these endpoints live
// under /v2 while the rest of the Admin API used by this provider is on /v1.
func TestOrganisationDefaultsURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "default v1 base url",
			baseURL: "https://api.portkey.ai/v1",
			want:    "https://api.portkey.ai/v2/admin/organisation/defaults",
		},
		{
			name:    "default v1 base url with trailing slash",
			baseURL: "https://api.portkey.ai/v1/",
			want:    "https://api.portkey.ai/v2/admin/organisation/defaults",
		},
		{
			name:    "self-hosted base url with path prefix",
			baseURL: "https://portkey.internal.example.com/api/v1",
			want:    "https://portkey.internal.example.com/api/v2/admin/organisation/defaults",
		},
		{
			name:    "base url without version suffix",
			baseURL: "https://portkey.internal.example.com",
			want:    "https://portkey.internal.example.com/v2/admin/organisation/defaults",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{BaseURL: tc.baseURL}
			if got := c.organisationDefaultsURL(); got != tc.want {
				t.Errorf("organisationDefaultsURL() = %q; want %q", got, tc.want)
			}
		})
	}
}

// TestGetOrganisationDefaults_OK verifies the {id, slug} read shape is parsed.
func TestGetOrganisationDefaults_OK(t *testing.T) {
	var capturedMethod, capturedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{
			"object":"organisation_defaults",
			"input_guardrails":[{"id":"11111111-1111-1111-1111-111111111111","slug":"gr_input"}],
			"output_guardrails":[{"id":"22222222-2222-2222-2222-222222222222","slug":"gr_output"}]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL+"/v1", "test-key")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	defaults, err := c.GetOrganisationDefaults(context.Background())
	if err != nil {
		t.Fatalf("GetOrganisationDefaults failed: %v", err)
	}

	if capturedMethod != http.MethodGet {
		t.Errorf("method = %q; want GET", capturedMethod)
	}
	if capturedPath != "/v2/admin/organisation/defaults" {
		t.Errorf("path = %q; want /v2/admin/organisation/defaults", capturedPath)
	}
	if defaults.Object != "organisation_defaults" {
		t.Errorf("object = %q; want organisation_defaults", defaults.Object)
	}
	if len(defaults.InputGuardrails) != 1 || defaults.InputGuardrails[0].Slug != "gr_input" {
		t.Errorf("unexpected input_guardrails: %+v", defaults.InputGuardrails)
	}
	if len(defaults.OutputGuardrails) != 1 || defaults.OutputGuardrails[0].ID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("unexpected output_guardrails: %+v", defaults.OutputGuardrails)
	}
}

// TestUpdateOrganisationDefaults_OmitsNilFields verifies the tri-state write
// behaviour: a nil field is omitted from the body (preserve), while a marshaled
// `[]` is sent (clear). It also verifies the PUT is followed by a re-fetch,
// since the API returns an empty object.
func TestUpdateOrganisationDefaults_OmitsNilFields(t *testing.T) {
	var methods []string
	var putBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		if r.Method == http.MethodPut {
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &putBody)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		_, _ = w.Write([]byte(`{"object":"organisation_defaults","input_guardrails":[],"output_guardrails":[]}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL+"/v1", "test-key")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if _, err := c.UpdateOrganisationDefaults(context.Background(), UpdateOrganisationDefaultsRequest{
		InputGuardrails: json.RawMessage("[]"),
	}); err != nil {
		t.Fatalf("UpdateOrganisationDefaults failed: %v", err)
	}

	if len(methods) != 2 || methods[0] != http.MethodPut || methods[1] != http.MethodGet {
		t.Errorf("requests = %v; want [PUT GET]", methods)
	}
	if _, ok := putBody["input_guardrails"]; !ok {
		t.Errorf("input_guardrails missing from body: %+v", putBody)
	}
	if _, ok := putBody["output_guardrails"]; ok {
		t.Errorf("output_guardrails should be omitted when nil, body: %+v", putBody)
	}
}
