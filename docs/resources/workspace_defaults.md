---
page_title: "portkey_workspace_defaults Resource - portkey"
subcategory: ""
description: |-
  Manages default input/output guardrails for a Portkey workspace.
---

# portkey_workspace_defaults (Resource)

Manages default input/output guardrails for a Portkey workspace. Every request routed through the workspace's API keys will pass through the configured guardrails in order.

This resource is separate from `portkey_workspace` because the Portkey Admin API requires guardrails to live in the target workspace they are attached to. Modeling guardrails as an attribute on `portkey_workspace` would create an unresolvable Terraform DAG cycle when the workspace, guardrails, and attachment are all managed in a single config. Exactly one `portkey_workspace_defaults` may exist per workspace.

## Example Usage

```terraform
resource "portkey_workspace" "prod" {
  name        = "Production"
  description = "Prod workspace"
}

resource "portkey_guardrail" "pii" {
  name         = "pii-check"
  workspace_id = portkey_workspace.prod.id
  checks = jsonencode([{
    id = "default.wordCount"
    parameters = {
      minWords = 1
      maxWords = 4000
    }
  }])
  actions = jsonencode({
    onFail  = "log"
    message = "guardrail triggered"
  })
}

resource "portkey_workspace_defaults" "prod" {
  workspace_id      = portkey_workspace.prod.id
  input_guardrails  = [portkey_guardrail.pii.slug]
  output_guardrails = [portkey_guardrail.pii.slug]
}
```

~> **Note:** Reference guardrails by `slug` for stable plans. The Admin API accepts either guardrail IDs or slugs on write but returns slugs on read (under admin-API-key auth), so state will always contain slugs. Storing `portkey_guardrail.foo.slug` keeps state and HCL in sync; using `portkey_guardrail.foo.id` (a UUID) produces a permanent plan diff.

~> **Note:** Setting either list to `[]` clears all attached guardrails of that kind. Removing the attribute from the config has the same effect. Destroying the resource clears both lists via the workspace update endpoint; there is no separate delete endpoint for workspace defaults.

## Schema

### Required

- `workspace_id` (String) ID or slug of the workspace to configure defaults for.

### Optional

- `input_guardrails` (List of String) Guardrails applied to inbound requests, as a list of guardrail slugs (or IDs — the API accepts both). The API returns slugs on read under admin-API-key auth, so prefer `portkey_guardrail.foo.slug` in HCL to avoid a permanent plan diff. Setting to `[]` clears all input guardrails.
- `output_guardrails` (List of String) Guardrails applied to model responses, as a list of guardrail slugs (or IDs — the API accepts both). The API returns slugs on read under admin-API-key auth, so prefer `portkey_guardrail.foo.slug` in HCL to avoid a permanent plan diff. Setting to `[]` clears all output guardrails.

### Read-Only

- `id` (String) Resource identifier. Equal to `workspace_id`.

## Import

Import by workspace ID or slug:

```
terraform import portkey_workspace_defaults.prod <workspace-id-or-slug>
```
