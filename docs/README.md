# Documentation

This directory contains documentation for the Census Terraform Provider.

## Resource Documentation

Detailed documentation for each Terraform resource:

- [`resources/workspace.md`](resources/workspace.md) - Manage Census workspaces
- [`resources/source.md`](resources/source.md) - Data warehouse connections
- [`resources/activation.md`](resources/activation.md) - Activations (use in place of `census_destination`)
- [`resources/destination.md`](resources/destination.md) - Business tool integrations
- [`resources/dataset.md`](resources/dataset.md) - SQL datasets
- [`resources/sync.md`](resources/sync.md) - Data syncs

## Data Source Documentation

Read-only data sources:

- [`data-sources/workspace.md`](data-sources/workspace.md) - Read workspace information
- [`data-sources/source.md`](data-sources/source.md) - Read source details
- [`data-sources/activation.md`](data-sources/activation.md) - Look up an activation
- [`data-sources/destination.md`](data-sources/destination.md) - Read destination configuration
- [`data-sources/dataset.md`](data-sources/dataset.md) - Read dataset information
- [`data-sources/sync.md`](data-sources/sync.md) - Read sync configuration

## Examples

See [`../examples/`](../examples/) for complete working examples.

## Contributing

See [`../CONTRIBUTING.md`](../CONTRIBUTING.md) for guidelines on contributing to documentation.
