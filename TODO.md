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

✅ **Phase 2 Complete: Core Census Operations** 
- **Major Architectural Achievement: PAT-only Authentication** - Eliminated workspace token dependency
- Dynamic workspace token retrieval using personal access tokens
- Complete source and destination management
- API field validation against Census connector schemas
- State management fixes for workspace_id persistence

✅ **Phase 3 Complete: Sync Operations**
- **Complete Census Workflow Available** - Full end-to-end data sync capability
- census_sync resource with comprehensive configuration options
- Field mappings, scheduling, sync modes (upsert, append, mirror)
- Source/destination attributes for table, dataset, model, topic sources
- API-compliant table source structure with proper table_name/table_schema/table_catalog
- Working examples connecting warehouses to CRM systems
- OpenAPI specification compliance for all sync operations

## Recently Completed Major Features

### ✅ census_destination Resource (FULLY IMPLEMENTED)
- ✅ Complete destination API client (`internal/client/destination.go`)
- ✅ Full destination resource implementation with CRUD operations
- ✅ Support for all connector types (Salesforce, HubSpot, etc.) via dynamic API validation
- ✅ Connection credential management with real-time validation
- ✅ Connection testing and field validation against `/connectors` API
- ✅ Destination-specific configuration schemas from Census API
- ✅ Auto-refresh metadata after creation
- ✅ Complete working examples with real credentials

### ✅ census_sync Resource (FULLY IMPLEMENTED)
- ✅ Complete sync API client (`internal/client/sync.go`)
- ✅ Full sync resource implementation with CRUD operations
- ✅ Support for all source types (table, dataset, model, topic, segment, cohort)
- ✅ Field mapping configuration with direct, hash, and constant operations
- ✅ Sync scheduling with hourly, daily, weekly, and manual modes
- ✅ Sync mode support (upsert, append, mirror)
- ✅ Dynamic workspace token authentication for all operations
- ✅ OpenAPI-compliant source_attributes with proper table schema
- ✅ Working examples with Salesforce CRM integration
- ✅ Complete sync data source implementation

### ✅ census_source Resource (FULLY IMPLEMENTED)
- ✅ Complete source API client (`internal/client/source.go`)
- ✅ Full source resource implementation with CRUD operations
- ✅ Database connection support (Postgres, Snowflake, BigQuery, etc.)
- ✅ Connection credential management with validation
- ✅ Source validation against `/source_types` API
- ✅ Auto table refresh functionality
- ✅ State management fixes for workspace_id persistence
- ✅ Import support for existing resources

### ✅ Advanced Technical Features
- ✅ **Dynamic Token Authentication**: PAT → Workspace Token conversion
- ✅ **API Schema Validation**: Real-time validation against Census connector requirements
- ✅ **State Persistence**: Fixed workspace_id state management issues
- ✅ **Pagination Support**: Proper API pagination handling
- ✅ **Error Handling**: Comprehensive error handling with helpful messages
- ✅ **Complete Example Setup**: Working terraform.tfvars with Salesforce integration

## Next Development Phases

### 🚧 Phase 3: Sync Operations (Implementation Complete, Testing Required)

#### 🧪 Sync Resources (IMPLEMENTED - AWAITING TESTING)
- ✅ **census_sync Resource** - Create and manage Census data syncs
  - ✅ Complete sync API client methods (`internal/client/sync.go`)
  - ✅ Full sync resource implementation (`internal/provider/resource_sync.go`)
  - ✅ Complete sync data source (`internal/provider/data_source_sync.go`)
  - ✅ Comprehensive sync configuration schema (source, destination, field mappings)
  - ✅ Full sync scheduling and trigger options (hourly, daily, weekly, manual)
  - ✅ Sync status monitoring and management
  - ✅ Complete working example configurations
  - ✅ OpenAPI specification compliance for all operations

### 📋 Phase 4: Data & Execution Management (Medium Priority)

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

### 👥 Phase 5: Organization & User Management (Lower Priority)

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

### 🧪 Phase 6: Testing & Quality Assurance

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

### 📚 Phase 7: Documentation & Publishing

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

### 🔧 Phase 8: Production Features

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
1. **Integration Testing** - Real API tests with complete Census workflow
2. **census_dataset Resource** - Advanced data modeling and SQL query support
3. **Documentation Updates** - Reflect new sync capabilities and OpenAPI compliance
4. **Terraform Registry Preparation** - Ready for public release with full workflow
5. **Performance Testing** - Test sync creation and execution at scale

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

Last updated: 2025-01-17

## Major Milestones Achieved
- **2025-01-16**: Completed source and destination resources with full API validation
- **2025-01-17**: Implemented PAT-only authentication architecture
- **2025-01-17**: Fixed state management and workspace_id persistence
- **2025-01-17**: Added comprehensive connector validation via Census API
- **2025-01-17**: Completed census_sync resource with OpenAPI-compliant table sources - Full Census workflow now available!
- **2025-01-17**: Fixed table source structure to use proper table_name/table_schema/table_catalog per OpenAPI spec