.PHONY: fmt fmt-check build run test all

# Build the project
build:
	go build -v ./...

# Run the example in main.go
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Check and fix formatting with gofmt
fmt:
	@echo "Checking code formatting..."
	@GOFMT=$$(gofmt -l .); \
	if [ -z "$$GOFMT" ]; then \
		echo "✓ No formatting issues found"; \
	else \
		echo "Formatting issues found in:"; \
		echo "$$GOFMT"; \
		echo "Fixing formatting issues..."; \
		gofmt -w .; \
		echo "✓ Formatting fixed"; \
	fi

# Check formatting without fixing
fmt-check:
	@echo "Checking code formatting..."
	@GOFMT=$$(gofmt -l .); \
	if [ -z "$$GOFMT" ]; then \
		echo "✓ No formatting issues found"; \
	else \
		echo "Formatting issues found in:"; \
		echo "$$GOFMT"; \
		exit 1; \
	fi

# Update dependencies
deps:
	go mod tidy

# Run all checks
all: deps build fmt-check test 