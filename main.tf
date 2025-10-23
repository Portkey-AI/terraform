terraform {
  required_providers {
    portkey = {
      source = "portkey-ai/portkey"
    }
  }
}

provider "portkey" {
  # API key can also be set via PORTKEY_API_KEY environment variable
  api_key = var.portkey_api_key

  # Optional: Override base URL for self-hosted deployments
  # base_url = "https://your-portkey-instance.com/v1"
}

variable "portkey_api_key" {
  description = "Portkey Admin API Key"
  type        = string
  sensitive   = true
}

# Create a workspace
resource "portkey_workspace" "engineering" {
  name        = "Engineering Team"
  description = "Workspace for the engineering team"
}

resource "portkey_workspace" "data_science" {
  name        = "Data Science"
  description = "Workspace for data science projects"
}

# Invite a user to the organization
resource "portkey_user_invite" "john" {
  email = "john@example.com"
  role  = "member"

  # Add user to specific workspaces with roles
  workspaces = [
    {
      id   = portkey_workspace.engineering.id
      role = "admin"
    },
    {
      id   = portkey_workspace.data_science.id
      role = "member"
    }
  ]

  # Grant specific API scopes
  scopes = [
    "logs.export",
    "logs.list",
    "logs.view",
    "configs.read",
    "configs.list"
  ]
}

# Add a workspace member (for existing users)
# Note: This requires the user to already exist in the organization
resource "portkey_workspace_member" "alice_engineering" {
  workspace_id = portkey_workspace.engineering.id
  user_id      = "user-id-from-portkey"
  role         = "manager"
}

# Data sources for reading existing resources
data "portkey_workspaces" "all" {}

data "portkey_workspace" "specific" {
  id = portkey_workspace.engineering.id
}

data "portkey_users" "all" {}

data "portkey_user" "specific" {
  id = "user-id-from-portkey"
}

# Outputs
output "engineering_workspace_id" {
  description = "ID of the engineering workspace"
  value       = portkey_workspace.engineering.id
}

output "all_workspaces" {
  description = "All workspaces in the organization"
  value       = data.portkey_workspaces.all.workspaces
}

output "user_invite_status" {
  description = "Status of John's invitation"
  value       = portkey_user_invite.john.status
}
