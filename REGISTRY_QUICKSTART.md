# Publishing to Terraform Registry - Quick Start

Your Terraform provider is now properly configured and ready for publication! 🎉

## ✅ Repository Setup Complete

- **Repository**: [https://github.com/Portkey-AI/terraform-provider-portkey](https://github.com/Portkey-AI/terraform-provider-portkey)
- **Module Path**: `github.com/portkey-ai/terraform-provider-portkey`
- **Registry Address**: `registry.terraform.io/portkey-ai/portkey`

All code references have been updated to match your renamed repository!

## Files Created

✅ `.goreleaser.yml` - Release automation configuration  
✅ `.github/workflows/release.yml` - GitHub Actions workflow  
✅ `docs/REGISTRY_PUBLICATION.md` - Complete step-by-step guide

## Next Steps to Publish

### 1. Set Up GPG Key (Required)

Generate a GPG key for signing releases:

```bash
# Generate GPG key (choose RSA 4096 bits, no expiration)
gpg --full-generate-key

# Export public key for Terraform Registry
gpg --armor --export your-email@example.com > portkey-gpg-public.asc

# Export private key for GitHub Secrets
gpg --armor --export-secret-keys YOUR_KEY_ID > portkey-gpg-private.asc

# Get fingerprint (40-character hex string)
gpg --fingerprint your-email@example.com
```

### 2. Add GitHub Secrets

Go to: [https://github.com/Portkey-AI/terraform-provider-portkey/settings/secrets/actions](https://github.com/Portkey-AI/terraform-provider-portkey/settings/secrets/actions)

Add these three secrets:

- **`GPG_PRIVATE_KEY`**: Contents of `portkey-gpg-private.asc`
- **`PASSPHRASE`**: Your GPG key passphrase
- **`GPG_FINGERPRINT`**: Your 40-character GPG fingerprint (no spaces)

### 3. Commit and Push Changes

```bash
cd /Users/ra/workspace/terraform-provider

# Review changes
git status

# Add all files
git add .

# Commit changes
git commit -m "Add Terraform Registry publishing configuration"

# Push to GitHub
git push
```

### 4. Create Your First Release

```bash
# Create and push a version tag
git tag -a v0.1.0 -m "Initial release v0.1.0"
git push origin v0.1.0
```

This will automatically trigger the GitHub Action to:
- Build binaries for Linux, macOS, and Windows
- Sign the release with your GPG key
- Create a GitHub release with all artifacts

Monitor the action at: [https://github.com/Portkey-AI/terraform-provider-portkey/actions](https://github.com/Portkey-AI/terraform-provider-portkey/actions)

### 5. Publish to Terraform Registry

1. Go to: [https://registry.terraform.io/](https://registry.terraform.io/)
2. Sign in with your GitHub account
3. Click "Publish" → "Provider"
4. Select your repository: `Portkey-AI/terraform-provider-portkey`
5. Add your GPG public key in Settings → Signing Keys
6. The registry will automatically verify and publish your provider!

## After Publication

Users will be able to install your provider with:

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
  api_key = "your-admin-api-key"
}
```

## Future Releases

For subsequent versions, simply create and push a new tag:

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

The GitHub Action will automatically build and publish the release!

## Documentation

For detailed instructions, troubleshooting, and best practices, see:
- **`docs/REGISTRY_PUBLICATION.md`** - Complete guide

## Resources

- [Terraform Registry Documentation](https://www.terraform.io/docs/registry/providers/publishing.html)
- [Provider Development Guide](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework)
- [Your Repository](https://github.com/Portkey-AI/terraform-provider-portkey)

---

**Ready to proceed?** Start with Step 1 (GPG Key Setup) above!

