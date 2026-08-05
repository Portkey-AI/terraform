resource "portkey_workspace" "prod" {
  name        = "Production"
  description = "Prod workspace"
}

# Workspace-scoped guardrail: set workspace_id. Attach it with
# portkey_workspace_defaults.
resource "portkey_guardrail" "workspace_scoped" {
  name         = "workspace-word-count"
  workspace_id = portkey_workspace.prod.id
  checks = jsonencode([{
    id = "default.wordCount"
    parameters = {
      minWords = 1
      maxWords = 1000
    }
  }])
  actions = jsonencode({
    onFail  = "log"
    message = "Word count check failed"
  })
}

# Organisation-scoped guardrail: omit workspace_id. The Admin API derives the
# organisation from the configured API key. Only guardrails created this way can
# be attached with portkey_organisation_defaults.
resource "portkey_guardrail" "organisation_scoped" {
  name = "org-word-count"
  checks = jsonencode([{
    id = "default.wordCount"
    parameters = {
      minWords = 1
      maxWords = 1000
    }
  }])
  actions = jsonencode({
    onFail  = "log"
    message = "Word count check failed"
  })
}

# Scope is immutable: adding or removing workspace_id replaces the guardrail.
#
# Organisation-scoped guardrails need the organisation_guardrails.create /
# .read / .update / .delete scopes on the Admin API key, which are separate from
# the workspace-scoped guardrails.* family.
