# Census Terraform Provider Examples

This directory contains practical examples that you can use to test the Census Terraform Provider with your own Census organization and workspaces.

## Prerequisites

1. **Census Account**: You need access to a Census organization
2. **API Tokens**: Generate the appropriate tokens from your Census account:
   - **Personal Access Token**: For organization-level operations (creating/managing workspaces)
   - **Workspace Access Token**: For workspace-level operations (syncs, destinations, etc.)

## Getting Your Tokens

### Personal Access Token (Organization-level)
1. Go to your Census dashboard
2. Navigate to Settings → Developer → Personal Access Tokens
3. Generate a new token
4. Copy the token value

### Workspace Access Token (Workspace-level)  
1. Go to your specific workspace in Census
2. Navigate to Settings → API
3. Generate or copy your workspace API token

## Examples Available

### 1. Basic Workspace Management (`basic-workspace/`)
Tests workspace creation, reading, updating, and deletion.
- **Requires**: Personal Access Token
- **Tests**: Core workspace CRUD operations

### 2. Multi-Workspace Setup (`multi-workspace/`)
Creates multiple workspaces with different configurations.
- **Requires**: Personal Access Token  
- **Tests**: Workspace creation with various settings

### 3. Data Source Usage (`data-sources/`)
Demonstrates reading existing workspace information.
- **Requires**: Personal Access Token
- **Tests**: Data source functionality

## Quick Start

1. **Choose an example directory**:
   ```bash
   cd examples/basic-workspace/
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
   cd examples/basic-workspace/
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

| Example | Tests | Required Token |
|---------|-------|----------------|
| `basic-workspace/` | Workspace CRUD, terraform state management | Personal |
| `multi-workspace/` | Multiple workspace creation, different configs | Personal |
| `data-sources/` | Reading existing workspace data | Personal |

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
   # Check token permissions in Census dashboard
   ```

3. **Resource already exists**:
   ```bash
   # Import existing resources
   terraform import census_workspace.example 123
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