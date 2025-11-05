# Terraform Provider for Portkey

[![Terraform Registry](https://img.shields.io/badge/registry-terraform-blue.svg)](https://registry.terraform.io/providers/portkey-ai/portkey/latest)
[![CI](https://github.com/Portkey-AI/terraform-provider-portkey/actions/workflows/ci.yml/badge.svg)](https://github.com/Portkey-AI/terraform-provider-portkey/actions/workflows/ci.yml)
[![Acceptance Tests](https://github.com/Portkey-AI/terraform-provider-portkey/actions/workflows/acc-tests.yml/badge.svg)](https://github.com/Portkey-AI/terraform-provider-portkey/actions/workflows/acc-tests.yml)

A Terraform provider for managing Portkey workspaces, users, and organization resources through the [Portkey Admin API](https://portkey.ai/docs/api-reference/admin-api/introduction).

## Features

This provider enables you to manage:

- **Workspaces**: Create, update, and manage workspaces for organizing teams and projects
- **Workspace Members**: Assign users to workspaces with specific roles
- **User Invitations**: Send invitations to users with organization and workspace access
- **Users**: Query and manage existing users in your organization

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- A Portkey account with Admin API access

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    portkey = {
      source  = "portkey-ai/portkey"
      version = "~> 0.1"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/Portkey-AI/terraform-provider-portkey
cd terraform-provider-portkey
make install
```

This will build and install the provider in your local Terraform plugins directory.

## Authentication

The provider requires a Portkey Admin API key. You can provide this in one of two ways:

### Environment Variable (Recommended)

```bash
export PORTKEY_API_KEY="your-admin-api-key"
```

### Provider Configuration

```hcl
provider "portkey" {
  api_key = "your-admin-api-key"
}
```

### Getting Your Admin API Key

1. Log in to your Portkey dashboard
2. Navigate to Admin Settings
3. Create an Organization Admin API key
4. Ensure you have Organization Owner or Admin role

**Note**: Admin API keys provide broad access to your organization. Store them securely and never commit them to version control.

## Usage Examples

### Basic Configuration

```hcl
terraform {
  required_providers {
    portkey = {
      source = "portkey-ai/portkey"
    }
  }
}

provider "portkey" {
  # API key read from PORTKEY_API_KEY environment variable
}

# Create a workspace
resource "portkey_workspace" "production" {
  name        = "Production"
  description = "Production environment workspace"
}
```

### Complete Organization Setup

```hcl
# Create multiple workspaces
resource "portkey_workspace" "engineering" {
  name        = "Engineering"
  description = "Engineering team workspace"
}

resource "portkey_workspace" "ml_research" {
  name        = "ML Research"
  description = "Machine Learning research workspace"
}

# Invite a user with workspace access
resource "portkey_user_invite" "data_scientist" {
  email = "scientist@example.com"
  role  = "member"

  workspaces = [
    {
      id   = portkey_workspace.ml_research.id
      role = "admin"
    },
    {
      id   = portkey_workspace.engineering.id
      role = "member"
    }
  ]

  scopes = [
    "logs.export",
    "logs.list",
    "logs.view",
    "configs.read",
    "configs.list",
    "virtual_keys.read",
    "virtual_keys.list"
  ]
}

# Add an existing user to a workspace
resource "portkey_workspace_member" "senior_engineer" {
  workspace_id = portkey_workspace.engineering.id
  user_id      = "existing-user-id"
  role         = "manager"
}

# Query all workspaces
data "portkey_workspaces" "all" {}

# Query all users
data "portkey_users" "all" {}

output "workspace_count" {
  value = length(data.portkey_workspaces.all.workspaces)
}
```

### Self-Hosted Portkey

For self-hosted Portkey deployments, configure the base URL:

```hcl
provider "portkey" {
  api_key  = var.portkey_api_key
  base_url = "https://your-portkey-instance.com/v1"
}
```

## Resources

### `portkey_workspace`

Manages a Portkey workspace.

#### Arguments

- `name` (Required, String) - Name of the workspace
- `description` (Optional, String) - Description of the workspace

#### Attributes

- `id` (String) - Workspace identifier
- `created_at` (String) - Creation timestamp
- `updated_at` (String) - Last update timestamp

#### Import

```bash
terraform import portkey_workspace.example workspace-id
```

### `portkey_workspace_member`

Manages workspace membership for users.

#### Arguments

- `workspace_id` (Required, String) - ID of the workspace
- `user_id` (Required, String) - ID of the user
- `role` (Required, String) - Role in the workspace (`admin`, `manager`, `member`)

#### Attributes

- `id` (String) - Workspace member identifier
- `created_at` (String) - Timestamp when member was added

#### Import

```bash
terraform import portkey_workspace_member.example workspace-id/member-id
```

### `portkey_user_invite`

Sends invitations to users.

#### Arguments

- `email` (Required, String) - Email address of the user to invite
- `role` (Required, String) - Organization role (`admin`, `member`)
- `workspaces` (Optional, List of Objects) - Workspaces to add the user to
  - `id` (Required, String) - Workspace ID
  - `role` (Required, String) - Role in the workspace
- `scopes` (Optional, List of Strings) - API scopes for the user's workspace API key

#### Attributes

- `id` (String) - Invitation identifier
- `status` (String) - Invitation status (`pending`, `accepted`, `expired`)
- `created_at` (String) - Creation timestamp
- `expires_at` (String) - Expiration timestamp

**Note**: User invitations cannot be updated. To change an invitation, delete and recreate it.

## Data Sources

### `portkey_workspace`

Fetches a single workspace by ID.

#### Arguments

- `id` (Required, String) - Workspace identifier

#### Attributes

- `name` (String) - Workspace name
- `description` (String) - Workspace description
- `created_at` (String) - Creation timestamp
- `updated_at` (String) - Last update timestamp

### `portkey_workspaces`

Fetches all workspaces in the organization.

#### Attributes

- `workspaces` (List of Objects) - List of all workspaces

### `portkey_user`

Fetches a single user by ID.

#### Arguments

- `id` (Required, String) - User identifier

#### Attributes

- `email` (String) - User email
- `role` (String) - Organization role
- `status` (String) - Account status
- `created_at` (String) - Creation timestamp
- `updated_at` (String) - Last update timestamp

### `portkey_users`

Fetches all users in the organization.

#### Attributes

- `users` (List of Objects) - List of all users

## Development

### Building

```bash
make build
```

### Testing

```bash
# Run unit tests
go test ./...

# Run acceptance tests (requires valid API key)
make testacc
```

### Installing Locally

```bash
make install
```

### Generating Documentation

```bash
make generate
```

## API Scopes

When inviting users, you can grant the following API scopes:

- `logs.export`, `logs.list`, `logs.view` - Log access
- `configs.create`, `configs.update`, `configs.delete`, `configs.read`, `configs.list` - Configuration management
- `virtual_keys.create`, `virtual_keys.update`, `virtual_keys.delete`, `virtual_keys.read`, `virtual_keys.list`, `virtual_keys.copy` - Virtual key management
- `completions.write` - Completion API access

## Roles

### Organization Roles

- `admin` - Full organization access
- `member` - Standard user access

### Workspace Roles

- `admin` - Full workspace access
- `manager` - Manage workspace resources and members
- `member` - Standard workspace access

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

This provider is distributed under the Mozilla Public License 2.0. See `LICENSE` for more information.

## Support

- Documentation: [Portkey Docs](https://portkey.ai/docs)
- Admin API Reference: [Admin API Docs](https://portkey.ai/docs/api-reference/admin-api/introduction)
- Issues: [GitHub Issues](https://github.com/Portkey-AI/terraform-provider-portkey/issues)
- Community: [Discord](https://portkey.sh/discord-1)

## Related Projects

- [Portkey Gateway](https://github.com/Portkey-AI/gateway) - Open-source AI Gateway
- [Portkey Python SDK](https://github.com/Portkey-AI/portkey-python-sdk)
- [Portkey Node SDK](https://github.com/Portkey-AI/portkey-node-sdk)
