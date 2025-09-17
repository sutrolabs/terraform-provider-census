# Census Terraform Provider - Development TODO

This document tracks all planned features, improvements, and tasks for the Census Terraform Provider.

## Current Status

✅ **Phase 1 Complete: Foundation & Basic Workspace Management**
- Provider configuration with dual authentication (personal/workspace tokens)
- Multi-region support (US/EU)
- Workspace resource with full CRUD operations
- Workspace data source
- Basic testing framework
- Documentation and examples
- Local development setup

## Next Development Phases

### 🚧 Phase 2: Core Census Operations (High Priority)

#### ⏳ Sync Resources
- [ ] **census_sync Resource** - Create and manage Census data syncs
  - [ ] Add sync API client methods (`internal/client/sync.go`)
  - [ ] Create sync resource implementation (`internal/provider/resource_sync.go`)
  - [ ] Add sync data source (`internal/provider/data_source_sync.go`)
  - [ ] Sync configuration schema (source, destination, field mappings)
  - [ ] Sync scheduling and trigger options
  - [ ] Sync status and monitoring
  - [ ] Example configurations
  - [ ] Unit and integration tests

#### ⏳ Destination Resources  
- [ ] **census_destination Resource** - Manage data sync destinations
  - [ ] Add destination API client methods (`internal/client/destination.go`)
  - [ ] Create destination resource implementation
  - [ ] Support major destination types (Salesforce, HubSpot, etc.)
  - [ ] Connection credential management
  - [ ] Connection testing and validation
  - [ ] Destination-specific configuration schemas

#### ⏳ Source Resources
- [ ] **census_source Resource** - Manage data sources
  - [ ] Add source API client methods (`internal/client/source.go`) 
  - [ ] Create source resource implementation
  - [ ] Database connection support (Snowflake, BigQuery, etc.)
  - [ ] Connection string and credential management
  - [ ] Source validation and testing
  - [ ] Schema introspection capabilities

### 📋 Phase 3: Data & Execution Management (Medium Priority)

#### ⏳ Dataset Resources
- [ ] **census_dataset Resource** - Data modeling and transformation
  - [ ] Add dataset API client methods (`internal/client/dataset.go`)
  - [ ] Create dataset resource implementation  
  - [ ] SQL model definitions
  - [ ] Column mapping and transformations
  - [ ] Dataset validation and preview
  - [ ] Dependency management between datasets

#### ⏳ Sync Run Operations
- [ ] **census_sync_run Resource** - Execute and monitor syncs
  - [ ] Add sync run API client methods (`internal/client/sync_run.go`)
  - [ ] Trigger sync executions
  - [ ] Monitor sync status and progress
  - [ ] Handle sync failures and retries
  - [ ] Sync run history and logging
  - [ ] Scheduling and automated triggers

#### ⏳ Webhook Management
- [ ] **census_webhook Resource** - Event notifications
  - [ ] Add webhook API client methods (`internal/client/webhook.go`)
  - [ ] Create webhook resource implementation
  - [ ] Webhook endpoint configuration
  - [ ] Event type filtering
  - [ ] Webhook authentication and security

### 👥 Phase 4: Organization & User Management (Lower Priority)

#### ⏳ User Management (Complete Implementation)
- [ ] **census_user Resource** - User management (currently read-only)
  - [ ] Add user creation capabilities
  - [ ] User role management
  - [ ] User status and permissions
  - [ ] Bulk user operations

#### ⏳ Invitation Management (Complete Implementation)  
- [ ] **census_invitation Resource** - User invitations (basic implementation exists)
  - [ ] Enhance invitation workflows
  - [ ] Invitation expiry and management
  - [ ] Workspace-specific invitations
  - [ ] Bulk invitation operations

#### ⏳ Workspace Variables
- [ ] **census_workspace_variable Resource** - Environment configuration
  - [ ] Add workspace variable API client methods
  - [ ] Variable creation and management
  - [ ] Secret variable handling
  - [ ] Variable validation and scoping
  - [ ] Bulk variable operations

### 🧪 Phase 5: Testing & Quality Assurance

#### ⏳ Comprehensive Testing
- [ ] **Integration Test Suite**
  - [ ] Real Census API integration tests
  - [ ] Error scenario coverage
  - [ ] Rate limiting and retry logic
  - [ ] Cross-resource dependency testing
  - [ ] Performance and load testing

- [ ] **Mock Server Enhancement**
  - [ ] Complete mock Census API implementation
  - [ ] All resource types supported in mock
  - [ ] Realistic error scenarios
  - [ ] State persistence between tests

- [ ] **Acceptance Test Coverage**
  - [ ] Full resource lifecycle testing
  - [ ] Import functionality testing
  - [ ] Update and drift detection
  - [ ] Error handling validation

### 📚 Phase 6: Documentation & Publishing

#### ⏳ Terraform Registry Preparation
- [ ] **Provider Documentation**
  - [ ] Complete resource documentation
  - [ ] Data source documentation  
  - [ ] Configuration examples
  - [ ] Migration guides
  - [ ] Best practices documentation

- [ ] **Registry Publication**
  - [ ] Provider metadata and manifest
  - [ ] Release automation pipeline
  - [ ] Versioning and changelog management
  - [ ] GPG signing for releases

### 🔧 Phase 7: Production Features

#### ⏳ Advanced Provider Features
- [ ] **Enhanced Authentication**
  - [ ] Token refresh mechanisms
  - [ ] Multiple authentication methods
  - [ ] Role-based access control
  - [ ] Audit logging

- [ ] **Performance Optimization**
  - [ ] Request batching and pagination
  - [ ] Caching strategies
  - [ ] Concurrent operation support
  - [ ] Rate limiting handling

- [ ] **Error Handling & Resilience**
  - [ ] Retry mechanisms with backoff
  - [ ] Circuit breaker patterns
  - [ ] Graceful degradation
  - [ ] Detailed error reporting

## Implementation Priority

### **Next Immediate Tasks** (Recommended Order)
1. **census_sync Resource** - Most critical Census functionality
2. **census_destination Resource** - Required for sync operations  
3. **census_source Resource** - Required for sync operations
4. **Integration Testing** - Ensure reliability
5. **census_dataset Resource** - Advanced data modeling

### **Resource Dependencies**
- Syncs depend on: Workspaces, Sources, Destinations
- Sync Runs depend on: Syncs
- Datasets depend on: Sources
- Webhooks depend on: Workspaces

## Technical Debt & Improvements

### 🔧 Code Quality
- [ ] Add comprehensive error handling patterns
- [ ] Implement consistent logging throughout
- [ ] Add configuration validation helpers
- [ ] Optimize API client connection pooling
- [ ] Add request/response debugging tools

### 📖 Documentation  
- [ ] Add inline code documentation
- [ ] Create architecture decision records (ADRs)
- [ ] Expand troubleshooting guides
- [ ] Add video tutorials and walkthroughs

### 🚀 Developer Experience
- [ ] Improve local development workflow
- [ ] Add hot-reload for development
- [ ] Create provider debugging tools
- [ ] Enhance error messages and validation

## Long-term Vision

### **Advanced Features** (Future Releases)
- [ ] Multi-workspace resource management
- [ ] Cross-workspace data sharing
- [ ] Advanced scheduling and orchestration
- [ ] Data lineage and impact analysis
- [ ] Cost optimization and monitoring
- [ ] Integration with other Terraform providers (dbt, Snowflake, etc.)

### **Ecosystem Integration**
- [ ] Terraform Cloud integration
- [ ] CI/CD pipeline templates
- [ ] Monitoring and observability
- [ ] GitOps workflows
- [ ] Policy as code integration

---

## Notes

- **This TODO is tracked in version control** and should be updated with every significant change
- **Priority can shift** based on user feedback and Census API changes
- **Each major feature** should include: implementation, testing, documentation, and examples
- **Breaking changes** should follow semantic versioning and provide migration paths

Last updated: 2025-01-XX