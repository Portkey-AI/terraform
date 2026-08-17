# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **`terraform import` then `terraform destroy` now cascades correctly** — `ImportState` was a bare passthrough, so `force_delete` (which has no API representation) was absent from state after import. Because schema defaults are applied during plan — not during import or destroy — a `terraform destroy` after import always saw `force_delete = false` and skipped cascade deletion, failing with `409 AB07`. The provider now seeds `force_delete = true` in state at import time.

### Changed
- **`force_delete` test uses configs instead of providers** — the acceptance test for cascade deletion previously created an out-of-band provider, which required an integration grant (`TEST_INTEGRATION_ID`) that a freshly created workspace doesn't have. The test now creates an out-of-band config (no grant needed) and also adds a negative case (`force_delete = false` must fail with AB07) and a `CheckDestroy` assertion.

## [0.2.32] - 2026-08-05

### Added
- **`portkey_organisation_defaults` resource** — manage default `input_guardrails` / `output_guardrails` for the whole organisation via `GET`/`PUT /v2/admin/organisation/defaults`. These are the first `/v2` endpoints used by the provider, so the client swaps the `/v1` suffix on the configured `base_url` for `/v2`. The resource is a singleton: the Admin API derives the organisation from the API key, so there is no `organisation_id` attribute, exactly one instance may exist per provider configuration, and `terraform import` ignores the supplied ID (`id` is always `organisation_defaults`). At least one of the two lists must be set — the API rejects a write carrying neither, and the provider surfaces this at plan time rather than as a 400 on apply. Only organisation-scoped guardrails are accepted; passing a workspace-scoped guardrail fails with `Workspace-scoped guardrails cannot be set as organisation defaults`. Both lists are `Optional+Computed`: omitting one means Terraform does not manage it, so guardrails attached elsewhere (e.g. through the Portkey UI) are preserved and adopted into state, while setting a list to `[]` always clears it (same semantics as `portkey_workspace_defaults`, see resource docs). Reads return both guardrail IDs and slugs and the provider stores slugs, so reference guardrails via `portkey_guardrail.foo.slug` to avoid a permanent plan diff. Destroying the resource clears both lists; there is no separate delete endpoint. Requires the `organisation_settings.read` / `organisation_settings.update` scopes on the Admin API key.
- **Organisation-scoped `portkey_guardrail` support** — `workspace_id` is now optional. Omitting it creates an organisation-scoped guardrail, which is what `portkey_organisation_defaults` requires; the organisation is taken from the Admin API key (the `/guardrails` create controller never reads `organisation_id` from the body, so scope is decided purely by the presence of `workspace_id`). Existing configurations are unaffected, since setting `workspace_id` keeps the previous workspace-scoped behaviour and scope remains immutable (changing it forces replacement). Organisation-scoped guardrails need the `organisation_guardrails.create` / `.read` / `.update` / `.delete` scopes on the Admin API key, which are separate from the workspace-scoped `guardrails.*` family.

### Changed
- **`portkey_workspace_defaults` no longer clears a guardrail list when the attribute is removed** — `input_guardrails` and `output_guardrails` are now `Optional+Computed`, and an omitted attribute uniformly means "Terraform does not manage this list" on update as well as on create. Previously, removing an attribute you had been managing was treated as an explicit removal and cleared those guardrails. To clear a list, set it to `[]` explicitly. This is what makes it possible to manage only one of the two lists, and to adopt the resource against a workspace whose guardrails were attached through the Portkey UI.

### Fixed
- **"Provider produced inconsistent result after apply" on `portkey_workspace_defaults` and `portkey_organisation_defaults`** — omitting one guardrail list while the workspace or organisation already had guardrails of that kind attached failed the apply outright. The provider preserved those guardrails (correctly) and wrote them into state, but because the attribute was `Optional` without `Computed`, Terraform had planned `null` for it and rejected the mismatch. This broke the documented adoption path — pointing the resource at a workspace or organisation whose defaults were set through the Portkey UI — and was most likely to bite at organisation level, where defaults are far more likely to be pre-populated than in a fresh workspace. Both lists are now `Optional+Computed`, so state legitimately adopts what the API reports. The same error also hit a second path: dropping an attribute that had previously been set to an explicit `[]`. Under `Optional+Computed` an omitted attribute plans to the prior state (`[]`), but the apply wrote the API's value straight through and the API reports an empty list as null, so the applied value contradicted the plan. Empty and null are now treated as equivalent on apply, matching how refresh has always reconciled them.
- **`portkey_workspace_defaults` silently cleared the list it was not managing** — the workspace `PUT` replaces the whole `defaults` object, so a request carrying only `input_guardrails` also cleared `output_guardrails`, and vice versa. Omitting an attribute destroyed those guardrails instead of preserving them, contrary to the documented behaviour. The provider now reads the workspace's current defaults and sends the unmanaged list back unchanged — the same overlay pattern `portkey_workspace_security_settings` uses for its partial writes. `portkey_organisation_defaults` needs no such merge: `PUT /v2/admin/organisation/defaults` preserves fields absent from the request body.
- **`portkey_guardrail` empty `workspace_id` in state** — an organisation-scoped guardrail (no workspace) is returned by the API with an empty `workspace_id`, which the provider stored as `""` instead of null and would surface as a permanent plan diff against a config that omits the attribute. The API's empty string now maps back to null.

## [0.2.31] - 2026-08-04

### Fixed
- **Departed-user 403s no longer wedge `plan`/`apply`** - `portkey_api_key` and `portkey_workspace_member` Read now treat 403/404 responses as missing-resource via the shared `client.IsNotFound` helper, matching `portkey_workspace` and `portkey_workspace_defaults`. When a user is removed from the organisation, the Admin API returns `403 AB03` (not 404) for reads of that user's workspace API keys and memberships; `portkey_api_key` only removed state on a literal `"404"` string match and `portkey_workspace_member` had no not-found handling at all, so a single departed user permanently broke `plan`/`apply` for any state containing their resources until every affected address was manually `terraform state rm`'d. This was especially painful for `for_each` over `data.portkey_users`. Replacing the string match also removes a false-positive path where any error body merely containing `"404"` (e.g. a request ID) was treated as not-found.

## [0.2.30] - 2026-08-04

### Added
- **Custom usage-limit reset intervals** — `portkey_usage_limits_policy` now supports `periodic_reset_days` (1–365) for reset cadences the fixed `periodic_reset` enum can't express, and exposes the API-computed `next_usage_reset_at` timestamp. `periodic_reset` and `periodic_reset_days` are mutually exclusive; the provider now rejects the combination at plan time rather than surfacing the API's `AB01 "Cannot set both periodic_reset and periodic_reset_days"` at apply time. Both new attributes are also exposed on the `portkey_usage_limits_policy` and `portkey_usage_limits_policies` data sources. Changing `periodic_reset_days` forces replacement, matching the existing `periodic_reset` behavior.
- **`excludes` in policy conditions** — conditions on `portkey_usage_limits_policy` and `portkey_rate_limits_policy` now accept an optional `excludes` key (string or array of strings) alongside `key`/`value`, so a policy can target a broad set while carving out exceptions (e.g. all `gpt-4*` models except the `-mini` variants).

### Changed
- **`periodic_reset` is now validated against allowed values** — `portkey_usage_limits_policy` rejects anything other than `monthly` or `weekly` at plan time. The Portkey API already returned `400` for other values, so this converts an apply-time failure into a plan-time one; no previously-working configuration is affected.
- **Policy `conditions` / `group_by` use `jsontypes.Normalized`** — these attributes are now semantically compared as JSON rather than as raw strings, so key ordering and whitespace differences between the config and the API response no longer surface as permanent plan diffs.

## [0.2.29] - 2026-07-07

### Added
- **`portkey_workspace_defaults` resource** — manage default `input_guardrails` / `output_guardrails` for a Portkey workspace via the `defaults` object on `PUT /admin/workspaces/{id}`. Kept separate from `portkey_workspace` because the Admin API requires guardrails to live in the target workspace, which would otherwise create an unresolvable Terraform DAG cycle (`workspace` → `guardrail` → workspace defaults). Setting a list to `[]` always clears it; omitting an attribute preserves existing guardrails on create but clears them on update (see resource docs for the full semantics). The API returns guardrail slugs on read under admin-API-key auth, so reference guardrails via `portkey_guardrail.foo.slug` in HCL to avoid a permanent plan diff. Supports import by `workspace_id`. Destroying the resource clears both lists; there is no separate delete endpoint for workspace defaults.
- **`portkey_workspace_security_settings` resource** — manage the per-workspace role-permission flags (`security_settings`) exposed by `PUT /admin/workspaces/{id}`. Every flag is `Optional+Computed`: omit a flag to keep its current API value. Because the Portkey API requires the full object on every PUT, the provider reads the current settings and overlays user-specified values before writing, so partial configs never clobber untouched flags. Supports import by `workspace_id`. Note: a workspace can only override a section (e.g. `logs`, `data_visibility`) when the organization has enabled workspace-level override for that section; otherwise the API returns `AB01 "Workspace override is not enabled"`. Destroying the resource removes it from Terraform state only and leaves the underlying API values untouched.

### Changed
- **Workspace Cascade-Delete on Destroy** - `portkey_workspace` now defaults `force_delete` to `true`, so `terraform destroy` cascades through all dependent resources (prompts, prompt partials, configs, guardrails, virtual keys) automatically — including those created outside Terraform. Previously, workspaces with any dependents would fail with `409 AB07`. Set `force_delete = false` to preserve the old behavior where the API rejects deletion when dependents exist.

### Fixed
- **SCIM Workspace Mappings Pagination** - Fixed `ListScimWorkspaceMappings` to paginate through all results instead of returning only the first page (100 items). Organizations with more than 100 SCIM workspace mappings would see `terraform import` fail with "Cannot import non-existent remote object" for mappings beyond the first page, and the `portkey_scim_workspace_mappings` data source would return incomplete results.
- **Workspace Deleted Out-of-Band State Reconciliation** - `portkey_workspace` Read now treats 403/404 responses as missing-resource (instead of a hard error), allowing Terraform to reconcile state when a workspace is deleted outside Terraform (e.g., via the Portkey UI). Previously, deleting a workspace out-of-band caused every subsequent `terraform plan` to fail.

## [0.2.28] - 2026-06-24

### Changed
- **Enhanced `portkey_integration` documentation** - Expanded the `configurations` field documentation and resource guide to include comprehensive examples for Azure OpenAI (Entra Federated, Workload Identity), Azure AI Foundry (Default, Entra, Entra Federated, Managed Identity, Workload Identity), and Google Vertex AI (Workload Identity) authentication modes. Added 8 new acceptance tests to validate these configuration patterns.

## [0.2.27] - 2026-06-10

### Added
- **`portkey_scim_workspace_mapping` resource** and
  **`portkey_scim_workspace_mappings` data source** — manage SCIM-group →
  workspace + role bindings via the Portkey SCIM Workspace Mappings Admin
  API. Supports binding either by `scim_group_name` (pre-create before
  the identity provider pushes the group) or by `scim_group_id` (bind to
  an existing SCIM group). Roles: `admin`, `member`, `manager`. The
  Portkey API has no update endpoint for SCIM mappings, so changing any
  field forces resource replacement (RequiresReplace). SCIM endpoints
  live under `/v1/scim/*`, alongside the rest of the Admin API, so the
  default `base_url` (`.../v1`) works for both SaaS and self-hosted
  callers without reconfiguration.

## [0.2.26] - 2026-05-22

### Fixed
- **Workspace Usage/Rate Limits Update Consistency** - Fixed "Provider produced inconsistent result after apply" errors when updating `usage_limits` or `rate_limits` on `portkey_workspace`. The Portkey API has eventual consistency and PUT responses may return stale data. The Update handler now trusts plan values (mirroring Create behavior) and reconciles on the next Read.

## [0.2.25] - 2026-05-22

### Added
- **Workspace Icon Support** - `portkey_workspace` now supports an `icon` attribute to manage workspace emoji icons:
  - Eliminates permanent plan drift caused by the Portkey API prepending icons to workspace names
  - When `icon` is set, the provider strips the icon prefix from API responses so `name` always reflects the clean user-configured value
  - Fully backwards compatible — existing configs without `icon` see no changes
  - Set `icon = ""` to explicitly clear an icon
  - The `portkey_workspaces` data source now includes the `icon` field
- **HTTP client retry on transient failures** - API requests that fail due
  to network errors or transient 5xx responses are now retried automatically
  with exponential backoff (500ms base, 5s cap, up to 4 retries by default).
  4xx responses (except 429) are returned immediately. At scale, a single
  transient 503 from the control plane would previously fail an entire
  `terraform plan` or `apply`; this change makes the provider resilient to
  ordinary upstream flakiness. Uses `hashicorp/go-retryablehttp`.
- **`max_retries` provider attribute** - Tunes the retry count to match
  deployment needs (e.g., reduce in fast-fail CI environments, increase
  for slow networks). Falls back to the `PORTKEY_MAX_RETRIES` environment
  variable; defaults to 4. Set to 0 to disable retries entirely. Follows
  the same pattern as `hashicorp/terraform-provider-vault` (`max_retries`
  + `VAULT_MAX_RETRIES`).
- **Retry visibility under `TF_LOG=DEBUG`** - Retry attempts are now
  logged via `terraform-plugin-log` at Debug level using the per-request
  context, so they appear with the correct Terraform operation tagging.
  By default (no `TF_LOG`), retries remain silent.
- **`client.NewClientWithConfig`** - New constructor accepting a
  `client.ClientConfig` struct for callers that need to set retry
  behavior. The existing `client.NewClient(baseURL, apiKey)` is retained
  as a thin wrapper for backwards compatibility.

## [0.2.17] - 2026-04-10

### Added
- **API Key Config Binding** - `portkey_api_key` now supports binding a default Portkey config:
  - `config_id` - ID of the Portkey config to bind as the default for all requests made using this API key
  - `allow_config_override` - Controls whether callers can override the bound config at request time (defaults to `false`)
  - Plan-time validation prevents setting `allow_config_override = true` without a `config_id`
  - Once set, clearing these fields in Terraform preserves the existing binding (use API directly to unset)

## [0.2.16] - 2026-03-13

### Fixed
- **Integration Workspace Access UUID/Slug Resolution** - Fixed `portkey_integration_workspace_access` read-after-create failures when `workspace_id` is a UUID but integration workspace APIs return workspace slugs. `GetIntegrationWorkspace()` now resolves UUID input to slug before lookup.

## [0.2.15] - 2026-03-06

### Added
- **Integration Model Access Control** - `portkey_integration` now supports `allow_all_models` attribute:
  - Defaults to `true` (all models available, matching API behavior)
  - Set to `false` to restrict access to only models explicitly enabled via `portkey_integration_model_access` resources

## [0.2.14] - 2026-02-27

### Added
- **Workspace-Scoped Integrations** - `portkey_integration` now supports workspace-level scoping:
  - `workspace_id` - Optional attribute to scope an integration to a specific workspace
  - `type` - Computed attribute showing "organisation" or "workspace" level
  - Organisation-level integrations are accessible across all workspaces
  - Workspace-level integrations are only accessible within the specified workspace

## [0.2.13] - 2026-02-22

### Fixed
- **Prompt & Prompt Partial Drift Detection** - Fixed issues where external changes made in the Portkey console were invisible to Terraform:
  - Now detects when someone edits a prompt or partial outside Terraform (new versions or rollbacks)
  - Terraform will show the drift and overwrite back to config values on next apply
- **Version Description Perpetual Drift** - Fixed infinite plan loop when `version_description` was set via console but not in Terraform config
- **MakeDefault Version Lookup** - Fixed incorrect version targeting when versions are created outside Terraform (e.g., console edits creating gaps in version sequence). Now uses version list API to find the correct version number instead of assuming `+1` increment

## [0.2.12] - 2026-02-20

### Added
- **Prompt Partials** - New resource and data sources for reusable template fragments:
  - `portkey_prompt_partial` - Create and manage prompt partials with versioning support
  - `portkey_prompt_partial` (data source) - Look up a single partial by slug with optional version
  - `portkey_prompt_partials` (data source) - List all partials, optionally filtered by workspace
  - Partials can be referenced in prompts via Mustache syntax (`{{>partial-slug}}`)
- **Usage & Rate Limits** - Added budget controls and rate limiting to workspace and API key resources:
  - `portkey_workspace` - New `usage_limits` and `rate_limits` attributes for workspace-level controls
  - `portkey_api_key` - New `usage_limits` and `rate_limits` attributes for key-level controls
  - All related data sources now return limit configurations

### Fixed
- **Prompt Resource Eventual Consistency** - Fixed issues where Portkey API returns stale data after mutations:
  - Template, version, and version_id are now preserved from state during Read
  - Added `virtual_key` to all version-creating prompt updates (API requirement)
  - Fixed Parameters field handling for Optional+Computed attributes
- **Version Description Warning** - Added diagnostic warning when `version_description` changes without content (no-op against API)

### Changed
- Refactored shared limit conversion helpers into `limits_helpers.go` to eliminate duplication

## [0.2.11] - 2026-02-04

### Added
- **Prompt Collections** - New resource and data sources for organizing prompts within workspaces:
  - `portkey_prompt_collection` - Create and manage prompt collections with hierarchical nesting support
  - `portkey_prompt_collection` (data source) - Look up a single collection by ID
  - `portkey_prompt_collections` (data source) - List all collections, optionally filtered by workspace
  - Support for nested collections via `parent_collection_id`

### Documentation
- Added documentation and examples for prompt collection resource and data sources
- Updated RESOURCE_MATRIX.md with prompt collection support

## [0.2.10] - 2026-01-28

### Fixed
- **Azure OpenAI Integration Configuration** - Fixed incorrect field names in documentation that caused 400 "Invalid request" errors:
  - Changed `resource_name` to `azure_resource_name`
  - Changed flat `deployment_id`/`api_version` to nested `azure_deployment_config` array structure
  - Added required `azure_auth_mode` field (values: "default", "entra", "managed")
  - Added required `azure_model_slug` field in deployment config

### Added
- **Azure OpenAI Authentication Modes** - Full documentation and examples for all 3 authentication methods:
  - `default` - API key authentication
  - `entra` - Microsoft Entra ID (Azure AD) authentication
  - `managed` - Azure Managed Identity authentication
- **Azure OpenAI Acceptance Tests** - 5 new tests to prevent future regressions:
  - Basic configuration with default auth
  - Multiple deployments
  - Entra ID authentication
  - Managed Identity authentication
  - Configuration updates

## [0.2.9] - 2026-01-27

### Added
- **Integration Model Access** - New resource and data source for managing model access per integration:
  - `portkey_integration_model_access` - Enable/disable specific models for an integration with optional custom pricing
  - `portkey_integration_models` - Data source to list all models available for an integration
  - Support for custom/fine-tuned models (`is_custom`, `is_finetune`, `base_model_slug`)
  - Support for custom pricing configuration (`pricing_config` with `pay_as_you_go` token prices)
  - Built-in models are disabled on delete; custom models are fully removed

### Documentation
- Added documentation for integration model access resource and data source

## [0.2.8] - 2026-01-26

### Added
- **Write-Only API Key Support** - `portkey_integration` now supports write-only API keys using Terraform 1.11+'s `WriteOnly` attribute:
  - `key_wo` - API key that is never stored in Terraform state or shown in plan output
  - `key_version` - Trigger to control when the key is sent to the API (increment to update)
  - Provides enhanced security for teams with strict compliance requirements
  - Original `key` attribute remains available for simpler workflows
- **Integration Workspace Access** - New resource and data source for managing integration access per workspace:
  - `portkey_integration_workspace_access` - Enable integrations for specific workspaces with optional limits
  - `portkey_integration_workspaces` - Data source to list workspace access configurations
  - Support for `usage_limits` (cost/request limits with alerts and periodic reset)
  - Support for `rate_limits` (requests per minute/hour/day)
  - Enables full IaC workflows without manual UI enablement

### Documentation
- Updated `portkey_integration` documentation with write-only key examples
- Added documentation for integration workspace access resource and data source

## [0.2.7] - 2026-01-08

### Added
- **API Key Metadata & Alert Emails** - `portkey_api_key` now supports:
  - `metadata` - Custom metadata (map of strings) attached to the API key for tracking, observability, and service identification. Example: `{"_user": "service-name", "service_uuid": "abc123"}`
  - `alert_emails` - List of email addresses to receive alerts related to the API key's usage
- **Workspace Metadata** - `portkey_workspace` now supports:
  - `metadata` - Custom metadata (map of strings) attached to the workspace for tracking teams, environments, and services

### Documentation
- Updated documentation for `portkey_api_key` resource and data sources
- Updated documentation for `portkey_workspace` resource and data sources

## [0.2.6] - 2026-01-05

### Documentation
- Added Terraform Registry documentation for all resources and data sources
- Documentation auto-generated using `tfplugindocs`

## [0.2.5] - 2026-01-05

### Added
- **AWS Bedrock IAM Role Support** - `portkey_integration` now supports a `configurations` field for provider-specific settings:
  - AWS Bedrock with IAM Role authentication (`aws_role_arn`, `aws_region`, `aws_external_id`)
  - AWS Bedrock with Access Keys (`aws_access_key_id`, `aws_region`)
  - Azure OpenAI configurations (`resource_name`, `deployment_id`, `api_version`)

### Documentation
- Added comprehensive examples for AWS Bedrock and Azure OpenAI integrations
- Updated `portkey_integration` documentation with `configurations` field

## [0.2.4] - 2026-01-05

### Fixed
- Fixed lint errors (gofmt, unused function, errcheck)
- **Critical: Fixed "Provider produced inconsistent result after apply" errors** - Resolved issues where Terraform would report inconsistent results due to state handling

## [0.2.3] - 2026-01-04

### Documentation
- Added known issue for workspace deletion with emoji names in README

## [0.2.2] - 2026-01-04

### Fixed
- **Critical: Resources no longer unnecessarily recreated on every apply** - Fixed a bug where `RequiresReplace` attributes (like `workspace_id`) were being overwritten during `Read()` operations, causing Terraform to detect false changes and trigger destroy/create cycles. Affected resources:
  - `portkey_config`
  - `portkey_guardrail`
  - `portkey_provider`
  - `portkey_prompt`
  - `portkey_integration`
  - `portkey_api_key`
  - `portkey_user_invite`
  - `portkey_rate_limits_policy`
  - `portkey_usage_limits_policy`
- Fixed CI linting issues and code formatting
- Reverted golangci-lint config to v1 format for CI compatibility

## [0.2.1] - 2026-01-03

### Documentation
- Added Prerequisites section to README
- Added Troubleshooting section to README
- Added Known Issues section to README
- Fixed README examples to use `jsonencode()` for JSON fields

### Fixed
- Fixed provider unit tests with correct resource counts
- Added Terraform setup to CI and formatted example files
- Fixed gofmt formatting and removed unused functions

## [0.2.0] - 2026-01-02

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

### Documentation
- Added guide for adding new APIs to the Terraform provider
- Added Registry and CI badges to README

## [0.1.0] - 2026-01-01

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

[Unreleased]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.32...HEAD
[0.2.32]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.31...v0.2.32
[0.2.31]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.30...v0.2.31
[0.2.30]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.29...v0.2.30
[0.2.29]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.28...v0.2.29
[0.2.28]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.27...v0.2.28
[0.2.27]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.26...v0.2.27
[0.2.26]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.25...v0.2.26
[0.2.25]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.24...v0.2.25
[0.2.17]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.16...v0.2.17
[0.2.16]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.15...v0.2.16
[0.2.15]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.14...v0.2.15
[0.2.14]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.13...v0.2.14
[0.2.13]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.12...v0.2.13
[0.2.12]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.11...v0.2.12
[0.2.11]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.10...v0.2.11
[0.2.10]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.9...v0.2.10
[0.2.9]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.8...v0.2.9
[0.2.8]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.7...v0.2.8
[0.2.7]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Portkey-AI/terraform-provider-portkey/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Portkey-AI/terraform-provider-portkey/releases/tag/v0.1.0

