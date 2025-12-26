# Adding New APIs to the Terraform Provider

Quick playbook for adding new admin API endpoints.

## Process Overview

```
1. Client Methods → 2. Resource/DataSource → 3. Register → 4. Example → 5. Test → 6. Deploy
```

---

## Step 1: Add Client Methods

**File:** `internal/client/client.go`

```go
// Add types
type VirtualKey struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Provider string `json:"provider"`
    // ...
}

type CreateVirtualKeyRequest struct {
    Name     string `json:"name"`
    Provider string `json:"provider"`
    Key      string `json:"key"`
}

// Add CRUD methods
func (c *Client) CreateVirtualKey(ctx context.Context, req CreateVirtualKeyRequest) (*VirtualKey, error) {
    respBody, err := c.doRequest(ctx, http.MethodPost, "/virtual-keys", req)
    // ...
}

func (c *Client) GetVirtualKey(ctx context.Context, id string) (*VirtualKey, error) { ... }
func (c *Client) UpdateVirtualKey(ctx context.Context, id string, req UpdateVirtualKeyRequest) (*VirtualKey, error) { ... }
func (c *Client) DeleteVirtualKey(ctx context.Context, id string) error { ... }
func (c *Client) ListVirtualKeys(ctx context.Context) ([]VirtualKey, error) { ... }
```

---

## Step 2: Create Resource/DataSource

**File:** `internal/provider/virtual_key_resource.go`

Copy an existing resource (e.g., `workspace_resource.go`) and modify:

1. Update struct names and types
2. Define schema attributes
3. Implement CRUD methods: `Create`, `Read`, `Update`, `Delete`
4. Add `ImportState` for import support

**Key pattern:**
```go
func (r *virtualKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan virtualKeyResourceModel
    req.Plan.Get(ctx, &plan)
    
    result, err := r.client.CreateVirtualKey(ctx, client.CreateVirtualKeyRequest{...})
    
    // Map response to state
    plan.ID = types.StringValue(result.ID)
    resp.State.Set(ctx, plan)
}
```

---

## Step 3: Register in Provider

**File:** `internal/provider/provider.go`

```go
func (p *portkeyProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewWorkspaceResource,
        NewVirtualKeyResource,  // Add here
    }
}

func (p *portkeyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewVirtualKeysDataSource,  // Add here if needed
    }
}
```

---

## Step 4: Create Example

**File:** `examples/<feature>/main.tf`

```hcl
resource "portkey_virtual_key" "openai" {
  name     = "OpenAI Production"
  provider = "openai"
  key      = var.openai_api_key
}

output "virtual_key_id" {
  value = portkey_virtual_key.openai.id
}
```

---

## Step 5: Test & Verify

### Build
```bash
make build
make install
```

### Test
```bash
cd examples/<feature>
export PORTKEY_API_KEY="your-key"

terraform init
terraform plan
terraform apply

# Verify outputs
terraform output

# Test updates
# Edit main.tf
terraform apply  # Should show "update in-place"

# Cleanup
terraform destroy
```

### Verify Using APIs

Use data sources or direct API calls to verify entities exist (automated alternative to dashboard):

#### Option 1: Data Sources (Recommended)
```hcl
# Add to your example to verify creation
data "portkey_workspaces" "verify" {
  depends_on = [portkey_workspace.test]
}

output "exists" {
  value = length([for w in data.portkey_workspaces.verify.workspaces : w if w.id == portkey_workspace.test.id]) > 0
}
```

#### Option 2: Direct API Calls
```bash
# Verify CREATE
curl -s -H "x-portkey-api-key: $PORTKEY_API_KEY" \
  "https://api.portkey.ai/v1/admin/workspaces/$ID" | jq -e '.id' \
  && echo "✅ Created" || echo "❌ Not found"

# Verify UPDATE (check updated field)
curl -s -H "x-portkey-api-key: $PORTKEY_API_KEY" \
  "https://api.portkey.ai/v1/admin/workspaces/$ID" | jq -e '.name == "Updated Name"' \
  && echo "✅ Updated" || echo "❌ Update failed"

# Verify DELETE (should 404)
curl -s -o /dev/null -w "%{http_code}" -H "x-portkey-api-key: $PORTKEY_API_KEY" \
  "https://api.portkey.ai/v1/admin/workspaces/$ID" | grep -q "404" \
  && echo "✅ Deleted" || echo "❌ Still exists"
```

#### Option 3: Verification Script
Add to `examples/<feature>/verify.sh`:
```bash
#!/bin/bash
set -e

# Apply
terraform apply -auto-approve
ID=$(terraform output -raw resource_id)

# Verify CREATE via API
curl -sf -H "x-portkey-api-key: $PORTKEY_API_KEY" \
  "https://api.portkey.ai/v1/admin/workspaces/$ID" > /dev/null \
  && echo "✅ CREATE verified"

# Update config, apply, verify UPDATE
# ...

# Destroy and verify DELETE
terraform destroy -auto-approve
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "x-portkey-api-key: $PORTKEY_API_KEY" \
  "https://api.portkey.ai/v1/admin/workspaces/$ID")
[ "$STATUS" = "404" ] && echo "✅ DELETE verified"
```

### Verification Checklist
- ✅ Outputs show expected values
- ✅ `terraform show` displays correct state
- ✅ **API confirms resource exists** (CREATE)
- ✅ **API confirms fields updated** (UPDATE)  
- ✅ **API returns 404** (DELETE)
- ✅ Updates work in-place (no recreation)

---

## Step 6: Deploy

### Update Documentation
1. Add resource to `README.md`
2. Update `docs/AVAILABLE_APIS.md`
3. Add example README

### Commit & Push
```bash
git add .
git commit -m "feat: add virtual_key resource"
git push
```

### CI Checks
- ✅ `go build ./...` passes
- ✅ `go test ./...` passes
- ✅ Examples validate

### Release
Push a version tag to trigger release:
```bash
git tag v0.2.0
git push origin v0.2.0
```

---

## Pre-Implementation: Test API First

**Always test the API with curl before implementing:**

```bash
# Test CREATE - discover required fields
curl -s -X POST "https://api.portkey.ai/v1/<endpoint>" \
  -H "x-portkey-api-key: $PORTKEY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}' | jq .

# Check the albus codebase for validation rules
grep -r "body.*notEmpty\|body.*isString" src/api/v2/<endpoint>/routes.js
```

**Common surprises discovered this way:**
- `key` is required for integrations (most AI providers need an API key)
- Workspace DELETE requires `{"name": "..."}` in body
- User UPDATE rejects same-role updates (no-op detection)
- Some endpoints use `slug` instead of `id` for lookups

---

## Slug vs ID Resources

Some resources use **slug** as the primary identifier:

| Resource | Identifier | Example |
|----------|------------|---------|
| Workspaces | ID (UUID) | `ws-abc-123456` |
| Integrations | Slug | `my-openai-prod` |

**For slug-based resources:**

```go
// Import by slug, not ID
func (r *integrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

// Read uses slug from state
func (r *integrationResource) Read(ctx context.Context, ...) {
    integration, err := r.client.GetIntegration(ctx, state.Slug.ValueString())
    
    // IMPORTANT: Always set slug in Read, or import tests fail!
    state.Slug = types.StringValue(integration.Slug)
}
```

---

## Common Pitfalls

### 1. Import State Verification Fails
**Symptom:** `ImportStateVerify attributes not equivalent`

**Cause:** Forgot to set a field in `Read()` function

**Fix:** Ensure Read sets ALL fields from API response:
```go
state.ID = types.StringValue(integration.ID)
state.Slug = types.StringValue(integration.Slug)  // Don't forget!
state.Name = types.StringValue(integration.Name)
```

### 2. Create Returns "Invalid request"
**Symptom:** `AB01: Invalid request. Please check and try again.`

**Cause:** Missing required field or wrong format

**Fix:** Check albus routes.js for validation:
```javascript
// Example: integrations require one of these
oneOf([
    body('workspace_id').notEmpty(),
    body('organisation_id').notEmpty().isUUID(),
    header('x-portkey-api-key').notEmpty(),
]),
body('key').optional().isString(),  // But controller requires it!
```

### 3. Sensitive Fields
**For write-only fields (like API keys):**
```go
"key": schema.StringAttribute{
    Description: "API key (write-only, not returned by API)",
    Optional:    true,
    Sensitive:   true,  // Masks in logs/output
},
```

In tests, ignore on import:
```go
ImportStateVerifyIgnore: []string{"key", "created_at", "updated_at"},
```

---

## Acceptance Tests

**File:** `internal/provider/<name>_resource_test.go`

```go
func TestAccIntegrationResource_basic(t *testing.T) {
    rName := acctest.RandomWithPrefix("tf-acc-test")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create
            {
                Config: testAccConfig(rName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("portkey_integration.test", "id"),
                    resource.TestCheckResourceAttr("portkey_integration.test", "name", rName),
                ),
            },
            // Import
            {
                ResourceName:            "portkey_integration.test",
                ImportState:             true,
                ImportStateVerify:       true,
                ImportStateVerifyIgnore: []string{"key"},  // Write-only fields
            },
            // Update
            {
                Config: testAccConfig(rName + "-updated"),
                Check: resource.TestCheckResourceAttr("portkey_integration.test", "name", rName+"-updated"),
            },
            // Delete happens automatically
        },
    })
}
```

**Run tests:**
```bash
export PORTKEY_API_KEY="your-key"
TF_ACC=1 go test ./internal/provider -v -run TestAccIntegration -timeout 10m
```

---

## Quick Checklist

```
□ TEST API FIRST with curl (discover required fields!)
□ Client methods in client.go (Create, Get, Update, Delete, List)
□ Resource file created (internal/provider/<name>_resource.go)
□ Data source file if needed (internal/provider/<name>_data_source.go)
□ Registered in provider.go
□ Acceptance tests written and passing
□ Build passes: make install
□ API verification:
  □ CREATE: API returns resource after apply
  □ UPDATE: API shows updated fields after modify+apply
  □ DELETE: API returns 404 after destroy
□ Import works correctly
□ RESOURCE_MATRIX.md updated
□ Committed and pushed
```

---

## File Reference

| What | Where |
|------|-------|
| API Client | `internal/client/client.go` |
| Resources | `internal/provider/<name>_resource.go` |
| Data Sources | `internal/provider/<name>_data_source.go` |
| Registration | `internal/provider/provider.go` |
| Examples | `examples/<feature>/main.tf` |
| Docs | `README.md`, `docs/AVAILABLE_APIS.md` |

---

## Next APIs to Add

| Priority | API | Endpoint | Notes |
|----------|-----|----------|-------|
| ✅ | Integrations | `/integrations` | Done - AI provider connections |
| 1 | Virtual Keys | `/virtual-keys` | Core feature, workspace-scoped |
| 2 | Configs | `/configs` | Gateway routing |
| 3 | Prompts | `/prompts` | Template management |
| 4 | API Keys | `/api-keys` | Access management |

