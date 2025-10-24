# ============================================================================
# WORKSPACE OUTPUTS
# ============================================================================

output "production_workspaces" {
  description = "Production workspace IDs"
  value = {
    api       = portkey_workspace.prod_api.id
    ml        = portkey_workspace.prod_ml.id
    analytics = portkey_workspace.prod_analytics.id
  }
}

output "staging_workspaces" {
  description = "Staging workspace IDs"
  value = {
    api = portkey_workspace.staging_api.id
    ml  = portkey_workspace.staging_ml.id
  }
}

output "development_workspace" {
  description = "Development workspace ID"
  value       = portkey_workspace.dev_shared.id
}

# ============================================================================
# USER INVITATION OUTPUTS
# ============================================================================

output "pending_invitations" {
  description = "Status of all user invitations"
  value = {
    engineering_lead  = portkey_user_invite.eng_lead.status
    ml_lead          = portkey_user_invite.ml_lead.status
    analyst          = portkey_user_invite.analyst.status
    backend_engineer = portkey_user_invite.backend_engineer.status
  }
}

output "invitation_details" {
  description = "Details of all invitations"
  value = {
    total_invitations = 4
    invitations = [
      {
        email  = portkey_user_invite.eng_lead.email
        role   = portkey_user_invite.eng_lead.role
        status = portkey_user_invite.eng_lead.status
      },
      {
        email  = portkey_user_invite.ml_lead.email
        role   = portkey_user_invite.ml_lead.role
        status = portkey_user_invite.ml_lead.status
      },
      {
        email  = portkey_user_invite.analyst.email
        role   = portkey_user_invite.analyst.role
        status = portkey_user_invite.analyst.status
      },
      {
        email  = portkey_user_invite.backend_engineer.email
        role   = portkey_user_invite.backend_engineer.role
        status = portkey_user_invite.backend_engineer.status
      }
    ]
  }
  sensitive = true
}

# ============================================================================
# ORGANIZATION SUMMARY
# ============================================================================

output "organization_summary" {
  description = "Summary of Portkey organization structure"
  value = {
    total_workspaces = length(data.portkey_workspaces.all.workspaces)
    workspace_breakdown = {
      production  = 3
      staging     = 2
      development = 1
    }
    environment_access = local.workspace_map
  }
}

output "workspace_urls" {
  description = "Portkey dashboard URLs for each workspace"
  value = {
    prod_api       = "https://app.portkey.ai/workspaces/${portkey_workspace.prod_api.id}"
    prod_ml        = "https://app.portkey.ai/workspaces/${portkey_workspace.prod_ml.id}"
    prod_analytics = "https://app.portkey.ai/workspaces/${portkey_workspace.prod_analytics.id}"
    staging_api    = "https://app.portkey.ai/workspaces/${portkey_workspace.staging_api.id}"
    staging_ml     = "https://app.portkey.ai/workspaces/${portkey_workspace.staging_ml.id}"
    dev_shared     = "https://app.portkey.ai/workspaces/${portkey_workspace.dev_shared.id}"
  }
}

