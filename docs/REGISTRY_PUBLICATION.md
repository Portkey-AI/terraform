# Publishing to Terraform Registry

This guide walks you through the process of publishing the Portkey Terraform provider to the official Terraform Registry.

## Prerequisites

Before you begin, ensure you have:

- Admin access to the `portkey-ai` GitHub organization
- A public GitHub repository
- Go 1.21+ installed
- GPG installed on your system

## ⚠️ IMPORTANT: Repository Naming Requirement

✅ **Your repository is correctly named: `terraform-provider-portkey`**

The Terraform Registry requires provider repositories to follow the naming convention: `terraform-provider-{NAME}`, and your repository at [https://github.com/Portkey-AI/terraform-provider-portkey](https://github.com/Portkey-AI/terraform-provider-portkey) follows this convention perfectly!

You're ready to proceed with publication to the official Terraform Registry.

## Step 1: Set Up GPG Key

The Terraform Registry requires GPG-signed releases for security.

### Generate a GPG Key

```bash
# Generate a new GPG key (choose RSA and RSA, 4096 bits, no expiration)
gpg --full-generate-key

# Follow the prompts:
# - Select: (1) RSA and RSA
# - Key size: 4096
# - Expiration: 0 (does not expire)
# - Real name: Your name or "Portkey AI"
# - Email: Your email (must match your GitHub verified email)
```

### Export Your Keys

```bash
# List your keys and note the key ID (the long hex string)
gpg --list-secret-keys --keyid-format=long

# Export your public key (for Terraform Registry)
gpg --armor --export YOUR_EMAIL > portkey-gpg-public.asc

# Export your private key (for GitHub Secrets)
gpg --armor --export-secret-keys YOUR_KEY_ID > portkey-gpg-private.asc

# Get your GPG fingerprint (40-character hex string)
gpg --fingerprint YOUR_EMAIL
```

**Important**: Keep your private key (`portkey-gpg-private.asc`) secure and never commit it to version control!

## Step 2: Configure GitHub Repository

### Add GitHub Secrets

Navigate to your repository on GitHub:
`https://github.com/Portkey-AI/terraform-provider-portkey/settings/secrets/actions`

Add the following secrets:

1. **`GPG_PRIVATE_KEY`**
   - Content: Full contents of `portkey-gpg-private.asc`
   - Include the `-----BEGIN PGP PRIVATE KEY BLOCK-----` header and footer

2. **`PASSPHRASE`**
   - Content: Your GPG key passphrase
   - If you didn't set a passphrase, you can leave this empty or use an empty string

3. **`GPG_FINGERPRINT`**
   - Content: Your 40-character GPG fingerprint (no spaces)
   - Example: `1234567890ABCDEF1234567890ABCDEF12345678`

### Verify GitHub Actions Workflow

The `.github/workflows/release.yml` file has been created. This workflow:
- Triggers on version tags (e.g., `v0.1.0`)
- Builds binaries for multiple platforms
- Signs the release with your GPG key
- Creates a GitHub release with all artifacts

## Step 3: Create Your First Release

### Prepare the Release

1. Ensure all changes are committed:
```bash
git status
git add .
git commit -m "Prepare v0.1.0 release"
```

2. Create and push a version tag:
```bash
# Create an annotated tag
git tag -a v0.1.0 -m "Initial release v0.1.0"

# Push the tag to trigger the release workflow
git push origin v0.1.0
```

3. Monitor the GitHub Action:
   - Go to: `https://github.com/Portkey-AI/terraform-provider-portkey/actions`
   - Watch the "Release" workflow run
   - Verify it completes successfully

### Verify the Release

After the workflow completes:
1. Go to: `https://github.com/Portkey-AI/terraform-provider-portkey/releases`
2. Verify the release contains:
   - Binaries for multiple platforms (Linux, macOS, Windows)
   - `terraform-provider-portkey_X.X.X_SHA256SUMS` file
   - `terraform-provider-portkey_X.X.X_SHA256SUMS.sig` signature file

## Step 5: Register on Terraform Registry

### Sign In to Terraform Registry

1. Go to: https://registry.terraform.io/
2. Click "Sign In" in the top right
3. Authenticate with your GitHub account (must have access to `portkey-ai` organization)

### Publish the Provider

1. Click "Publish" in the top navigation
2. Select "Provider"
3. Choose your GitHub repository: `Portkey-AI/terraform-provider-portkey`

### Add Your GPG Public Key

1. Click on your profile/settings in the registry
2. Navigate to "Signing Keys"
3. Click "Add a key"
4. Paste the contents of `portkey-gpg-public.asc`
5. Submit the key

### Complete the Publication

1. The registry will:
   - Verify your repository structure
   - Check for valid releases
   - Verify GPG signatures
   - Parse your documentation
   
2. If everything is correct, your provider will be published!

## Step 6: Verify Installation

Test that users can install your provider:

```hcl
terraform {
  required_providers {
    portkey = {
      source  = "portkey-ai/portkey"
      version = "~> 0.1"
    }
  }
}

provider "portkey" {
  api_key = "your-api-key"
}
```

Run:
```bash
terraform init
```

The provider should download successfully from the registry!

## Publishing Future Releases

For subsequent releases:

1. Update `CHANGELOG.md` with changes
2. Commit your changes
3. Create and push a new version tag:
```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

The GitHub Action will automatically:
- Build and sign the release
- Publish to GitHub Releases
- The Terraform Registry will auto-detect and publish the new version

## Troubleshooting

### Release Workflow Fails

**GPG Import Error**:
- Verify `GPG_PRIVATE_KEY` secret contains the full private key
- Ensure the passphrase is correct
- Check that the key hasn't expired

**GoReleaser Error**:
- Verify `.goreleaser.yml` syntax
- Check Go module path matches repository name
- Ensure all dependencies are available

### Registry Verification Fails

**GPG Signature Invalid**:
- Verify the public key uploaded to the registry matches your private key
- Check that `GPG_FINGERPRINT` secret is correct
- Ensure release artifacts are properly signed

**Documentation Errors**:
- Run `go generate ./...` to regenerate documentation
- Ensure examples are valid Terraform code
- Check that documentation follows Terraform's format requirements

### Provider Not Found After Publication

**Cache Issues**:
- Wait a few minutes for CDN propagation
- Try `terraform init -upgrade`
- Clear Terraform's plugin cache: `rm -rf ~/.terraform.d/plugins`

**Namespace Issues**:
- Verify repository name is exactly `terraform-provider-portkey`
- Check that the provider address in `main.go` is correct:
  ```go
  Address: "registry.terraform.io/portkey-ai/portkey"
  ```

## Additional Resources

- [Terraform Registry Provider Publishing Documentation](https://www.terraform.io/docs/registry/providers/publishing.html)
- [GoReleaser Documentation](https://goreleaser.com/intro/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework)

## Security Best Practices

1. **Protect Your GPG Key**: Never commit your private key or passphrase
2. **Use GitHub Secrets**: Store all sensitive information in repository secrets
3. **Enable 2FA**: Use two-factor authentication on GitHub and Terraform Registry
4. **Regular Key Rotation**: Consider rotating GPG keys periodically
5. **Monitor Releases**: Review all automated releases to ensure they're legitimate

