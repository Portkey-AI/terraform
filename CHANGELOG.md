# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **AI Gateway Resources:**
  - `portkey_integration` - Manage AI provider integrations (OpenAI, Anthropic, Azure, etc.)
  - `portkey_provider` - Manage providers/virtual keys for workspace-scoped AI access
  - `portkey_config` - Manage gateway configurations with routing and fallbacks
  - `portkey_prompt` - Manage versioned prompt templates
- **Governance Resources:**
  - `portkey_guardrail` - Set up content validation and safety checks
  - `portkey_usage_limits_policy` - Control costs with spending limits
  - `portkey_rate_limits_policy` - Manage request rate limiting
- **Access Control Resources:**
  - `portkey_api_key` - Create and manage Portkey API keys
- **Data Sources for all new resources:**
  - `portkey_integration`, `portkey_integrations`
  - `portkey_provider`, `portkey_providers`
  - `portkey_config`, `portkey_configs`
  - `portkey_prompt`, `portkey_prompts`
  - `portkey_guardrail`, `portkey_guardrails`
  - `portkey_usage_limits_policy`, `portkey_usage_limits_policies`
  - `portkey_rate_limits_policy`, `portkey_rate_limits_policies`
  - `portkey_api_key`, `portkey_api_keys`

## [0.1.0] - TBD

### Added
- Initial release of the Portkey Terraform Provider
- **Organization Resources:**
  - `portkey_workspace` - Manage Portkey workspaces
  - `portkey_workspace_member` - Manage workspace membership
  - `portkey_user_invite` - Send user invitations with workspace access and scopes
- **Data Sources:**
  - `portkey_workspace` - Query single workspace by ID
  - `portkey_workspaces` - List all workspaces in organization
  - `portkey_user` - Query single user by ID
  - `portkey_users` - List all users in organization
- Provider configuration with API key authentication
- Support for environment variable `PORTKEY_API_KEY`
- Import functionality for all resources
- Comprehensive documentation and examples
- Multi-environment setup example

### Supported Operations
- Full CRUD operations for workspaces
- User invitation with granular scope management
- Workspace member role assignment
- Organization and workspace role management

### Known Limitations
- User invitations cannot be updated (must delete and recreate)
- Workspace deletion may be blocked by existing resources
- Prompt template updates create new versions (use makeDefault to promote)

[Unreleased]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Portkey-AI/terraform-provider-portkey/releases/tag/v0.1.0

