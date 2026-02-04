.PHONY: build test lint install clean integration-test

BINARY_NAME=ios-agent
BUILD_DIR=./bin
MAIN_PATH=./main.go

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run integration tests (requires simulator)
integration-test:
	@echo "Running integration tests..."
	go test -v -tags=integration ./test/...

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Lint code
lint:
	@echo "Linting..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting..."
	go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run the CLI (for development)
run:
	go run $(MAIN_PATH) $(ARGS)

# Generate mocks
mocks:
	@echo "Generating mocks..."
	@which mockgen > /dev/null || go install go.uber.org/mock/mockgen@latest
	go generate ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  test             - Run unit tests"
	@echo "  integration-test - Run integration tests (requires simulator)"
	@echo "  install          - Install to /usr/local/bin"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  clean            - Clean build artifacts"
	@echo "  deps             - Download dependencies"
	@echo "  run ARGS=...     - Run the CLI"
	@echo "  mocks            - Generate mocks"
