.PHONY: build test lint clean examples help

# Default target
help:
	@echo "Available targets:"
	@echo "  build             - Build all packages"
	@echo "  test              - Run unit tests"
	@echo "  test-integration  - Run integration tests (requires AWS credentials)"
	@echo "  lint              - Run linting"
	@echo "  clean             - Clean build artifacts"
	@echo "  examples          - Build example applications"
	@echo "  mocks             - Generate test mocks"
	@echo "  deps              - Download dependencies"
	@echo "  fmt               - Format code"
	@echo "  vet               - Run go vet"

# Build all packages
build:
	@echo "Building packages..."
	go build ./...

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v -race ./... -short

# Run integration tests (requires AWS credentials)
test-integration:
	@echo "Running integration tests..."
	@echo "Note: This requires valid AWS credentials and may incur AWS charges"
	go test -v -tags=integration ./...

# Run all tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linting
lint:
	@echo "Running linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean ./...
	rm -rf dist/
	rm -f coverage.out coverage.html

# Build example applications
examples:
	@echo "Building examples..."
	go build -o dist/bedrock-chat ./cmd/examples/bedrock-chat
	go build -o dist/monitoring-demo ./cmd/examples/monitoring-demo
	@echo "Examples built in dist/ directory"

# Generate test mocks
mocks:
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		go generate ./...; \
	else \
		echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"; \
		exit 1; \
	fi

# Run a complete check (format, vet, lint, test)
check: fmt vet lint test

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest

# Run examples locally (requires AWS credentials)
run-bedrock-example:
	@echo "Running Bedrock chat example..."
	@echo "Note: This requires valid AWS credentials"
	cd cmd/examples/bedrock-chat && go run main.go "What is artificial intelligence?"

run-monitoring-example:
	@echo "Running monitoring demo..."
	@echo "Note: This requires valid AWS credentials and will send metrics to CloudWatch"
	cd cmd/examples/monitoring-demo && go run main.go

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t genkit-aws:latest .

docker-test:
	@echo "Running tests in Docker..."
	docker run --rm -v $(PWD):/app -w /app golang:1.23-alpine go test -v ./...

# Release preparation
pre-release: clean fmt vet lint test examples
	@echo "Pre-release checks completed successfully"

# Display project info
info:
	@echo "GenKit AWS Plugins"
	@echo "=================="
	@echo "Go version: $$(go version)"
	@echo "Module: $$(go list -m)"
	@echo "Dependencies:"
	@go list -m all | head -10