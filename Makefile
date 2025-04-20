# --- Variables ---

BINARY_NAME=phase4
CMD_PATH=.


# --- Go Targets ---

.PHONY: build
build:
## Build the Go binary locally
	@echo "Building $(BINARY_NAME) ..."
	@mkdir -p bin # Ensure bin directory exists
	@go build -v -o bin/$(BINARY_NAME) $(CMD_PATH)

.PHONY: test
test:
## Run Go tests
	@echo "Running tests ..."
	@go test ./...

.PHONY: cover
cover:
## Run Go tests with coverage report
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@echo "To view coverage: go tool cover -html=coverage.out"

.PHONY: lint
lint:
## Run golangci-lint (install if needed: https://golangci-lint.run/usage/install/)
	@echo "Running linter ..."
	@golangci-lint run ./... || echo "Linter found issues."

.PHONY: clean
clean:
## Clean up build artifacts
	@echo "Cleaning ..."
	@rm -f bin/$(BINARY_NAME)
	@rm -f coverage.out


# --- Help Target ---

.PHONY: help
help:
## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
