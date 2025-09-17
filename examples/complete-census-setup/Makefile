# Census Terraform Provider - Complete Setup Example
# Usage: make setup && terraform plan

PROVIDER_BINARY = terraform-provider-census
ROOT_DIR = ../..

.PHONY: setup clean plan apply validate

# One-command setup: build provider, copy binary, and start debug mode
setup:
	@echo "🧹 Cleaning up any existing processes..."
	@pkill -f "$(PROVIDER_BINARY) -debug" || true
	@rm -f .tf_reattach_providers.txt $(PROVIDER_BINARY)
	@echo "🔨 Building provider..."
	@cd $(ROOT_DIR) && go build -o bin/$(PROVIDER_BINARY)
	@echo "📋 Copying provider binary..."
	@cp $(ROOT_DIR)/bin/$(PROVIDER_BINARY) ./$(PROVIDER_BINARY)
	@echo "🚀 Starting provider in debug mode..."
	@echo "⏳ Please wait while provider starts..."
	@./$(PROVIDER_BINARY) -debug > .tf_reattach_providers.txt 2>&1 &
	@sleep 5
	@if grep -q "TF_REATTACH_PROVIDERS=" .tf_reattach_providers.txt 2>/dev/null; then \
		echo "✅ Setup complete! Provider is running in debug mode."; \
		echo ""; \
		echo "🔗 Provider connection ready!"; \
		echo ""; \
		echo "Copy and run this export command:"; \
		grep "TF_REATTACH_PROVIDERS=" .tf_reattach_providers.txt | head -1 | sed 's/^[[:space:]]*/export /' ; \
		echo ""; \
		echo "Then use regular terraform commands:"; \
		echo "  terraform validate"; \
		echo "  terraform plan"; \
		echo "  terraform apply"; \
	else \
		echo "⚠️  Provider is still starting up..."; \
		echo "Wait a moment and check:"; \
		echo "  cat .tf_reattach_providers.txt"; \
		echo "  then run 'make setup' again if needed"; \
	fi

# Note: Use the export command from 'make setup' output, then run terraform commands directly

# Initialize terraform (only needed once)
init:
	@echo "🔧 Initializing terraform..."
	@terraform init

# Clean up debug files and stop provider processes
clean:
	@echo "🧹 Cleaning up..."
	@pkill -f "$(PROVIDER_BINARY) -debug" || true
	@rm -f .tf_reattach_providers.txt $(PROVIDER_BINARY)
	@echo "✅ Cleanup complete."

# Show help
help:
	@echo "Census Terraform Provider - Complete Setup Example"
	@echo ""
	@echo "Usage:"
	@echo "  make setup     - Build provider and start in debug mode (one-command setup)"
	@echo "  make init      - Initialize terraform (run once)"
	@echo "  make clean     - Stop provider and cleanup files"
	@echo "  make help      - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make setup"
	@echo "  2. Copy and run the export command from setup output"
	@echo "  3. Edit terraform.tfvars with your credentials"  
	@echo "  4. terraform validate"
	@echo "  5. terraform plan"
	@echo "  6. terraform apply"