#!/bin/bash

# Census Terraform Provider Testing Script
# This script helps you test the provider with your real Census credentials

set -e

echo "🏗️  Census Terraform Provider Testing Script"
echo "============================================="
echo

# Check if user has set up their environment
if [ -z "$CENSUS_PERSONAL_ACCESS_TOKEN" ]; then
    echo "⚠️  Environment variable CENSUS_PERSONAL_ACCESS_TOKEN is not set."
    echo "   Please set it with your Census personal access token:"
    echo "   export CENSUS_PERSONAL_ACCESS_TOKEN='your-token-here'"
    echo
fi

echo "🔧 Step 1: Building the provider locally..."
make build
if [ $? -eq 0 ]; then
    echo "✅ Provider built successfully"
else
    echo "❌ Provider build failed"
    exit 1
fi
echo

echo "🧪 Step 2: Running unit tests..."
make test
if [ $? -eq 0 ]; then
    echo "✅ Unit tests passed"
else
    echo "❌ Unit tests failed"
    exit 1
fi
echo

echo "📦 Step 3: Installing provider locally..."
make dev
if [ $? -eq 0 ]; then
    echo "✅ Provider installed locally"
else
    echo "❌ Provider installation failed"
    exit 1
fi
echo

echo "📋 Step 4: Available examples to test with your Census account:"
echo
echo "   1. basic-workspace/     - Create and manage a single workspace"
echo "   2. multi-workspace/     - Create multiple workspaces with different configs"
echo "   3. data-sources/        - Read existing workspace information"
echo

echo "🚀 Step 5: Quick validation test"
echo "   Let's validate the basic-workspace example configuration:"
echo

cd examples/basic-workspace

# Check if terraform.tfvars exists
if [ ! -f "terraform.tfvars" ]; then
    echo "   Creating terraform.tfvars from example..."
    cp terraform.tfvars.example terraform.tfvars
    echo "   ⚠️  Please edit examples/basic-workspace/terraform.tfvars with your Census token!"
    echo
fi

echo "   Running terraform init..."
terraform init -no-color
echo

echo "   Running terraform validate..."
terraform validate
if [ $? -eq 0 ]; then
    echo "✅ Terraform configuration is valid"
else
    echo "❌ Terraform configuration validation failed"
    exit 1
fi
echo

echo "   Running terraform plan (dry run)..."
if [ -n "$CENSUS_PERSONAL_ACCESS_TOKEN" ]; then
    # If token is set in environment, try a plan
    terraform plan -no-color -input=false -var="census_personal_token=$CENSUS_PERSONAL_ACCESS_TOKEN"
    PLAN_RESULT=$?
    
    if [ $PLAN_RESULT -eq 0 ]; then
        echo "✅ Terraform plan succeeded - provider can communicate with Census API!"
        echo "   🎉 Your Census Terraform Provider is ready to use!"
        echo
        echo "   To apply the plan (create real resources):"
        echo "   cd examples/basic-workspace && terraform apply"
        echo
    else
        echo "❌ Terraform plan failed - check your Census token and permissions"
        echo "   Common issues:"
        echo "   - Invalid or expired token"
        echo "   - Incorrect region (try census_region = \"eu\" if using EU instance)"  
        echo "   - Network connectivity issues"
    fi
else
    echo "   ⏭️  Skipping plan - CENSUS_PERSONAL_ACCESS_TOKEN not set"
    echo "   Set your token and run: terraform plan"
fi

cd ../../

echo
echo "📖 Next Steps:"
echo "   1. Edit examples/*/terraform.tfvars files with your Census credentials"
echo "   2. Choose an example directory: cd examples/basic-workspace/"
echo "   3. Run: terraform plan (to see what will be created)"
echo "   4. Run: terraform apply (to create real Census resources)"
echo "   5. Run: terraform destroy (to clean up when done)"
echo
echo "🔍 Debugging:"
echo "   - Enable debug logging: export TF_LOG=DEBUG"
echo "   - View provider logs: terraform apply"
echo "   - Check Census dashboard to verify resources"
echo
echo "📚 Documentation:"
echo "   - Provider docs: README.md"
echo "   - Example docs: examples/*/README.md" 
echo "   - Testing guide: TESTING.md"
echo
echo "🎯 What you can test:"
echo "   ✅ Workspace creation, reading, updating, deletion"
echo "   ✅ Notification email configuration"
echo "   ✅ API key retrieval"
echo "   ✅ Data source functionality" 
echo "   ✅ Terraform state management"
echo "   ✅ Error handling and validation"
echo
echo "🚧 Future features (not yet implemented):"
echo "   - Syncs management"
echo "   - Destinations configuration"
echo "   - Sources management"
echo "   - Sync runs and webhooks"
echo

if [ -n "$CENSUS_PERSONAL_ACCESS_TOKEN" ] && [ $PLAN_RESULT -eq 0 ]; then
    echo "🎉 SUCCESS: Your Census Terraform Provider is working correctly!"
else
    echo "⚠️  NEXT: Set your Census token and test with real API"
fi