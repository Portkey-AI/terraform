variable "portkey_api_key" {
  description = "Portkey Admin API Key"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "production"

  validation {
    condition     = contains(["production", "staging", "development"], var.environment)
    error_message = "Environment must be production, staging, or development."
  }
}

variable "team" {
  description = "Team name"
  type        = string
  default     = "platform"
}

variable "engineering_lead_email" {
  description = "Email for engineering team lead"
  type        = string
}

variable "ml_lead_email" {
  description = "Email for ML team lead"
  type        = string
}

variable "analyst_email" {
  description = "Email for analytics team member"
  type        = string
}

variable "backend_engineer_email" {
  description = "Email for backend engineer"
  type        = string
}

