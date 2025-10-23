# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

The Portkey team takes security issues seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by emailing:

**security@portkey.ai**

### What to Include

Please include the following information in your report:

- Type of vulnerability
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability, including how an attacker might exploit it

### Response Timeline

- **Initial Response**: Within 48 hours of report
- **Status Update**: Within 7 days with our assessment
- **Fix Timeline**: Depends on severity and complexity

### What to Expect

1. **Acknowledgment**: We'll confirm receipt of your vulnerability report
2. **Investigation**: We'll investigate and validate the issue
3. **Fix Development**: We'll develop and test a fix
4. **Disclosure**: We'll coordinate disclosure timing with you
5. **Credit**: We'll acknowledge your contribution (unless you prefer to remain anonymous)

## Security Best Practices

When using this provider:

### API Key Security

- **Never commit API keys** to version control
- **Use environment variables** or secret management systems
- **Rotate keys regularly**
- **Use the minimum required scopes** for API keys
- **Store Terraform state securely** with encryption and access controls

### State File Security

Terraform state files may contain sensitive information:

- Use **encrypted remote state** backends (S3 with encryption, Terraform Cloud, etc.)
- Enable **state locking** to prevent concurrent modifications
- Restrict **access to state files** to authorized users only
- Consider using **separate state files** for different environments

### Network Security

- Use **private networks** when possible for Terraform operations
- Consider **IP allowlisting** for API access if supported
- Use **VPNs or bastion hosts** for accessing production infrastructure

### Principle of Least Privilege

- Grant users **minimum required permissions**
- Use **workspace-specific roles** instead of organization-wide admin access
- Regularly **audit user permissions** and remove unused access
- Use **separate API keys** for different environments

## Known Security Considerations

### Sensitive Attributes

The following attributes are marked as sensitive in the provider:
- `api_key` in provider configuration
- User invitation details may contain email addresses

### API Authentication

The provider uses the Portkey Admin API which requires Organization-level API keys. These keys have broad permissions and should be handled with care.

## Updates and Patches

Security updates will be released as needed. Subscribe to:
- [GitHub Security Advisories](https://github.com/portkey-ai/terraform/security/advisories)
- [GitHub Releases](https://github.com/portkey-ai/terraform/releases)

## Disclosure Policy

When a security issue is fixed:

1. A security advisory will be published on GitHub
2. A new version will be released with the fix
3. The CHANGELOG will document the security fix
4. Users will be notified through GitHub releases

## Questions?

For questions about security that are not vulnerabilities, please open a [GitHub Discussion](https://github.com/portkey-ai/terraform/discussions).

---

**Thank you for helping keep Portkey Terraform Provider and our users safe!**

