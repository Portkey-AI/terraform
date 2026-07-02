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

# Setting input_guardrails = [] or output_guardrails = [] clears them on the
# next apply. Removing the attribute has the same effect. Destroying this
# resource clears both lists via the workspace update endpoint.
