terraform {
  required_providers {
    portkey = {
      source  = "registry.terraform.io/portkey-ai/portkey"
      version = "0.1.0"
    }
  }
}

provider "portkey" {
  # Using PORTKEY_API_KEY environment variable
}

# Test 1: Create multiple workspaces
resource "portkey_workspace" "dev" {
  name        = "Development Workspace"
  description = "Updated description for testing"
}

resource "portkey_workspace" "staging" {
  name        = "Staging Workspace"
  description = "For staging environment"
}

resource "portkey_workspace" "prod" {
  name        = "Production Workspace"
  description = "For production environment"
}

# Test 2: Data source - Read a specific workspace
data "portkey_workspace" "dev_read" {
  id = portkey_workspace.dev.id
}

# Test 3: Data source - List all workspaces
data "portkey_workspaces" "all" {}

# Test 4: Data source - List all users
data "portkey_users" "all" {}

# Outputs to verify everything
output "created_workspaces" {
  description = "All created workspaces"
  value = {
    dev = {
      id          = portkey_workspace.dev.id
      name        = portkey_workspace.dev.name
      description = portkey_workspace.dev.description
      created_at  = portkey_workspace.dev.created_at
    }
    staging = {
      id          = portkey_workspace.staging.id
      name        = portkey_workspace.staging.name
      description = portkey_workspace.staging.description
      created_at  = portkey_workspace.staging.created_at
    }
    prod = {
      id          = portkey_workspace.prod.id
      name        = portkey_workspace.prod.name
      description = portkey_workspace.prod.description
      created_at  = portkey_workspace.prod.created_at
    }
  }
}

output "data_source_workspace" {
  description = "Workspace read via data source"
  value = {
    id          = data.portkey_workspace.dev_read.id
    name        = data.portkey_workspace.dev_read.name
    description = data.portkey_workspace.dev_read.description
  }
}

output "all_workspaces_count" {
  description = "Total number of workspaces in organization"
  value       = length(data.portkey_workspaces.all.workspaces)
}

output "all_users_count" {
  description = "Total number of users in organization"
  value       = length(data.portkey_users.all.users)
}

output "test_summary" {
  description = "Summary of test results"
  value = {
    resources_created      = 3
    data_sources_tested    = 3
    workspace_resource_ids = [
      portkey_workspace.dev.id,
      portkey_workspace.staging.id,
      portkey_workspace.prod.id
    ]
  }
}

