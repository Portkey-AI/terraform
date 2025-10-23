terraform {
  required_providers {
    portkey = {
      source = "registry.terraform.io/portkey-ai/portkey"
      version = "0.1.0"
    }
  }
}

provider "portkey" {
  # Set your API key via environment variable PORTKEY_API_KEY
  # Or uncomment and set it here:
  # api_key = "your-api-key-here"
}

# Simple test: Create a workspace
resource "portkey_workspace" "test" {
  name        = "Test Workspace"
  description = "Testing the Terraform provider"
}

# Output the workspace ID
output "workspace_id" {
  value = portkey_workspace.test.id
}

