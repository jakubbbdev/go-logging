.PHONY: build test clean examples install fmt lint docs

# Build the library
build:
	go build -o bin/go-logging ./pkg/logging

# Run tests
test:
	go test -v ./pkg/logging

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./pkg/logging
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test -bench=. -benchmem ./pkg/logging

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf logs/
	rm -f *.log
	rm -f coverage.out
	rm -f coverage.html

# Install the library
install:
	go install ./pkg/logging

# Run examples
examples:
	@echo "Running basic example..."
	go run cmd/examples/basic/main.go

# Run web example
web-example:
	@echo "Running web example..."
	go run cmd/examples/web/main.go

# Format code
fmt:
	go fmt ./pkg/logging ./cmd/examples

# Lint code
lint:
	golangci-lint run ./pkg/logging ./cmd/examples

# Generate documentation
docs:
	godoc -http=:6060

# Create logs directory
logs-dir:
	mkdir -p logs

# Run all checks
check: fmt lint test

# Build and test everything
all: clean build test examples

# Show project structure
tree:
	@echo "Project Structure:"
	@echo "go-logging/"
	@echo "├── pkg/logging/          # Main library package"
	@echo "│   ├── logging.go        # Package entry point"
	@echo "│   ├── logger.go         # Core logger interface and implementation"
	@echo "│   ├── handlers.go       # Console, file, and multi handlers"
	@echo "│   ├── formatters.go     # Text and JSON formatters"
	@echo "│   ├── context.go        # Context support and utilities"
	@echo "│   └── logger_test.go    # Test files"
	@echo "├── cmd/examples/         # Example applications"
	@echo "│   ├── basic/            # Basic usage examples"
	@echo "│   └── web/              # Web server examples"
	@echo "├── docs/                 # Documentation"
	@echo "├── Makefile              # Build and development tools"
	@echo "├── LICENSE               # MIT License"
	@echo "└── README.md             # Project documentation" 