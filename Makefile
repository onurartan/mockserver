# =========================================
# MockServer Makefile (v1.0.0)
# =========================================

BINARY_NAME := mockserver
CONFIG_DIR := ./example
SRC := .

.PHONY: all build build-release build-all build-npm run run-json run-bm convert-json convert-yaml test test-isole clean

all: build

build: ## Standard build (outputs to ./bin)
	@go run scripts/builder.go -current

build-release: ## Build releases 
	@go run scripts/builder.go -all -out="releases/latest"
    
build-all: ## All builds (outputs to ./bin)
	@go run scripts/builder.go -all -npm

build-npm: ## Build for NPM distribution (outputs to ./npm/bin)
	@go run scripts/builder.go -npm

# Run commands
run: ## Run with default YAML config
	@go run $(SRC) start --config $(CONFIG_DIR)/mockserver.yaml

run-json: ## Run with JSON config
	@go run $(SRC) start --config $(CONFIG_DIR)/mockserver.json

run-bm: ## Run with benchmark config
	@go run $(SRC) start --config ./scripts/mockserver.benchmark.yaml

# Convert commands
convert-json: ## YAML -> JSON
	@go run . convert -i $(CONFIG_DIR)/mockserver.yaml -o $(CONFIG_DIR)/output/mockserver.json

convert-yaml: ## JSON -> YAML 
	@go run . convert -i $(CONFIG_DIR)/mockserver.json -o $(CONFIG_DIR)/output/mockserver.yaml

# Test commands
test: ## Run All Tests
	@echo "Running tests..."
	@go test ./...
	@echo "Tests passed!"

test-isole: ## Run each test in isolation
	@echo "Running tests..."
	@go test -v ./tests ./server/utils ./config
	@echo "Tests passed!"

clean: ## Cleanup binaries and build artifacts
	@go clean
	@if exist bin rmdir /s /q bin 2>nul || rm -rf bin
	@echo "Cleaned up!"