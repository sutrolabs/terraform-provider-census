# Documentation Structure

This directory contains all documentation for the Census Terraform Provider.

## File Organization

### **Root Level Documentation**
- [`../README.md`](../README.md) - Main project overview, installation, and quick start
- [`../TESTING.md`](../TESTING.md) - Testing guide and strategies
- [`../SECURITY.md`](../SECURITY.md) - Security guidelines and best practices
- [`../CHANGELOG.md`](../CHANGELOG.md) - Version history and changes
- [`../CONTRIBUTING.md`](../CONTRIBUTING.md) - Contribution guidelines
- [`../INTERNAL_INSTALLATION.md`](../INTERNAL_INSTALLATION.md) - Internal installation guide for Sutro Labs

### **Resource Documentation** (`docs/resources/`)
Technical documentation for each Terraform resource:
- [`resources/workspace.md`](resources/workspace.md) - census_workspace resource

### **Data Source Documentation** (`docs/data-sources/`)
Technical documentation for each Terraform data source:
- [`data-sources/workspace.md`](data-sources/workspace.md) - census_workspace data source

### **Example Documentation** (`examples/*/README.md`)
Usage examples and tutorials:
- [`../examples/README.md`](../examples/README.md) - Examples overview and getting started
- [`../examples/basic-workspace/README.md`](../examples/basic-workspace/README.md) - Single workspace example
- [`../examples/multi-workspace/README.md`](../examples/multi-workspace/README.md) - Multiple workspaces example
- [`../examples/data-sources/README.md`](../examples/data-sources/README.md) - Data source usage example

## Documentation Guidelines

### **When to Update Which File**

| File | Update When | Purpose |
|------|-------------|---------|
| `README.md` | Major features, installation changes | Project overview and getting started |
| `TESTING.md` | New test types, testing procedures | Testing methodology |
| `SECURITY.md` | Security practices, credential handling | Security guidelines |
| `CHANGELOG.md` | Releases, breaking changes | Version history |
| `CONTRIBUTING.md` | Contribution process changes | How to contribute |
| `docs/resources/*.md` | New resources, schema changes | Resource reference |
| `docs/data-sources/*.md` | New data sources, schema changes | Data source reference |
| `examples/*/README.md` | Example changes, new use cases | Usage tutorials |

### **Documentation Standards**

- **Resource docs**: Follow Terraform provider documentation format
- **Examples**: Include working configurations with explanations
- **Security**: Keep credentials as examples/placeholders only
- **Code blocks**: Always specify language for syntax highlighting
- **Links**: Use relative links within the repository

## Contributing to Documentation

When adding new features:

1. **Add resource documentation** - Create `docs/resources/new-resource.md`
2. **Create examples** - Add working example in `examples/`
3. **Update main README** - Add to resource list
4. **Update CHANGELOG** - Document changes for next release