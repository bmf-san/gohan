# gohan Makefile

# ─── Version info injected at build time ─────────────────────────────────────
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS  = -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# ─── Paths ────────────────────────────────────────────────────────────────────
BIN      = gohan
CMD      = ./cmd/gohan
COVERAGE = coverage.out

.PHONY: all build test lint serve clean install coverage help

## all: build the binary (default target)
all: build

## build: compile the gohan binary with version ldflags
build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

## install: install gohan to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" $(CMD)

## test: run all tests with the race detector
test:
	go test -race -count=1 ./...

## coverage: run tests and report total coverage percentage
coverage:
	go test -race -coverprofile=$(COVERAGE) -covermode=atomic -count=1 ./...
	@echo ""
	@go tool cover -func=$(COVERAGE) | grep total

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## vet: run go vet
vet:
	go vet ./...

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
