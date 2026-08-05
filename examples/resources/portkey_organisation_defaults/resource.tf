# Organisation-scoped guardrail: omitting workspace_id creates the guardrail at
# the organisation level, which is required for organisation defaults.
resource "portkey_guardrail" "org_pii" {
  name = "org-pii-check"
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

resource "portkey_organisation_defaults" "this" {
  input_guardrails  = [portkey_guardrail.org_pii.slug]
  output_guardrails = [portkey_guardrail.org_pii.slug]
}

# There is no organisation_id argument: the Admin API derives the organisation
# from the configured API key, so exactly one portkey_organisation_defaults may
# exist per provider configuration.
#
# At least one of input_guardrails / output_guardrails must be set.
#
# Setting input_guardrails = [] or output_guardrails = [] ALWAYS clears them on
# the next apply.
#
# Omitting an attribute means Terraform does not manage that list. It is never
# sent to the API, so guardrails already set on the organisation (e.g. attached
# via the Portkey UI) are PRESERVED and adopted into state. This holds uniformly
# on create and on update: removing an attribute you previously managed leaves
# those guardrails in place. Set the attribute to [] to clear it.
#
# Destroying this resource clears both lists via the update endpoint.
