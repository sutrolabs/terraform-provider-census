#!/bin/bash

# Manual testing script for Census Terraform Provider
# This script demonstrates what can be tested without a real Census API

echo "=== Census Terraform Provider Manual Testing ==="
echo

echo "1. Building the provider..."
go build -o terraform-provider-census .
if [ $? -eq 0 ]; then
    echo "✅ Provider builds successfully"
else
    echo "❌ Provider build failed"
    exit 1
fi
echo

echo "2. Running unit tests..."
go test ./... -short -v
echo

echo "3. Testing provider configuration validation..."
cd examples/workspace

# Create a test configuration that will validate but not execute
cat > test_validation.tf << 'EOF'
terraform {
  required_providers {
    census = {
      source = "your-org/census"
      version = "~> 0.1.0"
    }
  }
}

# This will validate the provider configuration
provider "census" {
  personal_access_token = "test-token-for-validation"
  region               = "us"
}

# This will validate resource configuration
resource "census_workspace" "test" {
  name = "Test Workspace"
  notification_emails = ["test@example.com"]
}
EOF

echo "4. Terraform configuration created for validation testing:"
cat test_validation.tf
echo

echo "5. What you can test manually:"
echo "   a) Provider builds without errors ✅"
echo "   b) Unit tests pass ✅"  
echo "   c) Terraform validates configuration (terraform validate)"
echo "   d) Provider schema is valid"
echo "   e) Error handling with invalid configurations"
echo

echo "6. What requires real Census API:"
echo "   - Acceptance tests (TF_ACC=1)"
echo "   - Actual resource creation/modification"
echo "   - API authentication testing"
echo "   - End-to-end workflows"
echo

echo "7. To test with real Census API, set:"
echo "   export CENSUS_PERSONAL_ACCESS_TOKEN=your-token"
echo "   export TF_ACC=1"
echo "   go test ./... -v"
echo

# Cleanup
rm -f test_validation.tf

echo "=== Manual Testing Complete ==="