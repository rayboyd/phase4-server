# --- Variables ---

BINARY_NAME=phase4
GOARCH=$(shell go env GOARCH)
CMD_PATH=.


# --- Go Targets ---

.PHONY: build
build:
## Build the Go binary locally
	@echo "Building $(BINARY_NAME) for $(GOARCH) ..."
	@mkdir -p bin # Ensure bin directory exists
	@CGO_ENABLED=1 GOARCH=$(GOARCH) go build -tags=$(GOARCH) -v -o bin/$(BINARY_NAME) $(CMD_PATH)

.PHONY: test
test:
## Run Go tests
	@echo "Running tests ..."
	@go test ./...

.PHONY: race
race:
## Run race condition detection tests
	@echo "Running race detection tests ..."
	@go test -race -timeout 30s ./internal/p4/... -run TestFFT_RaceConditions

.PHONY: cover
cover:
## Run Go tests with coverage report
	@echo "Running tests with coverage ..."
	@go test ./... -coverprofile=coverage.out
	@echo "To view coverage: go tool cover -html=coverage.out"

.PHONY: lint
lint:
## Run golangci-lint (install if needed: https://golangci-lint.run/usage/install/)
	@echo "Running linter ..."
	@golangci-lint run ./... || echo "Linter found issues."

.PHONY: check-struct-align
check-struct-align:
## Check struct field alignment for memory efficiency
	@echo "Checking struct field alignment..."
	@if command -v fieldalignment > /dev/null; then \
		fieldalignment ./... || echo "Struct alignment issues found."; \
	else \
		echo "Installing fieldalignment tool..."; \
		go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest; \
		fieldalignment ./... || echo "Struct alignment issues found."; \
	fi

.PHONY: fix-struct-align
fix-struct-align:
## Auto-fix struct field alignment issues
	@echo "Fixing struct field alignment..."
	@if command -v fieldalignment > /dev/null; then \
		fieldalignment -fix ./... && echo "Structs optimized successfully!"; \
	else \
		echo "Installing fieldalignment tool..."; \
		go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest; \
		fieldalignment -fix ./... && echo "Structs optimized successfully!"; \
	fi

.PHONY: bench
bench:
## Run benchmarks and report memory allocations
	@echo "Running benchmarks with memory profiling ..."
	@go test -bench=. -benchmem ./...

.PHONY: vet
vet:
## Run go vet including struct alignment checks
	@echo "Running go vet with alignment checks..."
	@go vet ./...
	@$(MAKE) check-struct-align

.PHONY: clean
clean:
## Clean up build artifacts
	@echo "Cleaning ..."
	@rm -f bin/$(BINARY_NAME)
	@rm -f coverage.out


# --- Help Target ---

.PHONY: help
help:
## Display help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
