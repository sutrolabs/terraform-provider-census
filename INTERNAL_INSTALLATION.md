# Internal Installation Guide - Sutro Labs

This guide is for Sutro Labs engineers who want to use the Census Terraform Provider internally before it's publicly released.

## Prerequisites

- Terraform 0.13 or later installed
- Access to this private GitHub repository
- macOS or Linux operating system

## Installation Methods

Choose one of these methods based on your use case:

### Method 1: Filesystem Mirror (Recommended for Regular Use)

This method installs the provider so Terraform can find it automatically.

#### Step 1: Download the Provider Binary

**For macOS (Apple Silicon M1/M2/M3):**
```bash
# Download from GitHub releases
curl -LO https://github.com/sutrolabs/terraform-provider-census/releases/download/v0.1.0/terraform-provider-census_darwin_arm64

# Make it executable
chmod +x terraform-provider-census_darwin_arm64
```

**For macOS (Intel):**
```bash
curl -LO https://github.com/sutrolabs/terraform-provider-census/releases/download/v0.1.0/terraform-provider-census_darwin_amd64
chmod +x terraform-provider-census_darwin_amd64
```

**For Linux:**
```bash
curl -LO https://github.com/sutrolabs/terraform-provider-census/releases/download/v0.1.0/terraform-provider-census_linux_amd64
chmod +x terraform-provider-census_linux_amd64
```

#### Step 2: Create Plugin Directory Structure

```bash
# Create the directory structure
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/darwin_arm64

# For Intel Mac, use: darwin_amd64
# For Linux, use: linux_amd64
```

#### Step 3: Move Binary to Plugin Directory

**For macOS (Apple Silicon):**
```bash
mv terraform-provider-census_darwin_arm64 \
  ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/darwin_arm64/terraform-provider-census_v0.1.0
```

**For macOS (Intel):**
```bash
mv terraform-provider-census_darwin_amd64 \
  ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/darwin_amd64/terraform-provider-census_v0.1.0
```

**For Linux:**
```bash
mv terraform-provider-census_linux_amd64 \
  ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/linux_amd64/terraform-provider-census_v0.1.0
```

#### Step 4: Use in Your Terraform Configuration

Create a `main.tf` file:

```hcl
terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "0.1.0"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
  region               = "us"  # or "eu"
}

# Example resource
resource "census_workspace" "example" {
  name = "My Workspace"
  notification_emails = ["team@sutrolabs.com"]
}
```

Run Terraform:

```bash
terraform init    # Should find the local provider
terraform plan
terraform apply
```

### Method 2: Build from Source (For Development)

If you're actively developing or want the latest changes:

#### Step 1: Clone and Build

```bash
# Clone the repository
git clone git@github.com:sutrolabs/terraform-provider-census.git
cd terraform-provider-census

# Build the provider
go build -o terraform-provider-census

# Or use Make
make build
```

#### Step 2: Use Dev Override

Create or edit `~/.terraform.d/terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "sutrolabs/census" = "/path/to/terraform-provider-census"
  }
  direct {}
}
```

Replace `/path/to/terraform-provider-census` with the actual path where you built the binary.

#### Step 3: Use in Terraform

With dev overrides, Terraform will use your local binary. Just run:

```bash
terraform init   # Will use the overridden provider
terraform plan
terraform apply
```

**Note:** When using dev overrides, you'll see a warning that the provider is being overridden. This is expected.

## Quick Start Example

Here's a complete example to get you started:

1. **Install the provider** using Method 1 above

2. **Create a new directory** for your Terraform config:
   ```bash
   mkdir census-test && cd census-test
   ```

3. **Create `terraform.tfvars`** with your Census credentials:
   ```hcl
   census_personal_token = "your-census-personal-access-token"
   ```

4. **Create `variables.tf`**:
   ```hcl
   variable "census_personal_token" {
     description = "Census personal access token"
     type        = string
     sensitive   = true
   }
   ```

5. **Create `main.tf`**:
   ```hcl
   terraform {
     required_providers {
       census = {
         source  = "sutrolabs/census"
         version = "0.1.0"
       }
     }
   }

   provider "census" {
     personal_access_token = var.census_personal_token
     region               = "us"
   }

   resource "census_workspace" "example" {
     name = "Test Workspace"
     notification_emails = ["you@sutrolabs.com"]
   }
   ```

6. **Run Terraform**:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Troubleshooting

### "Provider not found" Error

If you see:
```
│ Error: Failed to query available provider packages
│ Could not retrieve the list of available versions for provider sutrolabs/census
```

**Solutions:**
- Verify the binary is in the correct directory path
- Check the binary is executable: `chmod +x /path/to/binary`
- Verify the directory structure matches exactly (case-sensitive)
- Make sure you're using the correct platform (darwin_arm64, darwin_amd64, linux_amd64)

### "Binary not executable" Error

If you see permission denied errors:

```bash
chmod +x ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/*/terraform-provider-census_v0.1.0
```

### Version Mismatch

If Terraform complains about version constraints:
- Ensure your `required_providers` block specifies `version = "0.1.0"`
- The binary must be in a folder named `0.1.0`
- The binary filename must end with `_v0.1.0`

### Finding Your Platform

Not sure if you're on arm64 or amd64?

```bash
uname -m
# Output: arm64 = Apple Silicon (M1/M2/M3)
# Output: x86_64 = Intel (amd64)
```

## Getting Your Census API Token

1. Log in to Census at https://app.getcensus.com
2. Go to Settings → API Access
3. Generate a new Personal Access Token (PAT)
4. Copy the token and use it in your Terraform configuration

**Never commit your token to git!** Always use `terraform.tfvars` and add it to `.gitignore`.

## Available Resources

The provider currently supports:

- **census_workspace** - Manage Census workspaces
- **census_source** - Data warehouse connections (Snowflake, BigQuery, Postgres, etc.)
- **census_destination** - Business tool integrations (Salesforce, HubSpot, etc.)
- **census_dataset** - SQL datasets for data transformation
- **census_sync** - Data syncs between sources and destinations

See the [complete example](examples/complete-census-setup/) for comprehensive usage.

## Documentation

- **Resources**: See `docs/resources/` for detailed documentation on each resource
- **Data Sources**: See `docs/data-sources/` for read-only data source documentation
- **Examples**: See `examples/` directory for complete working examples

## Getting Help

- **Issues**: Open an issue on this GitHub repository
- **Questions**: Ask in #eng-data-platform Slack channel
- **Bugs**: Please include your Terraform version, OS, and error messages

## Updating to New Versions

When a new version is released:

1. Download the new binary from GitHub releases
2. Create the new version directory (e.g., `0.2.0`)
3. Move the binary to the new directory
4. Update your `required_providers` version constraint
5. Run `terraform init -upgrade`

## Security Notes

- Keep your Census API tokens secure
- Never commit credentials to version control
- Use `sensitive = true` in variable declarations
- Consider using environment variables for sensitive data