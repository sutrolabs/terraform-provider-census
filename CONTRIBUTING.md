# Contributing to Census Terraform Provider

Thank you for your interest in contributing! This document provides guidelines for contributing to the Census Terraform Provider.

## Getting Started

### Prerequisites

- Go 1.21+
- Terraform 0.13+
- A Census account with API access (for testing)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/sutrolabs/terraform-provider-census.git
cd terraform-provider-census

# Install dependencies
make deps

# Build the provider
make build

# Run tests
make test
```

### Local Installation

Install the provider locally for testing with Terraform:

```bash
make install-local
```

This installs the provider to `~/.terraform.d/plugins/` where Terraform can find it automatically.

## Development Workflow

1. Create a feature branch: `git checkout -b feature/your-feature-name`
2. Make your changes
3. Run tests: `make test`
4. Format code: `make fmt`
5. Run linters: `make vet`
6. Commit your changes
7. Push and open a pull request

## Code Standards

- Follow standard Go conventions
- Run `make fmt` before committing
- Run `make vet` to catch common issues
- Add tests for new functionality
- Update documentation for user-facing changes

## Testing

```bash
# Run unit tests
make test

# Run specific package tests
go test ./census/provider -v

# Run with coverage
go test ./... -cover
```

## Adding New Resources

When adding a new Census resource:

1. **Create API client methods** in `census/client/`:
   ```go
   func (c *Client) CreateResource(ctx context.Context, req *CreateResourceRequest, token string) (*Resource, error)
   func (c *Client) GetResourceWithToken(ctx context.Context, id int, token string) (*Resource, error)
   func (c *Client) UpdateResourceWithToken(ctx context.Context, id int, req *UpdateResourceRequest, token string) (*Resource, error)
   func (c *Client) DeleteResourceWithToken(ctx context.Context, id int, token string) error
   ```

2. **Implement the resource** in `census/provider/resource_*.go` with CRUD operations

3. **Add data source** in `census/provider/data_source_*.go`

4. **Register** in `census/provider/provider.go`:
   ```go
   ResourcesMap: map[string]*schema.Resource{
       "census_your_resource": resourceYourResource(),
   },
   ```

5. **Write tests** in `*_test.go` files

6. **Document** in `docs/resources/your_resource.md`

7. **Add examples** in `examples/`

## Pull Request Process

1. Ensure tests pass: `make test`
2. Update documentation for user-facing changes
3. Update CHANGELOG.md with your changes
4. Reference any related issue numbers
5. Use conventional commit format for PR titles:
   - `feat: add new resource for X`
   - `fix: resolve issue with Y`
   - `docs: update documentation for Z`

## API Guidelines

- Consult [Census API Documentation](https://developers.getcensus.com/api-reference/introduction/overview) for API structure
- Check [OpenAPI specifications](https://developers.getcensus.com/openapi/compiled/workspace_management.yaml) for request/response formats
- Use dynamic workspace token retrieval via PAT
- Follow existing patterns in `census/client/` for consistency

## Questions?

- Open an issue on GitHub
- Check existing issues and pull requests first

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
