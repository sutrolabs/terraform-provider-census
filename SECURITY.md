# Security Guidelines

## Credential Management

### ✅ **Safe Practices**

1. **Use `.tfvars.example` templates**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your actual credentials
   ```

2. **Environment variables** (alternative):
   ```bash
   export CENSUS_PERSONAL_ACCESS_TOKEN="your-token"
   export CENSUS_WORKSPACE_ACCESS_TOKEN="your-workspace-token"
   terraform plan  # Will use environment variables
   ```

3. **Terraform Cloud/Enterprise** (production):
   - Store tokens as sensitive workspace variables
   - Never store in version control

### ❌ **What NOT to Do**

- ❌ Commit `terraform.tfvars` files with real tokens
- ❌ Put tokens in `main.tf` or other `.tf` files
- ❌ Share tokens in chat, email, or documentation  
- ❌ Use production tokens in development/testing

## File Security

### **Gitignore Configuration**

The repository is configured to:
```gitignore
*.tfvars           # Ignore all tfvars files (contain secrets)
!*.tfvars.example  # Keep example files (no secrets)
.terraform/        # Ignore Terraform state directory
*.tfstate*         # Ignore state files (may contain secrets)
```

### **Token Security**

1. **Personal Access Tokens**:
   - Generate from: Census Dashboard → Settings → Developer → Personal Access Tokens
   - Scope: Organization-level operations
   - Rotate regularly

2. **Workspace Access Tokens**:
   - Generate from: Workspace Settings → API
   - Scope: Specific workspace operations
   - Use least-privilege principle

## Terraform State Security

### **Local Development**
- Terraform state files (`.tfstate`) may contain sensitive data
- These are gitignored by default
- Keep local state files secure

### **Team/Production Use**
- Use remote state backends (S3, Terraform Cloud, etc.)
- Enable state file encryption
- Restrict access to state files

## Example Security Checklist

Before committing:
- [ ] No `terraform.tfvars` files committed
- [ ] No hardcoded tokens in `.tf` files
- [ ] No `.tfstate` files committed
- [ ] All sensitive outputs marked with `sensitive = true`
- [ ] `.tfvars.example` files contain placeholder values only

## Incident Response

If you accidentally commit credentials:

1. **Immediately revoke** the exposed tokens in Census
2. **Remove from Git history**:
   ```bash
   git filter-branch --force --index-filter 'git rm --cached --ignore-unmatch terraform.tfvars' --prune-empty --tag-name-filter cat -- --all
   ```
3. **Generate new tokens** and update your local configuration
4. **Force push** the cleaned history (if safe to do so)

## Development Workflow

### **Safe Development Process**

1. Clone repository:
   ```bash
   git clone <repo>
   cd terraform-provider-census
   ```

2. Build provider:
   ```bash
   make dev
   ```

3. Set up credentials securely:
   ```bash
   cd examples/basic-workspace/
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your tokens (this file is gitignored)
   ```

4. Test:
   ```bash
   terraform init
   terraform plan
   ```

5. Commit changes (credentials are automatically excluded):
   ```bash
   git add .
   git commit -m "Add new feature"
   ```

## Production Considerations

1. **Use separate Census organizations** for development and production
2. **Implement token rotation** policies  
3. **Monitor token usage** in Census audit logs
4. **Use workspace-level tokens** when possible (least privilege)
5. **Implement approval processes** for production changes

## Reporting Security Issues

If you find a security vulnerability:
- **DO NOT** create a public GitHub issue
- Email security concerns to: [security@your-org.com]
- Include detailed reproduction steps
- Allow reasonable time for response before public disclosure