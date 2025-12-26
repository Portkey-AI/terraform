# Basic Example

This example demonstrates basic usage of the Portkey Terraform provider.

## What's Included

This example shows:
- Creating workspaces
- Inviting users with specific workspace access and permissions
- Adding existing users to workspaces
- Querying workspaces and users using data sources
- Outputting workspace information

## Usage

1. Set your Portkey Admin API key as an environment variable:
   ```bash
   export PORTKEY_API_KEY="your-admin-api-key"
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Review the planned changes:
   ```bash
   terraform plan
   ```

4. Apply the configuration:
   ```bash
   terraform apply
   ```

5. View the outputs:
   ```bash
   terraform output
   ```

## Customization

- Update the workspace names and descriptions in `main.tf`
- Change the email addresses for user invitations
- Modify the scopes and roles as needed for your organization
- Replace `"existing-user-id"` with an actual user ID from your organization

## Clean Up

To remove all resources created by this example:

```bash
terraform destroy
```

## Next Steps

- [AI Gateway Example](../ai-gateway/) - Set up integrations, providers, configs, guardrails, and policies
- [Multi-Environment Example](../multi-environment/) - Multiple environments with shared configurations

