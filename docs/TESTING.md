# Testing the Portkey Terraform Provider

## Prerequisites

1. ✅ Provider is built and installed (`make install`)
2. A Portkey Admin API key (get from your Portkey dashboard)

## Quick Test Steps

### 1. Set Your API Key

```bash
export PORTKEY_API_KEY="0a7cZ7EDNCi/HLzwfJx40W1wEM+p"
```

### 2. Initialize Terraform

```bash
terraform init
```

This will:
- Initialize the working directory
- Download the provider from your local installation
- Set up the backend

### 3. Validate the Configuration

```bash
terraform validate
```

This checks if your configuration is syntactically valid.

### 4. Plan the Changes

```bash
terraform plan
```

This shows what Terraform will do without actually making changes. Review the output to see what resources will be created.

### 5. Apply the Changes (Optional)

⚠️ **Warning**: This will make real changes to your Portkey organization!

```bash
terraform apply
```

Type `yes` when prompted to create the resources.

### 6. Inspect State

```bash
terraform show
```

This shows the current state of your managed infrastructure.

### 7. Destroy Resources (Cleanup)

```bash
terraform destroy
```

Type `yes` to remove all created resources.

## Testing Individual Resources

### Test Workspace Creation

Use the included `test.tf` file:

```bash
terraform init
terraform plan
terraform apply
```

### Test Data Sources

Create a file to test reading data:

```hcl
data "portkey_workspaces" "all" {}

output "all_workspaces" {
  value = data.portkey_workspaces.all.workspaces
}
```

Then run:

```bash
terraform plan
```

## Debugging

### Enable Detailed Logging

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
terraform plan
```

### Test in Debug Mode

Run the provider in debug mode:

```bash
make build
./terraform-provider -debug
```

Then in another terminal, use the provided reattach configuration.

## Running Unit Tests

If you have Go tests:

```bash
go test ./... -v
```

## Running Acceptance Tests

⚠️ **Warning**: Acceptance tests make real API calls!

```bash
make testacc
```

Or with specific tests:

```bash
TF_ACC=1 go test ./... -v -run TestAccWorkspaceResource
```

## Common Issues

### Provider Not Found

If you get "provider not found" errors:

1. Check the installation path:
   ```bash
   ls -la ~/.terraform.d/plugins/registry.terraform.io/portkey-ai/portkey/0.1.0/
   ```

2. Reinstall:
   ```bash
   make clean
   make install
   ```

3. Delete Terraform cache:
   ```bash
   rm -rf .terraform .terraform.lock.hcl
   terraform init
   ```

### API Authentication Errors

- Verify your API key is set: `echo $PORTKEY_API_KEY`
- Ensure the key has admin privileges
- Check the base URL if using self-hosted Portkey

### Validation Errors

Run `terraform validate` to check for configuration syntax errors.

## Example Test Workflow

```bash
# 1. Set up environment
export PORTKEY_API_KEY="your-key"

# 2. Clean start
rm -rf .terraform .terraform.lock.hcl terraform.tfstate*

# 3. Initialize
terraform init

# 4. Test with test.tf
terraform plan

# 5. If everything looks good, apply
terraform apply -auto-approve

# 6. Verify
terraform show

# 7. Clean up
terraform destroy -auto-approve
```

## Additional Testing Resources

- Use `terraform fmt` to format your configuration files
- Use `terraform graph | dot -Tsvg > graph.svg` to visualize dependencies
- Check the `examples/` directory for more complex scenarios

