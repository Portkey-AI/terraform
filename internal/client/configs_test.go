package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// TestCreateConfigRequest_HasNoIsDefaultField is a regression test for the
// "produced an unexpected new value: .is_default" bug on portkey_config.
//
// The Portkey /configs API silently drops is_default (and its historical
// misspelling isDefault) on both POST and PUT — verified against
// api.portkey.ai and against the backend source at
// albus/src/api/v2/configs/controllers/{create,update}.js. Sending it
// only produced the illusion that setting is_default worked, then a hard
// "provider produced inconsistent result" error at apply time because the
// API's response (is_default: 0) contradicted the plan value (true).
//
// The permanent fix is to remove any is_default-shaped field from
// CreateConfigRequest so no caller — however well-meaning — can ever
// serialise it into the outgoing body again. This test walks the struct
// with reflection so it stays valid if the request grows other fields,
// and fails loudly if the write path is reintroduced under either name.
func TestCreateConfigRequest_HasNoIsDefaultField(t *testing.T) {
	forbidden := map[string]bool{"is_default": true, "isDefault": true}
	rt := reflect.TypeOf(CreateConfigRequest{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag := f.Tag.Get("json")
		name := strings.SplitN(tag, ",", 2)[0]
		if forbidden[name] {
			t.Errorf("CreateConfigRequest.%s has forbidden JSON tag %q: the Portkey API silently drops it and having the field in the struct is what caused 'produced an unexpected new value: .is_default'", f.Name, name)
		}
	}
}

// TestCreateConfig_URL confirms the client posts to /configs on the
// configured BaseURL — kept alongside the is_default regression above so
// a refactor that moves the endpoint doesn't silently break either.
func TestCreateConfig_URL(t *testing.T) {
	var capturedPath, capturedMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"cfg-1","version_id":"ver-1","slug":"pc-test","object":"config"}`))
	}))
	t.Cleanup(srv.Close)

	c := newTestClient(t, srv.URL)
	if _, err := c.CreateConfig(context.Background(), CreateConfigRequest{
		Name:   "test",
		Config: map[string]any{"retry": map[string]any{"attempts": 3}},
	}); err != nil {
		t.Fatalf("CreateConfig returned error: %v", err)
	}

	if capturedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", capturedMethod)
	}
	if !strings.HasSuffix(capturedPath, "/configs") {
		t.Errorf("expected path to end with /configs, got %s", capturedPath)
	}
}
