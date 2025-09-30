# Census Terraform Provider - Roadmap

## Current Status (v0.1.0)

✅ **Production Ready - Complete Census Workflow**

The provider supports the full Census data sync pipeline:
- ✅ Workspaces - Organization and resource management
- ✅ Sources - Data warehouse connections (Snowflake, BigQuery, Postgres, etc.)
- ✅ Datasets - SQL-based data transformation
- ✅ Destinations - Business tool integrations (Salesforce, HubSpot, etc.)
- ✅ Syncs - Data synchronization with field mappings and scheduling

### Technical Achievements
- PAT-only authentication with dynamic workspace token retrieval
- OpenAPI specification compliance for all operations
- TypeSet-based field mappings (prevents order-based drift)
- Robust state management with workspace_id persistence
- Import support for all resources
- Comprehensive error handling and validation

## Future Development

### Phase 1: Execution & Monitoring (High Priority)

#### Sync Run Operations
- [ ] `census_sync_run` resource - Execute and monitor sync runs
- [ ] Trigger sync executions programmatically
- [ ] Monitor sync status and progress
- [ ] Handle sync failures and retries
- [ ] Access sync run history and logs

#### Webhook Management
- [ ] `census_webhook` resource - Event notifications
- [ ] Webhook endpoint configuration
- [ ] Event type filtering (sync completed, sync failed, etc.)
- [ ] Webhook authentication and security
- [ ] Test and validate webhook endpoints

### Phase 2: Testing & Quality (High Priority)

#### Comprehensive Test Suite
- [ ] Integration tests with real Census API
- [ ] Unit tests for all resources and data sources
- [ ] Acceptance tests using Terraform plugin testing framework
- [ ] Mock Census API server for offline testing
- [ ] Test coverage for error scenarios and edge cases
- [ ] Performance and load testing

#### Documentation
- [ ] Video tutorials and walkthroughs
- [ ] Migration guides for existing Census configurations
- [ ] Troubleshooting guides
- [ ] Best practices documentation
- [ ] Architecture decision records (ADRs)

### Phase 3: Publishing & Distribution (Medium Priority)

#### Terraform Registry
- [ ] Prepare provider metadata and manifest
- [ ] Set up GPG signing for releases
- [ ] Create automated release pipeline
- [ ] Publish to official Terraform Registry
- [ ] Version management and changelog automation

#### CI/CD Pipeline
- [ ] GitHub Actions workflows for testing
- [ ] Automated builds for multiple platforms
- [ ] Release automation
- [ ] Security scanning and vulnerability detection
- [ ] Documentation generation from code

### Phase 4: Advanced Features (Future)

#### User & Organization Management
- [ ] `census_user` resource - User management
- [ ] `census_invitation` resource - User invitations
- [ ] `census_workspace_variable` resource - Environment configuration
- [ ] Role-based access control management

#### Performance & Optimization
- [ ] Request batching and pagination improvements
- [ ] Caching strategies for API responses
- [ ] Concurrent operation support
- [ ] Rate limiting handling with backoff
- [ ] Connection pooling optimization

#### Ecosystem Integration
- [ ] Terraform Cloud integration
- [ ] CI/CD pipeline templates
- [ ] Integration with dbt, Airflow, etc.
- [ ] GitOps workflow examples
- [ ] Policy-as-code integration

## Resource Dependencies

Current workflow requires resources in this order:
1. Workspace (root resource)
2. Sources (depends on Workspace)
3. Destinations (depends on Workspace)
4. Datasets (optional - depends on Workspace + Source)
5. Syncs (depends on Workspace + Source/Dataset + Destination)

## Contributing

Contributions are welcome! Please see CONTRIBUTING.md for guidelines.

Key areas where contributions would be valuable:
- Test coverage expansion
- Additional connector support
- Documentation improvements
- Bug fixes and performance improvements

---

Last updated: 2025-09-29