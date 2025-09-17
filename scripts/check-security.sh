#!/bin/bash

# Security validation script for Census Terraform Provider

echo "🔒 Census Terraform Provider Security Check"
echo "==========================================="
echo

# Check .gitignore configuration
echo "1. Checking .gitignore configuration..."
if grep -q "^\*.tfvars$" .gitignore && grep -q "^!\*.tfvars.example$" .gitignore; then
    echo "   ✅ .gitignore correctly configured for tfvars files"
else
    echo "   ❌ .gitignore missing proper tfvars configuration"
    echo "   Should contain:"
    echo "     *.tfvars"
    echo "     !*.tfvars.example"
fi

# Check for accidentally committed tfvars files
echo
echo "2. Checking for accidentally committed credentials..."
TFVARS_FILES=$(find . -name "terraform.tfvars" -not -path "./.git/*" -not -name "*.example")
if [ -z "$TFVARS_FILES" ]; then
    echo "   ✅ No terraform.tfvars files found in repository"
else
    echo "   ⚠️  Found terraform.tfvars files:"
    echo "$TFVARS_FILES"
    echo "   These files should be removed and added to .gitignore"
fi

# Check for hardcoded tokens in .tf files
echo
echo "3. Scanning for hardcoded tokens in .tf files..."
if grep -r -i "census.*token.*=.*[a-zA-Z0-9]" --include="*.tf" --exclude-dir=".git" . | grep -v "your-.*-token-here" | grep -v "example"; then
    echo "   ⚠️  Potential hardcoded tokens found in .tf files"
else
    echo "   ✅ No hardcoded tokens found in .tf files"
fi

# Check example files have placeholder values
echo
echo "4. Validating .tfvars.example files..."
EXAMPLE_FILES=$(find examples/ -name "terraform.tfvars.example" 2>/dev/null)
if [ -n "$EXAMPLE_FILES" ]; then
    for file in $EXAMPLE_FILES; do
        if grep -q "your-.*-token-here" "$file"; then
            echo "   ✅ $file contains placeholder values"
        else
            echo "   ⚠️  $file may contain real credentials"
        fi
    done
else
    echo "   ⚠️  No .tfvars.example files found in examples/"
fi

# Check sensitive outputs are marked
echo  
echo "5. Checking for unmarked sensitive outputs..."
SENSITIVE_OUTPUTS=$(grep -r "api_key\|token\|secret" --include="*.tf" examples/ | grep "output\|value" | grep -v "sensitive.*=.*true")
if [ -z "$SENSITIVE_OUTPUTS" ]; then
    echo "   ✅ All sensitive outputs appear to be marked as sensitive"
else
    echo "   ⚠️  Potentially unmarked sensitive outputs:"
    echo "$SENSITIVE_OUTPUTS"
fi

# Check for state files
echo
echo "6. Checking for committed Terraform state files..."
STATE_FILES=$(find . -name "*.tfstate*" -not -path "./.git/*")
if [ -z "$STATE_FILES" ]; then
    echo "   ✅ No Terraform state files found in repository"
else
    echo "   ⚠️  Found Terraform state files (may contain secrets):"
    echo "$STATE_FILES"
fi

# Summary
echo
echo "🔒 Security Check Summary:"
echo "========================="

ISSUES=0

if ! (grep -q "^\*.tfvars$" .gitignore && grep -q "^!\*.tfvars.example$" .gitignore); then
    echo "❌ Fix .gitignore configuration"
    ISSUES=$((ISSUES + 1))
fi

if [ -n "$TFVARS_FILES" ]; then
    echo "❌ Remove terraform.tfvars files from repository"
    ISSUES=$((ISSUES + 1))
fi

if [ -n "$STATE_FILES" ]; then
    echo "❌ Remove Terraform state files from repository"
    ISSUES=$((ISSUES + 1))
fi

if [ $ISSUES -eq 0 ]; then
    echo "✅ All security checks passed!"
    echo
    echo "🚀 Safe to commit and share this repository"
    echo "📝 Remember to keep terraform.tfvars files local and private"
else
    echo "⚠️  Found $ISSUES security issues that should be addressed"
    echo
    echo "📖 See SECURITY.md for detailed guidelines"
fi