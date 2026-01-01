default: install

# Generate documentation
generate:
	go generate ./...

# Run unit tests only (no API calls)
.PHONY: test
test:
	go test ./... -v $(TESTARGS) -timeout 10m

# Run acceptance tests (requires PORTKEY_API_KEY in .env or environment)
# Usage: make testacc
#        make testacc TESTARGS="-run TestAccConfig"
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run a quick smoke test of acceptance tests
# Tests provider configuration and basic data sources
.PHONY: smoketest
smoketest:
	TF_ACC=1 go test ./internal/provider/... -v -run "TestAccProvider_Configure|TestAccWorkspacesDataSource" -timeout 10m

# Run acceptance tests for a specific resource
# Usage: make testacc-resource RESOURCE=config
.PHONY: testacc-resource
testacc-resource:
	TF_ACC=1 go test ./internal/provider/... -v -run "TestAcc.*$(RESOURCE)" -timeout 30m

# Build the provider
build:
	go build -o terraform-provider-portkey

# Install the provider locally for testing
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/portkey-ai/portkey/0.1.0/$(shell go env GOOS)_$(shell go env GOARCH)
	mv terraform-provider-portkey ~/.terraform.d/plugins/registry.terraform.io/portkey-ai/portkey/0.1.0/$(shell go env GOOS)_$(shell go env GOARCH)/terraform-provider-portkey_v0.1.0

# Format code
fmt:
	gofmt -s -w -e .
	terraform fmt -recursive ./examples/

# Lint code
lint:
	golangci-lint run

# Run all checks (same as CI)
check: fmt lint
	go build ./...
	go test ./... -v -count=1

# Clean build artifacts
clean:
	rm -f terraform-provider-portkey
	rm -rf ~/.terraform.d/plugins/registry.terraform.io/portkey-ai/

# Create a sample .env file
.PHONY: env-sample
env-sample:
	@echo "Creating .env.example..."
	@echo "# Portkey API Key (required for acceptance tests)" > .env.example
	@echo "PORTKEY_API_KEY=your-portkey-api-key" >> .env.example
	@echo "" >> .env.example
	@echo "# Test workspace configuration (optional - uses defaults if not set)" >> .env.example
	@echo "# TEST_WORKSPACE_ID=your-workspace-uuid" >> .env.example
	@echo "# TEST_WORKSPACE_SLUG=your-workspace-slug" >> .env.example
	@echo "# TEST_INTEGRATION_ID=your-integration-id" >> .env.example
	@echo "# TEST_COLLECTION_ID=your-collection-id" >> .env.example
	@echo "" >> .env.example
	@echo "Created .env.example - copy to .env and fill in your values"

.PHONY: default generate test testacc smoketest testacc-resource build install fmt lint check clean env-sample
