.PHONY: build test clean examples install

# Build the library
build:
	go build -o bin/go-logging .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf logs/
	rm -f *.log
	rm -f coverage.out
	rm -f coverage.html

# Install the library
install:
	go install .

# Run examples
examples:
	@echo "Running basic example..."
	go run examples/basic/main.go

# Run web example
web-example:
	@echo "Running web example..."
	go run examples/web/main.go

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

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