BINARY_NAME=terraform-provider-census
VERSION=0.1.0
BUILD_DIR=bin
LDFLAGS=-ldflags "-X main.version=${VERSION}"

.PHONY: build test clean install fmt vet lint help

help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the provider binary
	@echo "Building ${BINARY_NAME} v${VERSION}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} .

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./... -short

test-integration: ## Run integration tests (requires mock server)
	@echo "Running integration tests..."
	@echo "Start mock server first: go run scripts/mock_server.go"
	@go test -v ./internal/client -run TestWorkspaceIntegration

test-acc: ## Run acceptance tests
	@echo "Running acceptance tests..."
	@TF_ACC=1 go test -v ./... -timeout 30m

test-all: test test-integration ## Run unit and integration tests

mock-server: ## Start mock Census API server
	@echo "Starting mock Census API server on :8080..."
	@go run scripts/mock_server.go

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf ${BUILD_DIR}
	@go clean

install: build ## Install the provider locally
	@echo "Installing ${BINARY_NAME} locally..."
	@go install ${LDFLAGS} .

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docs: ## Generate documentation
	@echo "Generating documentation..."
	@terraform fmt -recursive examples/

release: clean test build ## Build release binaries
	@echo "Building release binaries..."
	@GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_${VERSION}_linux_amd64 .
	@GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_${VERSION}_darwin_amd64 .
	@GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_${VERSION}_darwin_arm64 .
	@GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_${VERSION}_windows_amd64.exe .

dev: build ## Build and set up for local development
	@echo "Setting up for local development..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/your-org/census/${VERSION}/$(shell go env GOOS)_$(shell go env GOARCH)/
	@cp ${BUILD_DIR}/${BINARY_NAME} ~/.terraform.d/plugins/registry.terraform.io/your-org/census/${VERSION}/$(shell go env GOOS)_$(shell go env GOARCH)/

.DEFAULT_GOAL := help