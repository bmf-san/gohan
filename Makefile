# gohan Makefile

# ─── Version info injected at build time ─────────────────────────────────────
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS  = -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# ─── Tool versions ───────────────────────────────────────────────────────────
GOLANGCI_LINT_VERSION ?= v2.10.1
GOVULNCHECK_VERSION   ?= latest

# ─── Paths ────────────────────────────────────────────────────────────────────
BIN      = gohan
CMD      = ./cmd/gohan
COVERAGE = coverage.out

.PHONY: all build test lint serve clean install coverage tidy vuln tools help

## all: build the binary (default target)
all: build

## build: compile the gohan binary with version ldflags
build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

## install: install gohan to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" $(CMD)

## test: run all tests with race detector and coverage collection
test:
	go test -race -coverprofile=$(COVERAGE) -covermode=atomic ./...

## coverage: print coverage summary (run 'make test' first)
coverage:
	@go tool cover -func=$(COVERAGE) | grep total

## tools: install development tools (golangci-lint, govulncheck)
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## tidy: verify go.mod and go.sum are tidy
tidy:
	go mod tidy
	git diff --exit-code go.mod go.sum

## vuln: run govulncheck
vuln:
	govulncheck ./...

## check: run all checks locally (build test lint tidy vuln)
check: build test lint tidy vuln

## serve: start the development server (requires config.yaml in current directory)
serve: build
	./$(BIN) serve

## clean: remove build outputs and cache
clean:
	rm -f $(BIN)
	rm -f $(COVERAGE)
	rm -rf dist/
	rm -rf .gohan/

## help: print this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
