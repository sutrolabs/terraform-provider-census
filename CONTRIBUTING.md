# Contributing to Census Terraform Provider

Thank you for your interest in contributing to the Census Terraform Provider! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions. We aim to foster a welcoming community for everyone.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Terraform 0.13 or later
- Git
- A Census account with API access (for testing)

### Setting Up Your Development Environment

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/sutrolabs/terraform-provider-census.git
   cd terraform-provider-census
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build the provider**:
   ```bash
   go build .
   ```

4. **Run tests**:
   ```bash
   go test ./...
   ```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Test your changes**:
   ```bash
   go test ./...
   go vet ./...
   go fmt ./...
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Description of your changes"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request**

## Coding Standards

### Go Code Style

- Follow standard Go conventions and idioms
- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch common issues
- Keep functions focused and well-named
- Add comments for exported functions and complex logic

### File Organization

```
terraform-provider-census/
├── internal/
│   ├── client/          # API client code
│   │   ├── client.go    # Base client
│   │   ├── workspace.go # Workspace operations
│   │   ├── source.go    # Source operations
│   │   └── ...
│   └── provider/        # Terraform provider code
│       ├── provider.go  # Provider configuration
│       ├── resource_*.go    # Resource implementations
│       ├── data_source_*.go # Data source implementations
│       └── utils.go     # Shared utilities
├── docs/                # Documentation
│   ├── resources/       # Resource documentation
│   └── data-sources/    # Data source documentation
└── examples/            # Usage examples
```

### Testing

- Write unit tests for all new functionality
- Include table-driven tests where appropriate
- Test error cases and edge conditions
- Aim for meaningful test coverage

Example test structure:
```go
func TestResourceWorkspaceCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   map[string]interface{}
        want    string
        wantErr bool
    }{
        // Test cases here
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Documentation

- Document all resources in `docs/resources/`
- Document all data sources in `docs/data-sources/`
- Include usage examples
- Document all arguments and attributes
- Update CHANGELOG.md for user-facing changes

## Adding New Resources

When adding a new Census resource:

1. **Create API client methods** in `internal/client/`:
   ```go
   // Add to appropriate client file or create new one
   func (c *Client) CreateResource(ctx context.Context, req *CreateResourceRequest, token string) (*Resource, error)
   func (c *Client) GetResourceWithToken(ctx context.Context, id int, token string) (*Resource, error)
   func (c *Client) UpdateResourceWithToken(ctx context.Context, id int, req *UpdateResourceRequest, token string) (*Resource, error)
   func (c *Client) DeleteResourceWithToken(ctx context.Context, id int, token string) error
   ```

2. **Implement the resource** in `internal/provider/resource_*.go`:
   - Define the schema
   - Implement CRUD operations
   - Handle errors appropriately
   - Support import functionality

3. **Add data source** in `internal/provider/data_source_*.go`:
   - Define read-only schema
   - Implement read operation

4. **Register** in `internal/provider/provider.go`:
   ```go
   ResourcesMap: map[string]*schema.Resource{
       "census_your_resource": resourceYourResource(),
   },
   DataSourcesMap: map[string]*schema.Resource{
       "census_your_resource": dataSourceYourResource(),
   },
   ```

5. **Write tests** in `internal/provider/*_test.go`

6. **Document** in `docs/resources/your_resource.md` and `docs/data-sources/your_resource.md`

7. **Add examples** in `examples/`

## API Client Guidelines

- Use `makeRequestWithToken()` for HTTP requests
- Use `handleResponse()` for response parsing
- Return structured errors with context
- Use dynamic workspace token retrieval via PAT
- Follow OpenAPI specification for request/response structures

Example:
```go
func (c *Client) GetResource(ctx context.Context, id int, token string) (*Resource, error) {
    endpoint := fmt.Sprintf("/resources/%d", id)
    resp, err := c.makeRequestWithToken(ctx, http.MethodGet, endpoint, nil, TokenTypeWorkspace, token)
    if err != nil {
        return nil, fmt.Errorf("failed to make get resource request: %w", err)
    }

    var result ResourceResponse
    if err := c.handleResponse(resp, &result); err != nil {
        return nil, fmt.Errorf("failed to get resource: %w", err)
    }

    return result.Data, nil
}
```

## Pull Request Process

1. **Ensure tests pass**: Run `go test ./...`
2. **Update documentation**: Include relevant docs changes
3. **Update CHANGELOG.md**: Add entry for your changes
4. **Link related issues**: Reference any related issue numbers
5. **Request review**: Tag maintainers for review

### PR Title Format

Use conventional commit format:
- `feat: add new resource for X`
- `fix: resolve issue with Y`
- `docs: update documentation for Z`
- `test: add tests for A`
- `refactor: improve code structure in B`

## Areas Needing Contribution

High-value contribution areas:

1. **Testing**
   - Expand unit test coverage
   - Develop mock Census API server
   - Add acceptance tests
   - Integration test improvements

2. **Resources**
   - census_sync_run (execute sync runs)
   - census_webhook (webhook management)
   - census_user (user management)
   - Additional resource enhancements

3. **Documentation**
   - Video tutorials
   - Migration guides
   - Troubleshooting documentation
   - Best practices guides

4. **Performance**
   - Request batching
   - Caching strategies
   - Connection pooling

## Questions or Problems?

- **Issues**: Open an issue on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Security**: Email security@sutrolabs.com for security concerns

## License

By contributing, you agree that your contributions will be licensed under the MIT License.