# Basic workspace
resource "portkey_workspace" "basic" {
  name        = "Development"
  description = "Development workspace"
}

# Workspace with usage limits and rate limits
resource "portkey_workspace" "with_limits" {
  name        = "Production"
  description = "Production workspace with budget controls"

  usage_limits = [{
    type            = "cost"
    credit_limit    = 1000
    alert_threshold = 800
    periodic_reset  = "monthly"
  }]

  rate_limits = [{
    type  = "requests"
    unit  = "rpm"
    value = 5000
  }]
}

# Opt in to cascade deletion — destroy removes dependent resources (prompts,
# configs, virtual keys, etc.) first, including those created outside
# Terraform. Without this, the destroy fails with 409 AB07 while they exist.
resource "portkey_workspace" "disposable" {
  name         = "Scratch"
  description  = "Workspace whose contents may be destroyed with it"
  force_delete = true
}

# To clear limits, simply remove the usage_limits or rate_limits blocks
# from your config and re-apply.

# Default input/output guardrails for a workspace are configured via the
# separate portkey_workspace_defaults resource. See
# examples/resources/portkey_workspace_defaults/resource.tf.
