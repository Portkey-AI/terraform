# Contributing to Portkey Terraform Provider

Thank you for your interest in contributing to the Portkey Terraform Provider! We welcome contributions from the community.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and collaborative environment.

## How to Contribute

### Reporting Issues

- **Search First**: Check if the issue already exists in the [issue tracker](https://github.com/portkey-ai/terraform/issues)
- **Provide Details**: Include as much information as possible:
  - Provider version
  - Terraform version
  - Go version (if building from source)
  - Steps to reproduce
  - Expected vs actual behavior
  - Relevant configuration files (sanitized)
  - Error messages and logs

### Suggesting Features

- Open an issue with the `enhancement` label
- Describe the use case and benefits
- Provide examples of how it would work

### Submitting Pull Requests

1. **Fork and Clone**
```bash
git clone https://github.com/YOUR-USERNAME/terraform
cd terraform
```

2. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Changes**
   - Write clear, maintainable code
   - Follow Go best practices
   - Add tests for new functionality
   - Update documentation

4. **Test Your Changes**
   ```bash
   # Run unit tests
   go test ./...
   
   # Build the provider
   make build
   
   # Install locally and test
   make install
   ```

5. **Commit and Push**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin feature/your-feature-name
   ```

6. **Open Pull Request**
   - Provide a clear description of changes
   - Reference any related issues
   - Ensure CI checks pass

## Development Setup

### Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Portkey Admin API key for testing

### Building from Source

```bash
# Clone the repository
git clone https://github.com/portkey-ai/terraform
cd terraform

# Install dependencies
go mod download

# Build the provider
make build

# Install locally
make install
```

### Running Tests

```bash
# Unit tests
go test ./...

# Acceptance tests (requires valid API key)
export PORTKEY_API_KEY="your-admin-api-key"
make testacc
```

## Code Style

### Go Code

- Follow standard Go formatting: `gofmt` and `go vet`
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and concise

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions or changes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

Examples:
```
feat: add virtual_key resource
fix: handle nil pointer in workspace update
docs: update installation instructions
```

## Project Structure

```
terraform/
├── internal/
│   ├── client/          # API client implementation
│   └── provider/        # Terraform resources and data sources
├── examples/            # Usage examples
├── docs/               # Documentation
└── main.go             # Provider entry point
```

### Adding a New Resource

1. Create `{resource_name}_resource.go` in `internal/provider/`
2. Implement the resource interface:
   - Schema definition
   - Create, Read, Update, Delete operations
   - Import functionality
3. Register the resource in `provider.go`
4. Add tests
5. Update documentation

### Adding a New Data Source

1. Create `{resource_name}_data_source.go` in `internal/provider/`
2. Implement the data source interface:
   - Schema definition
   - Read operation
3. Register the data source in `provider.go`
4. Add tests
5. Update documentation

## Testing Guidelines

### Unit Tests

- Test individual functions and methods
- Mock external dependencies
- Cover edge cases and error conditions

### Acceptance Tests

- Test full resource lifecycle (Create, Read, Update, Delete)
- Test import functionality
- Use unique resource names to avoid conflicts
- Clean up resources after tests

Example:
```go
func TestAccWorkspaceResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccWorkspaceConfig("test-name"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("portkey_workspace.test", "name", "test-name"),
                ),
            },
            // Import testing
            {
                ResourceName:      "portkey_workspace.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

## Documentation

- Update README.md for significant changes
- Add examples for new resources
- Document breaking changes
- Keep CHANGELOG.md updated

## Getting Help

- **Questions**: Open a [GitHub Discussion](https://github.com/portkey-ai/terraform/discussions)
- **Bugs**: File an [issue](https://github.com/portkey-ai/terraform/issues)
- **Chat**: Join the [Portkey Discord](https://portkey.sh/discord-1)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (Mozilla Public License 2.0).

## Thank You!

Your contributions help make this project better for everyone! 🙏

