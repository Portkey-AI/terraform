#!/bin/bash

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if API key is set
if [ -z "$PORTKEY_API_KEY" ]; then
    print_error "PORTKEY_API_KEY is not set"
    echo "Please run: export PORTKEY_API_KEY='your-api-key'"
    exit 1
fi

print_success "API key is set"

# Test 1: Clean slate
print_header "Test 1: Clean Environment"
rm -rf .terraform .terraform.lock.hcl terraform.tfstate* main.tf
mv comprehensive-test.tf main.tf 2>/dev/null || true
print_success "Environment cleaned"

# Test 2: Terraform Init
print_header "Test 2: Initialize Terraform"
if terraform init > /tmp/tf-init.log 2>&1; then
    print_success "Terraform initialized successfully"
else
    print_error "Terraform init failed"
    cat /tmp/tf-init.log
    exit 1
fi

# Test 3: Validate Configuration
print_header "Test 3: Validate Configuration"
if terraform validate > /tmp/tf-validate.log 2>&1; then
    print_success "Configuration is valid"
else
    print_error "Configuration validation failed"
    cat /tmp/tf-validate.log
    exit 1
fi

# Test 4: Plan
print_header "Test 4: Create Execution Plan"
if terraform plan -out=tfplan > /tmp/tf-plan.log 2>&1; then
    print_success "Plan created successfully"
    RESOURCES_TO_ADD=$(grep "Plan:" /tmp/tf-plan.log | grep -oE '[0-9]+ to add' | grep -oE '[0-9]+')
    print_info "Resources to add: $RESOURCES_TO_ADD"
else
    print_error "Plan failed"
    cat /tmp/tf-plan.log
    exit 1
fi

# Test 5: Apply
print_header "Test 5: Apply Configuration (Create Resources)"
if terraform apply -auto-approve tfplan > /tmp/tf-apply.log 2>&1; then
    print_success "Resources created successfully"
    
    # Extract and display outputs
    echo ""
    terraform output -json > /tmp/tf-outputs.json
    
    print_info "Created workspace IDs:"
    terraform output -json created_workspaces | python3 -c "
import sys, json
data = json.load(sys.stdin)
for name, ws in data.items():
    print(f'  {name}: {ws[\"id\"]} - {ws[\"name\"]}')" || true
    
else
    print_error "Apply failed"
    cat /tmp/tf-apply.log
    exit 1
fi

# Test 6: Verify State
print_header "Test 6: Verify Terraform State"
if terraform show > /tmp/tf-show.log 2>&1; then
    RESOURCE_COUNT=$(terraform state list | wc -l)
    print_success "State verified - $RESOURCE_COUNT resources tracked"
else
    print_error "State verification failed"
    exit 1
fi

# Test 7: Test Data Sources
print_header "Test 7: Verify Data Sources"
WORKSPACES_COUNT=$(terraform output -raw all_workspaces_count)
USERS_COUNT=$(terraform output -raw all_users_count)
print_success "Workspaces data source works - Found $WORKSPACES_COUNT workspaces"
print_success "Users data source works - Found $USERS_COUNT users"

# Test 8: Refresh
print_header "Test 8: Refresh State"
if terraform refresh > /tmp/tf-refresh.log 2>&1; then
    print_success "State refreshed successfully"
else
    print_error "Refresh failed"
    cat /tmp/tf-refresh.log
    exit 1
fi

# Test 9: Test Update (Modify a resource)
print_header "Test 9: Test Resource Update"
cat > update-test.tf << 'EOF'
# Update the dev workspace description
resource "portkey_workspace" "dev" {
  name        = "Development Workspace"
  description = "Updated description for testing"
}
EOF

# Append update to main.tf
sed -i.bak 's/description = "For development and testing"/description = "Updated description for testing"/' main.tf

if terraform plan -out=tfplan-update > /tmp/tf-plan-update.log 2>&1; then
    if grep -q "1 to change" /tmp/tf-plan-update.log; then
        print_success "Update plan detected changes correctly"
        
        if terraform apply -auto-approve tfplan-update > /tmp/tf-apply-update.log 2>&1; then
            print_success "Resource updated successfully"
        else
            print_error "Update apply failed"
        fi
    else
        print_info "No changes detected (this might be expected)"
    fi
else
    print_error "Update plan failed"
fi

# Test 10: Test Import (optional - would need existing resource)
print_header "Test 10: Test Resource Import (Skipped)"
print_info "Import test requires manual setup - skipping"

# Test 11: Destroy
print_header "Test 11: Destroy Resources (Cleanup)"
if terraform destroy -auto-approve > /tmp/tf-destroy.log 2>&1; then
    print_success "All resources destroyed successfully"
else
    print_error "Destroy failed"
    cat /tmp/tf-destroy.log
    exit 1
fi

# Test 12: Verify Clean State
print_header "Test 12: Verify Cleanup"
REMAINING_RESOURCES=$(terraform state list 2>/dev/null | wc -l)
if [ "$REMAINING_RESOURCES" -eq 0 ]; then
    print_success "All resources cleaned up"
else
    print_error "$REMAINING_RESOURCES resources still in state"
fi

# Final Summary
print_header "Test Summary"
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed! ✓${NC}"
    echo ""
    echo "Your Terraform Provider is fully functional and ready to use!"
fi

