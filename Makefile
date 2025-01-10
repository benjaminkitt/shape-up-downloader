.PHONY: build test lint clean

# Build the application
build:
	go build -v ./...

# Run all tests with coverage
test:
	go test -v -race -cover ./...

# Run linting checks
lint:
	golangci-lint run

# Clean build artifacts
clean:
	go clean
	rm -f shape-up-downloader
	rm -f *.epub

# Run all quality checks
check: lint test
