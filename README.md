# Terraform Provider for Census

A Terraform provider for managing [Census](https://getcensus.com) resources. Census enables you to sync data from your warehouse to all your operational tools, and this provider allows you to manage Census infrastructure as code.

## Features

- **Multi-region support**: Works with both US and EU Census instances
- **Dual authentication**: Supports both personal access tokens and workspace access tokens
- **Workspace management**: Create, update, and delete Census workspaces
- **Import support**: Import existing Census resources into Terraform state

## Installation

### Terraform 0.13+

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    census = {
      source = "your-org/census"
      version = "~> 0.1.0"
    }
  }
}
```

### Manual Installation

1. Download the provider binary from the [releases page](https://github.com/your-org/terraform-provider-census/releases)
2. Place it in your Terraform plugins directory
3. Run `terraform init`

## Authentication

The Census provider uses a Personal Access Token (PAT) for all operations:

```hcl
provider "census" {
  personal_access_token = "your-personal-access-token"
  region               = "us" # or "eu"
}
```

### Environment Variables

You can also set the authentication token using environment variables:

```bash
export CENSUS_PERSONAL_ACCESS_TOKEN="your-personal-token"
```

### How Authentication Works

- **Personal Access Token**: Used for all operations including workspace management and workspace-scoped operations
- **Dynamic Workspace Tokens**: The provider automatically retrieves workspace-specific API keys as needed using your PAT
- **Team Collaboration**: Only the PAT needs to be shared - workspace tokens are fetched dynamically

## Configuration

| Argument | Description | Default | Required |
|----------|-------------|---------|----------|
| `personal_access_token` | Personal access token for Census APIs | - | Yes |
| `region` | Census region (`us` or `eu`) | `us` | No |
| `base_url` | Custom base URL for Census API | Determined by region | No |

## Quick Testing with Your Census Account

The fastest way to test this provider is with the included examples:

1. **Build and install the provider**:
   ```bash
   make dev
   ```

2. **Run the testing script**:
   ```bash
   ./test-with-census.sh
   ```

3. **Or test manually with an example**:
   ```bash
   cd examples/basic-workspace/
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your Census token
   terraform init
   terraform plan
   terraform apply  # Creates real Census resources
   ```

### Available Examples

| Example | Purpose | What it Tests |
|---------|---------|---------------|
| `basic-workspace/` | Single workspace CRUD | Core functionality, data sources, state management |
| `multi-workspace/` | Multiple workspaces | Complex scenarios, different configurations |
| `data-sources/` | Read existing workspaces | Data source functionality, existing resource discovery |

## Usage Examples

### Complete Example

```hcl
terraform {
  required_providers {
    census = {
      source = "your-org/census"
      version = "~> 0.1.0"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
  region               = "us"
}

variable "census_personal_token" {
  description = "Census personal access token"
  type        = string
  sensitive   = true
}

resource "census_workspace" "data_team" {
  name = "Data Team Workspace"
  notification_emails = [
    "data-alerts@company.com"
  ]
}
```

## Resources

### Currently Available
- [`census_workspace`](docs/resources/workspace.md) - Manage Census workspaces

### Planned Resources  
See [TODO.md](TODO.md) for the complete development roadmap including:
- `census_sync` - Manage data syncs
- `census_destination` - Configure sync destinations  
- `census_source` - Manage data sources
- `census_dataset` - Data modeling and transformations
- And more...

## Data Sources

### Currently Available
- [`census_workspace`](docs/data-sources/workspace.md) - Read Census workspace information

### Planned Data Sources
All planned resources will have corresponding data sources for read operations.

## Development

### Requirements

- Go 1.21+
- Terraform 0.13+

### Building from Source

```bash
git clone https://github.com/your-org/terraform-provider-census
cd terraform-provider-census
go build .
```

### Testing

```bash
# Run unit tests
go test ./...

# Run acceptance tests (requires Census API access)
TF_ACC=1 go test ./... -v
```

### Local Development

1. Build the provider:
   ```bash
   go build -o terraform-provider-census
   ```

2. Create a local Terraform configuration file:
   ```bash
   # dev.tfrc
   provider_installation {
     dev_overrides {
       "your-org/census" = "/path/to/terraform-provider-census"
     }
     direct {}
   }
   ```

3. Set the Terraform CLI config:
   ```bash
   export TF_CLI_CONFIG_FILE=/path/to/dev.tfrc
   ```

4. Test with Terraform:
   ```bash
   terraform init
   terraform plan
   ```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run tests: `go test ./...`
5. Commit your changes: `git commit -am 'Add feature'`
6. Push to the branch: `git push origin feature-name`
7. Submit a pull request

## API Reference

This provider is built against the Census API. For detailed API documentation, visit:
- [Census API Documentation](https://developers.getcensus.com/api-reference/introduction/overview)

### Live OpenAPI Specifications
For development and debugging, the latest OpenAPI specifications are available at:
- [Organization Management API](https://developers.getcensus.com/openapi/compiled/organization_management.yaml) - Workspace management, user management
- [Workspace Management API](https://developers.getcensus.com/openapi/compiled/workspace_management.yaml) - Sources, destinations, syncs, and other workspace-scoped operations

## Support

- [GitHub Issues](https://github.com/your-org/terraform-provider-census/issues)
- [Census Documentation](https://docs.getcensus.com/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for details about changes in each version.

---

**Note**: This provider is not officially maintained by Census. It is a community project designed to provide Terraform integration with Census services.