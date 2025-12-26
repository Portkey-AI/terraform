default: install

# Generate documentation
generate:
	go generate ./...

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

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

.PHONY: default generate testacc build install fmt lint check clean
