# Basic Workspace Example

This example demonstrates the core functionality of the Census Terraform Provider by creating and managing a single workspace.

## What This Example Tests

✅ **Resource Creation**: Creates a Census workspace with custom configuration  
✅ **Resource Reading**: Uses data source to read the created workspace  
✅ **State Management**: Manages Terraform state for Census resources  
✅ **Output Values**: Demonstrates accessing resource attributes  
✅ **Data Consistency**: Validates resource vs data source consistency  

## Prerequisites

1. **Census Account**: Access to a Census organization
2. **Personal Access Token**: Generate from Census Dashboard → Settings → Developer → Personal Access Tokens
3. **Provider Built**: Run `make dev` from the root directory to build the provider locally

## Quick Start

1. **Copy configuration**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Add your token** to `terraform.tfvars`:
   ```hcl
   census_personal_token = "your-actual-token-here"
   ```

3. **Test the configuration**:
   ```bash
   terraform init
   terraform validate
   terraform plan
   ```

4. **Apply (creates real resources)**:
   ```bash
   terraform apply
   ```

5. **View outputs**:
   ```bash
   terraform output
   
   # View sensitive outputs
   terraform output workspace_api_key
   ```

6. **Clean up**:
   ```bash
   terraform destroy
   ```

## Configuration Options

### Required
- `census_personal_token`: Your Census personal access token

### Optional
- `census_region`: "us" (default) or "eu"
- `workspace_name`: Custom workspace name (default: "Terraform Test Workspace")
- `notification_emails`: List of email addresses for alerts
- `return_api_key`: Whether to return workspace API key (default: true)

## Expected Outputs

After running `terraform apply`, you should see outputs like:

```
workspace_id = "123"
workspace_name = "My Terraform Test Workspace"
workspace_organization_id = "456"
workspace_created_at = "2025-01-XX"
workspace_notification_emails = ["your-email@example.com"]
names_match = true
org_ids_match = true
```

## Validation Checks

This example includes validation outputs:
- `names_match`: Ensures resource and data source return the same workspace name
- `org_ids_match`: Ensures consistent organization ID between resource and data source

## Troubleshooting

### Common Issues

1. **"Provider not found"**:
   ```bash
   # Build provider locally
   cd ../../
   make dev
   cd examples/basic-workspace/
   ```

2. **Authentication error**:
   - Verify your token in `terraform.tfvars`
   - Check token permissions in Census dashboard
   - Ensure you're using a personal access token with organization-level permissions

3. **Region mismatch**:
   - Set `census_region = "eu"` if using Census EU instance

### Debug Mode

Enable detailed logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## What You Can Test

### ✅ Working Features
- Workspace creation with custom name
- Notification email configuration  
- API key retrieval during creation
- Reading workspace data via data source
- Terraform state management
- Resource updates (change workspace_name and re-apply)
- Resource destruction

### 🚧 Future Features (Not Yet Implemented)
- Syncs management  
- Destinations configuration  
- Sources management

## Next Steps

After testing this example:
1. Try the `multi-workspace` example for more complex scenarios
2. Test the `data-sources` example to read existing workspaces
3. Experiment with updating workspace configuration
4. Test importing existing workspaces: `terraform import census_workspace.test 123`