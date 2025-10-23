# Portkey Admin API - Available Endpoints

Based on API discovery and testing with your admin API key.

## ✅ Currently Implemented in Provider

These are already working in the Terraform provider:

### 1. Workspaces Management
- **Endpoint**: `/admin/workspaces`
- **Status**: ✅ Implemented
- **Operations**: Create, Read, Update, Delete, List
- **Terraform Resources**:
  - `portkey_workspace` (resource)
  - `portkey_workspaces` (data source)
  - `portkey_workspace` (data source - single)

### 2. Users Management
- **Endpoint**: `/admin/users`
- **Status**: ✅ Implemented (Read-only)
- **Operations**: List, Read, Update, Delete
- **Terraform Resources**:
  - `portkey_users` (data source)
  - `portkey_user` (data source - single)

### 3. User Invites
- **Endpoint**: `/admin/users/invites`
- **Status**: ✅ Implemented
- **Operations**: Create, Read, List, Delete
- **Terraform Resource**: `portkey_user_invite`

### 4. Workspace Members
- **Endpoint**: `/admin/workspaces/{id}/members`
- **Status**: ✅ Implemented
- **Operations**: Add, List, Read, Update, Remove
- **Terraform Resource**: `portkey_workspace_member`

## 🔓 Available but NOT Implemented

These APIs are accessible and could be added to the provider:

### 5. Virtual Keys (Workspace-level)
- **Endpoint**: `/virtual-keys`
- **Status**: 🟡 Available, Not Implemented
- **Access Level**: Workspace (not admin)
- **Purpose**: Manage API keys for LLM providers
- **Potential Resource**: `portkey_virtual_key`
- **Priority**: HIGH

**What it does**: Virtual Keys allow you to manage API keys for different LLM providers (OpenAI, Anthropic, etc.) in a centralized way.

### 6. Configs (Workspace-level)
- **Endpoint**: `/configs`
- **Status**: 🟡 Available, Not Implemented
- **Access Level**: Workspace (not admin)
- **Purpose**: Manage Gateway configurations (routing, fallbacks, load balancing)
- **Potential Resource**: `portkey_config`
- **Priority**: HIGH

**What it does**: Configs define how requests are routed, with features like:
- Load balancing across providers
- Fallback strategies
- Retry logic
- Caching rules

### 7. Prompts (Workspace-level)
- **Endpoint**: `/prompts`
- **Status**: 🟡 Available, Not Implemented
- **Access Level**: Workspace (not admin)
- **Purpose**: Manage prompt templates
- **Potential Resource**: `portkey_prompt`
- **Priority**: MEDIUM

**What it does**: Store and version prompt templates for LLM applications.

### 8. API Keys (Workspace-level)
- **Endpoint**: `/api-keys`
- **Status**: 🟡 Available, Not Implemented
- **Access Level**: Workspace (not admin)
- **Purpose**: Manage Portkey API keys
- **Potential Resource**: `portkey_api_key`
- **Priority**: MEDIUM

**What it does**: Create and manage API keys for accessing Portkey services.

## 🔒 Restricted / Permission-Based Endpoints

These endpoints exist but require additional permissions or organization-level access:

### Admin-level (Require elevated permissions)
- `/admin/virtual-keys` - 403 Forbidden
- `/admin/api-keys` - 403 Forbidden
- `/admin/configs` - 403 Forbidden
- `/admin/prompts` - 403 Forbidden
- `/admin/guardrails` - 403 Forbidden
- `/admin/providers` - 403 Forbidden
- `/admin/integrations` - 403 Forbidden
- `/admin/plugins` - 403 Forbidden
- `/admin/audit-logs` - 403 Forbidden
- `/admin/organization` - 403 Forbidden
- `/admin/billing` - 403 Forbidden
- `/admin/usage` - 403 Forbidden
- `/admin/webhooks` - 403 Forbidden

### Organization-level
- `/organisation` - 403 Forbidden
- `/organisation/users` - 403 Forbidden
- `/organisation/api-keys` - 403 Forbidden
- `/organisation/billing` - 403 Forbidden

### Workspace-level (Restricted)
- `/guardrails` - 403 Forbidden
- `/logs` - 403 Forbidden
- `/analytics` - 404 Not Found

## 🎯 Recommended Next Resources to Implement

Based on availability and usefulness:

### Priority 1: Virtual Keys ⭐⭐⭐
```hcl
resource "portkey_virtual_key" "openai" {
  workspace_id = portkey_workspace.dev.id
  provider     = "openai"
  key          = var.openai_api_key
  name         = "OpenAI Production Key"
  rate_limit   = 100
  budget_limit = 1000
}
```

**Why**: Core feature for managing LLM provider keys centrally.

### Priority 2: Configs ⭐⭐⭐
```hcl
resource "portkey_config" "production" {
  workspace_id = portkey_workspace.prod.id
  name         = "Production Gateway Config"
  strategy     = "loadbalance"
  targets      = [
    {
      provider     = "openai"
      virtual_key  = portkey_virtual_key.openai.id
      weight       = 70
    },
    {
      provider     = "anthropic"
      virtual_key  = portkey_virtual_key.anthropic.id
      weight       = 30
    }
  ]
  retry = {
    attempts = 3
  }
}
```

**Why**: Essential for production deployments with fallbacks and load balancing.

### Priority 3: Prompts ⭐⭐
```hcl
resource "portkey_prompt" "customer_support" {
  workspace_id = portkey_workspace.prod.id
  name         = "Customer Support Agent"
  template     = "You are a helpful customer support agent..."
  variables    = ["customer_name", "issue_type"]
}
```

**Why**: Useful for managing prompt versions in infrastructure as code.

### Priority 4: API Keys ⭐
```hcl
resource "portkey_api_key" "app_backend" {
  workspace_id = portkey_workspace.prod.id
  name         = "Backend Application Key"
  scopes       = ["logs.view", "configs.read"]
  rate_limit   = 1000
}
```

**Why**: Automate API key creation for applications.

## 📊 API Architecture

```
Portkey API Hierarchy:

Organization Level
├── /admin/* (Org-wide management, requires admin key)
│   ├── /admin/workspaces ✅ Implemented
│   ├── /admin/users ✅ Implemented
│   └── /admin/users/invites ✅ Implemented
│
└── Workspace Level (per-workspace resources)
    ├── /virtual-keys 🟡 Available
    ├── /configs 🟡 Available
    ├── /prompts 🟡 Available
    ├── /api-keys 🟡 Available
    ├── /guardrails 🔒 Restricted
    └── /logs 🔒 Restricted
```

## 🔄 Implementation Workflow

To add a new resource to the provider:

### 1. Add Client Methods
```go
// internal/client/client.go
func (c *Client) CreateVirtualKey(ctx context.Context, req CreateVirtualKeyRequest) (*VirtualKey, error) {
    respBody, err := c.doRequest(ctx, http.MethodPost, "/virtual-keys", req)
    // ...
}
```

### 2. Create Resource Implementation
```go
// internal/provider/virtual_key_resource.go
type virtualKeyResource struct {
    client *client.Client
}

func (r *virtualKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Implementation
}
```

### 3. Register in Provider
```go
// internal/provider/provider.go
func (p *portkeyProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewWorkspaceResource,
        NewVirtualKeyResource, // Add new resource
        // ...
    }
}
```

## 📝 API Documentation Links

For detailed API documentation:
- Official Docs: https://docs.portkey.ai/
- API Reference: https://portkey.ai/docs/api-reference

## 🎯 Summary

### Currently Working (4 resources, 4 data sources)
✅ Workspaces
✅ Workspace Members  
✅ Users (read-only)
✅ User Invites

### Ready to Implement (4 resources)
🟡 Virtual Keys - HIGH PRIORITY
🟡 Configs - HIGH PRIORITY
🟡 Prompts - MEDIUM PRIORITY
🟡 API Keys - MEDIUM PRIORITY

### Needs Investigation (13+ endpoints)
🔒 Various admin and organization-level endpoints
🔒 May require different authentication or permissions
🔒 May be enterprise-only features

## 💡 Next Steps

1. **Implement Virtual Keys** - Most commonly used feature
2. **Implement Configs** - Critical for production deployments
3. **Add Prompts** - Useful for version control
4. **Add API Keys** - Complete the workspace management suite
5. **Investigate restricted endpoints** - Contact Portkey for access requirements

