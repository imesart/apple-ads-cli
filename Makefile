BINARY := aads
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
TARGET_API_VERSION := $(shell python3 -c 'import json; from pathlib import Path; print(json.loads(Path("docs/apple_ads/openapi-latest.json").read_text())["info"]["version"])')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.targetAPIVersion=$(TARGET_API_VERSION)"

GO := go
GOBIN := $(shell $(GO) env GOPATH)/bin
GOLANGCI_LINT_TIMEOUT ?= 5m

BLUE := \033[38;2;1;165;178m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

.DEFAULT_GOAL := help

.PHONY: help build install install-git-hooks test test-integration test-live lint vet fmt format format-check commands-doc commands-doc-check openapi openapi-check schema-index schema-index-check release-artifacts clean

help:
	@echo ""
	@echo "$(GREEN)aads$(NC) build targets"
	@echo ""
	@echo "  build         Build the CLI into bin/$(BINARY)"
	@echo "  install       Install the CLI with go install"
	@echo "  install-git-hooks  Configure git to use the repo's tracked hooks"
	@echo "  test          Run the Go test suite"
	@echo "  test-integration  Run mocked command-integration coverage in ./cmd"
	@echo "  test-live     Run live command-integration coverage in ./cmd with AADS_INTEGRATION_TEST=1"
	@echo "  lint          Run golangci-lint, or fall back to go vet"
	@echo "  vet           Run go vet"
	@echo "  format        Format Go code with gofmt"
	@echo "  format-check  Check whether formatting changes are needed"
	@echo "  commands-doc  Generate CLI command reference docs"
	@echo "  commands-doc-check  Check whether docs/commands.md is up to date"
	@echo "  openapi       Generate Apple Ads OpenAPI docs data"
	@echo "  openapi-check  Check whether Apple Ads OpenAPI docs data is up to date"
	@echo "  schema-index  Generate embedded API schema index"
	@echo "  schema-index-check  Check whether embedded API schema index is up to date"
	@echo "  release-artifacts  Build release archives into release/"
	@echo "  clean         Remove build artifacts"
	@echo ""

build:
	@echo "$(BLUE)Building $(BINARY)...$(NC)"
	@mkdir -p bin
	$(GO) build $(LDFLAGS) -o bin/$(BINARY) .
	@echo "$(GREEN)Build complete: bin/$(BINARY)$(NC)"

install:
	@echo "$(BLUE)Installing $(BINARY)...$(NC)"
	$(GO) install $(LDFLAGS) .
	@echo "$(GREEN)Install complete$(NC)"

install-git-hooks:
	@echo "$(BLUE)Installing tracked git hooks...$(NC)"
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit
	@echo "$(GREEN)Git hooks installed from .githooks/$(NC)"

test:
	@echo "$(BLUE)Running tests...$(NC)"
	$(GO) test ./...

test-integration:
	@echo "$(BLUE)Running mocked command-integration tests...$(NC)"
	$(GO) test ./cmd -run '^TestIntegration_'

test-live:
	@echo "$(BLUE)Running live command-integration tests...$(NC)"
	AADS_INTEGRATION_TEST=1 $(GO) test ./cmd -run '^TestLive_' -count=1

lint:
	@echo "$(BLUE)Linting code...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=$(GOLANGCI_LINT_TIMEOUT) ./...; \
	elif [ -x "$$HOME/go/bin/golangci-lint" ]; then \
		"$$HOME/go/bin/golangci-lint" run --timeout=$(GOLANGCI_LINT_TIMEOUT) ./...; \
	else \
		echo "$(YELLOW)golangci-lint not found; falling back to '$(GO) vet ./...'.$(NC)"; \
		echo "$(YELLOW)Install with: $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
		$(GO) vet ./...; \
	fi

vet:
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GO) vet ./...

fmt: format

format:
	@echo "$(BLUE)Formatting Go files...$(NC)"
	gofmt -w .
	@echo "$(GREEN)Formatting complete$(NC)"

format-check:
	@echo "$(BLUE)Checking formatting...$(NC)"
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "$(YELLOW)Formatting issues detected:$(NC)"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "$(GREEN)Formatting is clean$(NC)"

commands-doc: build
	@echo "$(BLUE)Generating command reference docs...$(NC)"
	python3 scripts/generate_commands_doc.py docs/commands.md
	@echo "$(GREEN)Command reference updated$(NC)"

commands-doc-check: build
	@echo "$(BLUE)Checking command reference docs...$(NC)"
	python3 scripts/generate_commands_doc.py docs/commands.md --check
	@echo "$(GREEN)Command reference docs are up to date$(NC)"

openapi:
	@echo "$(BLUE)Generating Apple Ads OpenAPI docs...$(NC)"
	python3 scripts/generate_openapi.py docs/apple_ads/openapi.json --output-paths docs/apple_ads/paths.txt --yes
	@echo "$(GREEN)OpenAPI docs updated$(NC)"
	$(MAKE) schema-index

openapi-check:
	@echo "$(BLUE)Checking Apple Ads OpenAPI docs...$(NC)"
	python3 scripts/generate_openapi.py docs/apple_ads/openapi.json --output-paths docs/apple_ads/paths.txt --check
	@echo "$(GREEN)OpenAPI docs are up to date$(NC)"
	$(MAKE) schema-index-check

schema-index:
	@echo "$(BLUE)Generating embedded API schema index...$(NC)"
	python3 scripts/generate_schema_index.py docs/apple_ads/openapi-latest.json internal/cli/schema/schema_index.json
	@echo "$(GREEN)Schema index updated$(NC)"

schema-index-check:
	@echo "$(BLUE)Checking embedded API schema index...$(NC)"
	python3 scripts/generate_schema_index.py docs/apple_ads/openapi-latest.json internal/cli/schema/schema_index.json --check
	@echo "$(GREEN)Schema index is up to date$(NC)"

release-artifacts:
	@echo "$(BLUE)Building release archives for $(VERSION)...$(NC)"
	python3 scripts/create_release_artifacts.py --version "$(VERSION)" --commit "$(COMMIT)" --date "$(DATE)" --output-dir release --source-dir . --binary-name "$(BINARY)"
	@echo "$(GREEN)Release archives complete: release/$(NC)"

clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf bin/ release/ .gocache/ scripts/__pycache__/
	@echo "$(GREEN)Clean complete$(NC)"
