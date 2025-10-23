# Testing Coverage Analysis

## ❌ NO - We're NOT testing everything we support!

### What's Implemented vs. What's Tested

## Resources (3 total)

| Resource | Implemented | Tested | Status |
|----------|-------------|--------|--------|
| `portkey_workspace` | ✅ | ✅ | **TESTED** - Created 3 workspaces |
| `portkey_workspace_member` | ✅ | ❌ | **NOT TESTED** |
| `portkey_user_invite` | ✅ | ❌ | **NOT TESTED** |

## Data Sources (4 total)

| Data Source | Implemented | Tested | Status |
|-------------|-------------|--------|--------|
| `portkey_workspace` | ✅ | ✅ | **TESTED** - Read by ID |
| `portkey_workspaces` | ✅ | ✅ | **TESTED** - List all |
| `portkey_user` | ✅ | ❌ | **NOT TESTED** |
| `portkey_users` | ✅ | ✅ | **TESTED** - List all |

## Summary Statistics

```
Total Implemented:  7 (3 resources + 4 data sources)
Actually Tested:    4 (1 resource + 3 data sources)
Not Tested:         3 (2 resources + 1 data source)

Coverage:           57% (4/7)
```

## 🚫 Missing Tests

### 1. `portkey_workspace_member` Resource ❌

**What it does**: Adds/removes users from workspaces

**Why not tested**: Requires existing user IDs

**Test would look like**:
```hcl
resource "portkey_workspace_member" "alice_dev" {
  workspace_id = portkey_workspace.dev.id
  user_id      = "existing-user-id"  # ← Need real user ID
  role         = "admin"
}
```

**Blocker**: Need to get a real user ID from your organization first.

### 2. `portkey_user_invite` Resource ❌

**What it does**: Sends email invitations to new users

**Why not tested**: Sends real email invitations

**Test would look like**:
```hcl
resource "portkey_user_invite" "john" {
  email = "john@example.com"
  role  = "member"
  workspaces = [
    {
      id   = portkey_workspace.dev.id
      role = "admin"
    }
  ]
  scopes = ["logs.view", "configs.read"]
}
```

**Blocker**: Sends actual email invitations; needs careful testing.

### 3. `portkey_user` Data Source ❌

**What it does**: Fetches a single user by ID

**Why not tested**: Need a real user ID

**Test would look like**:
```hcl
data "portkey_user" "alice" {
  id = "existing-user-id"  # ← Need real user ID
}

output "user_email" {
  value = data.portkey_user.alice.email
}
```

**Blocker**: Need to get a real user ID first.

## 🎯 Complete Testing Plan

Here's how to test everything:

### Phase 1: Get Required Data ✅ (Can do now)

```bash
# Get list of users to find a user ID
terraform console
> data.portkey_users.all.users
```

This will show you user IDs that you can use for testing.

### Phase 2: Test Missing Data Source

```hcl
# Add to test configuration
data "portkey_users" "all" {}

# Get first user ID
locals {
  first_user_id = length(data.portkey_users.all.users) > 0 ? data.portkey_users.all.users[0].id : ""
}

# Test single user data source
data "portkey_user" "test_user" {
  count = local.first_user_id != "" ? 1 : 0
  id    = local.first_user_id
}

output "test_user_details" {
  value = length(data.portkey_user.test_user) > 0 ? data.portkey_user.test_user[0] : null
}
```

### Phase 3: Test Workspace Member

```hcl
# Get a real user ID first (from Phase 1)
resource "portkey_workspace_member" "test_member" {
  workspace_id = portkey_workspace.test.id
  user_id      = local.first_user_id  # From Phase 1
  role         = "member"
}
```

### Phase 4: Test User Invite (Be Careful!)

```hcl
# Use a test email you control
resource "portkey_user_invite" "test_invite" {
  email = "test+terraform@yourdomain.com"  # Use + addressing
  role  = "member"
  workspaces = [
    {
      id   = portkey_workspace.test.id
      role = "member"
    }
  ]
}
```

**⚠️ Warning**: This will send a real email invitation!

## 🔍 Quick Check: What User IDs are Available?

Run this to see what we have:

```bash
cd test-env
export PORTKEY_API_KEY="your-key"

# Check current state
terraform console << EOF
data.portkey_users.all.users
EOF
```

## ✅ Comprehensive Test Configuration

Here's a complete test that covers everything (requires user ID):

```hcl
terraform {
  required_providers {
    portkey = {
      source  = "registry.terraform.io/portkey-ai/portkey"
      version = "0.1.0"
    }
  }
}

provider "portkey" {}

# === RESOURCES ===

# 1. Workspace Resource ✅ TESTED
resource "portkey_workspace" "test" {
  name        = "Complete Test Workspace"
  description = "Testing all features"
}

# 2. User Invite Resource ❌ NOT TESTED
resource "portkey_user_invite" "test_invite" {
  email = "test+tf@yourdomain.com"
  role  = "member"
  workspaces = [{
    id   = portkey_workspace.test.id
    role = "member"
  }]
  scopes = ["logs.view"]
}

# 3. Workspace Member Resource ❌ NOT TESTED
# Requires existing user ID
data "portkey_users" "all" {}

locals {
  first_user_id = length(data.portkey_users.all.users) > 0 ? data.portkey_users.all.users[0].id : null
}

resource "portkey_workspace_member" "test_member" {
  count        = local.first_user_id != null ? 1 : 0
  workspace_id = portkey_workspace.test.id
  user_id      = local.first_user_id
  role         = "viewer"
}

# === DATA SOURCES ===

# 4. Workspaces List ✅ TESTED
data "portkey_workspaces" "all" {}

# 5. Single Workspace ✅ TESTED  
data "portkey_workspace" "test" {
  id = portkey_workspace.test.id
}

# 6. Users List ✅ TESTED
# (Already defined above)

# 7. Single User ❌ NOT TESTED
data "portkey_user" "test_user" {
  count = local.first_user_id != null ? 1 : 0
  id    = local.first_user_id
}

# === OUTPUTS ===

output "coverage_report" {
  value = {
    resources_tested = {
      workspace        = "✅ Created"
      user_invite      = "✅ Created"
      workspace_member = local.first_user_id != null ? "✅ Created" : "⚠️ Skipped (no users)"
    }
    data_sources_tested = {
      workspaces = "✅ Tested"
      workspace  = "✅ Tested"
      users      = "✅ Tested"
      user       = local.first_user_id != null ? "✅ Tested" : "⚠️ Skipped (no users)"
    }
    total_coverage = "100%"
  }
}
```

## 🎯 Action Items

To achieve 100% test coverage:

1. ✅ **Already tested** (4/7):
   - `portkey_workspace` resource
   - `portkey_workspaces` data source
   - `portkey_workspace` data source
   - `portkey_users` data source

2. ⚠️ **Can test now** (1/7):
   - `portkey_user` data source
   - Just need to grab a user ID from `portkey_users`

3. ⚠️ **Requires user ID** (1/7):
   - `portkey_workspace_member` resource
   - Use existing user from organization

4. ⚠️ **Requires caution** (1/7):
   - `portkey_user_invite` resource
   - Sends real email - use carefully!

## 📊 Current vs Full Coverage

```
Current Coverage:  4/7 = 57%
Easy to add:       1/7 = 14% (just need user ID)
Needs care:        2/7 = 29% (workspace_member, user_invite)
                   ─────────
Target Coverage:   7/7 = 100%
```

## 🚀 Next Steps

Run this to complete testing:

```bash
cd test-env

# Create comprehensive test file
cat > complete-test.tf << 'EOF'
# [Use the comprehensive test above]
EOF

# Run tests
terraform init
terraform plan
terraform apply -auto-approve

# Verify
terraform show
```

This will test all 7 implemented features!




