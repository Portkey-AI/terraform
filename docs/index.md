# Provider Configuration

## Overview

The Portkey provider is used to interact with Portkey's Admin API for managing workspaces, users, and organization resources.

## Example Usage

```hcl
terraform {
  required_providers {
    portkey = {
      source  = "portkey-ai/portkey"
      version = "~> 0.1"
    }
  }
}

# Configure the Portkey Provider
provider "portkey" {
  api_key = var.portkey_api_key
  # base_url is optional, defaults to https://api.portkey.ai/v1
}
```

## Argument Reference

The following arguments are supported:

* `api_key` - (Optional) The Admin API key for Portkey. Can also be set via the `PORTKEY_API_KEY` environment variable. This key must be an Organization Admin API key with appropriate permissions.

* `base_url` - (Optional) The base URL for the Portkey API. Defaults to `https://api.portkey.ai/v1`. Can also be set via the `PORTKEY_BASE_URL` environment variable. Use this for self-hosted Portkey deployments.

## Authentication

The provider supports two authentication methods:

### Environment Variables (Recommended)

```bash
export PORTKEY_API_KEY="your-admin-api-key"
terraform plan
```

### Provider Block Configuration

```hcl
provider "portkey" {
  api_key = "your-admin-api-key"
}
```

**Security Note**: Never commit API keys to version control. Use environment variables or a secrets management solution.

## Self-Hosted Deployments

For self-hosted Portkey Control Plane deployments:

```hcl
provider "portkey" {
  api_key  = var.portkey_api_key
  base_url = "https://portkey.your-company.com/v1"
}
```

Alternatively, set the environment variable:

```bash
export PORTKEY_BASE_URL="https://portkey.your-company.com/v1"
```

## Creating Admin API Keys

To use this provider, you need an Organization Admin API key:

1. Log in to your Portkey dashboard
2. Navigate to **Admin Settings**
3. Go to the **API Keys** section
4. Click **Create API Key**
5. Select **Organization** as the type
6. Select **Service** as the sub-type
7. Grant necessary scopes (or select all for full access)
8. Save the API key securely

**Important**: Organization Admin API keys provide broad access to your organization. Only Organization Owners and Admins can create these keys.

## Required Permissions

The Admin API key used with this provider should have the following scopes:

### For Full Provider Functionality
- `workspaces.create`
- `workspaces.read`
- `workspaces.update`
- `workspaces.delete`
- `workspaces.list`
- `users.read`
- `users.update`
- `users.delete`
- `users.list`
- `users.invite`

### Minimum Required Scopes by Resource

#### portkey_workspace
- `workspaces.create`
- `workspaces.read`
- `workspaces.update`
- `workspaces.delete`

#### portkey_workspace_member
- `workspaces.read`
- `workspaces.members.add`
- `workspaces.members.update`
- `workspaces.members.remove`

#### portkey_user_invite
- `users.invite`
- `users.read`

#### Data Sources
- `workspaces.read` / `workspaces.list` for workspace data sources
- `users.read` / `users.list` for user data sources

## Rate Limiting

The Portkey Admin API implements rate limiting. The provider respects these limits and will return appropriate errors if exceeded. Consider:

- Using `terraform apply` with the `-parallelism=1` flag for large deployments
- Implementing gradual rollouts for bulk operations
- Adding delays between resource creations if needed

## Timeouts

The provider uses a default timeout of 30 seconds for API requests. If you experience timeout issues with large operations, contact Portkey support.

## Error Handling

Common error scenarios:

### Authentication Errors
```
Error: Unable to Create Portkey API Client
API request failed with status 401: Unauthorized
```
**Solution**: Verify your API key is correct and has Organization Admin privileges.

### Permission Errors
```
Error: API request failed with status 403: Forbidden
```
**Solution**: Ensure your API key has the required scopes for the operation.

### Resource Not Found
```
Error: API request failed with status 404: Not Found
```
**Solution**: Verify the resource ID exists and is accessible with your API key.

### Workspace Deletion Error
```
Error: Cannot delete workspace with existing resources
```
**Solution**: Remove all resources from the workspace before deletion. This includes virtual keys, configs, and members.

## Best Practices

1. **Use Environment Variables**: Store API keys in environment variables, not in code
2. **Limit Scope**: Create API keys with minimum required scopes
3. **Separate Keys**: Use different API keys for different environments
4. **Rotate Keys**: Regularly rotate API keys
5. **Audit Logs**: Review audit logs for all Admin API operations
6. **State Management**: Use remote state with encryption for production
7. **Version Control**: Pin provider version in production

## Example: Secure Configuration

```hcl
# variables.tf
variable "portkey_api_key" {
  description = "Portkey Admin API Key"
  type        = string
  sensitive   = true
}

# main.tf
terraform {
  required_providers {
    portkey = {
      source  = "portkey-ai/portkey"
      version = "0.1.0"
    }
  }
  
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "portkey/terraform.tfstate"
    region = "us-east-1"
    encrypt = true
  }
}

provider "portkey" {
  api_key = var.portkey_api_key
}

# Pass API key via environment variable or secure secrets manager
# terraform plan -var="portkey_api_key=$PORTKEY_API_KEY"
```

## Troubleshooting

Enable detailed logging:

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
terraform plan
```

This will log all API requests and responses for debugging.

## Related Documentation

- [Portkey Admin API Documentation](https://portkey.ai/docs/api-reference/admin-api/introduction)
- [Portkey Organization Management](https://portkey.ai/docs/product/enterprise-offering/org-management)
- [Portkey API Keys Guide](https://portkey.ai/docs/product/enterprise-offering/org-management/api-keys-authn-and-authz)
