.PHONY: fmt vet lint test tidy coverage vulncheck ci help

GOFMT ?= gofmt
GOTEST ?= go test
GOVET ?= go vet
GOVULNCHECK ?= govulncheck

GOFMT_FLAGS ?= -w -s
GOTEST_FLAGS ?= -race -shuffle=on -count=1

fmt: ## Format Go source files (excludes vendor)
	@$(GOFMT) $(GOFMT_FLAGS) $$(find . -name '*.go' -not -path './vendor/*')

vet: ## Run go vet on all packages
	@$(GOVET) ./...

lint: vet ## Alias for vet (reserved for future linters)

test: ## Execute unit and integration tests with race/shuffle settings
	@$(GOTEST) $(GOTEST_FLAGS) ./...

tidy: ## Tidy module dependencies
	@go mod tidy

coverage: ## Generate coverage profile and HTML report
	@$(GOTEST) -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

vulncheck: ## Run govulncheck against all packages
	@if ! command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
	echo "govulncheck not installed. Install with 'go install golang.org/x/vuln/cmd/govulncheck@latest'"; \
	exit 1; \
	fi
	@$(GOVULNCHECK) ./...

ci: ## Run vet and test (useful for CI pipelines)
	@$(MAKE) vet
	@$(MAKE) test

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "} {printf "%-10s %s\n", $$1, $$2}'
