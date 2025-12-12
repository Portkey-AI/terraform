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

## Quick Checklist

```
□ Client methods in client.go (Create, Get, Update, Delete, List)
□ Resource file created (internal/provider/<name>_resource.go)
□ Data source file if needed (internal/provider/<name>_data_source.go)
□ Registered in provider.go
□ Example created (examples/<feature>/main.tf)
□ Build passes: make build
□ Example works: terraform apply
□ API verification:
  □ CREATE: API returns resource after apply
  □ UPDATE: API shows updated fields after modify+apply
  □ DELETE: API returns 404 after destroy
□ Updates work in-place (not recreate)
□ README.md updated
□ AVAILABLE_APIS.md updated
□ Committed and pushed
□ Tagged for release
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
| 1 | Virtual Keys | `/virtual-keys` | Core feature |
| 2 | Configs | `/configs` | Gateway routing |
| 3 | Prompts | `/prompts` | Template management |
| 4 | API Keys | `/api-keys` | Access management |

