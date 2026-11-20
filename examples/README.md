# Fivetran Activations Terraform Provider Examples

This directory contains practical examples that you can use to test the Fivetran Activations Terraform Provider with your own Census organization.

## Prerequisites

1. **Census Account**: You need access to a Census organization
2. **API Token**: Generate a Personal Access Token from your Census account:
   - **Personal Access Token**: Used for all operations

## Getting Your Token

### Personal Access Token
1. Go to your Activations dashboard
2. Navigate to Settings → Developer → Personal Access Tokens
3. Generate a new token
4. Copy the token value

## Examples Available

### 1. Staging Environment (`staging-example/`)
Complete data pipeline using the Activations staging environment.
- **Requires**: Personal Access Token, staging credentials
- **Tests**: Source, destination, and sync with staging URL
- **Complexity**: Intermediate

### 2. Complete Activations Setup (`complete-census-setup/`)
Full-featured example with all 4 resource types.
- **Requires**: Personal Access Token, source/destination credentials
- **Tests**: Sources, activations, datasets, syncs
- **Complexity**: Advanced

## Quick Start

1. **Choose an example directory**:
   ```bash
   cd examples/staging-example/
   ```

2. **Copy the variables template**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

3. **Add your credentials** to `terraform.tfvars`:
   ```hcl
   census_personal_token = "your-personal-access-token-here"
   ```

4. **Build the provider locally**:
   ```bash
   cd ../../  # Go back to root directory
   make dev   # Build and install provider locally
   ```

5. **Test the configuration**:
   ```bash
   cd examples/staging-example/
   terraform init
   terraform validate
   terraform plan
   terraform apply  # Only if you want to create real resources!
   ```

## Safety Tips

- **Start with `terraform plan`** to see what will be created
- **Use a test Census organization** if available
- **Review all configurations** before running `terraform apply`
- **Clean up resources** with `terraform destroy` when done
- **Keep your tokens secure** - never commit `terraform.tfvars` to git

## 🔒 Security Notes

- ✅ **`terraform.tfvars` is gitignored** - safe to add your real tokens
- ✅ **`terraform.tfvars.example` is tracked** - contains only placeholder values  
- ✅ **All sensitive outputs are marked** with `sensitive = true`
- ⚠️ **Always use `.tfvars.example` as your starting point**:
  ```bash
  cp terraform.tfvars.example terraform.tfvars
  # Edit terraform.tfvars with your actual credentials
  ```

## What Each Example Tests

| Example | What It Demonstrates | Complexity |
|---------|---------------------|------------|
| `staging-example/` | Staging environment with source → destination → sync | Intermediate |
| `complete-census-setup/` | Full Activations workflow with all 4 resources | Advanced |

## Troubleshooting

### Common Issues

1. **Provider not found**:
   ```bash
   # Make sure you built the provider locally
   make dev
   ```

2. **Authentication errors**:
   ```bash
   # Verify your token in terraform.tfvars
   # Check token permissions in Activations dashboard
   ```

3. **Resource already exists**:
   ```bash
   # Import existing resources
   terraform import census_source.example 123
   ```

### Debug Mode

Enable debug logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## Contributing Examples

When adding new examples:
1. Create a new directory with a descriptive name
2. Include `main.tf`, `variables.tf`, and `terraform.tfvars.example`
3. Add documentation in a `README.md`
4. Update this main README with the new example