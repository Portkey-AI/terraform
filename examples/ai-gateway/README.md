# AI Gateway Example

This example demonstrates how to set up a complete Portkey AI Gateway infrastructure using Terraform.

## Resources Created

This configuration creates:

1. **Workspace** - A production workspace to organize all resources
2. **Integration** - An OpenAI API integration at the organization level
3. **Provider (Virtual Key)** - A workspace-scoped key linked to the OpenAI integration
4. **Config** - Gateway configuration with retry logic and caching
5. **Guardrail** - Content validation rules for safety
6. **Usage Limits Policy** - Monthly cost controls
7. **Rate Limits Policy** - Request throttling
8. **API Key** - A service key for backend applications

## Prerequisites

- Portkey account with Admin API access
- OpenAI API key
- Terraform >= 1.0

## Usage

1. Set your Portkey API key:

```bash
export PORTKEY_API_KEY="your-portkey-admin-api-key"
```

2. Create a `terraform.tfvars` file:

```hcl
openai_api_key = "sk-your-openai-api-key"
```

3. Initialize and apply:

```bash
terraform init
terraform plan
terraform apply
```

## Configuration Details

### Gateway Config

The gateway config includes:
- **Retry Logic**: 3 attempts on status codes 429, 500, 502, 503
- **Caching**: Simple cache mode enabled

### Guardrail

Content validation:
- Minimum 1 word
- Maximum 10,000 words
- Blocks requests that fail validation

### Usage Limits

- Type: Cost-based
- Limit: $1,000/month
- Alert threshold: $800 (80%)
- Resets monthly
- Grouped by API key

### Rate Limits

- Type: Requests per minute
- Limit: 100 RPM
- Grouped by API key

## Outputs

After applying, you'll get:
- `workspace_id` - Use this in your Portkey SDK configuration
- `config_slug` - Reference this config in API calls
- `api_key_id` - The API key for your backend service

## Clean Up

```bash
terraform destroy
```

## Related Examples

- [Basic Setup](../basic/) - Organization and user management
- [Multi-Environment](../multi-environment/) - Multiple environments with shared configs

