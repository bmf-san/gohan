# Contributing to Gohan

Thank you for your interest in contributing to gohan! This document explains how to get involved.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)
- [Development Setup](#development-setup)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Coding Guidelines](#coding-guidelines)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Reporting Bugs

Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md) issue template. Please include:

- gohan version, OS, and Go version
- Steps to reproduce
- Expected vs. actual behavior
- Relevant config and log output

## Requesting Features

Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md) issue template. Please describe the motivation and your proposed solution.

## Development Setup

**Prerequisites**: Go 1.22 or later, Git

```bash
# Clone the repository
git clone https://github.com/bmf-san/gohan.git
cd gohan

# Download dependencies
go mod download

# Run tests
go test ./...

# Run linter
golangci-lint run
```

## Submitting a Pull Request

1. Fork the repository and create a branch from `main`:
   ```bash
   git checkout -b your-feature-name
   ```
2. Make your changes and add tests where appropriate.
3. Ensure all tests and linting pass:
   ```bash
   go test ./...
   golangci-lint run
   ```
4. Commit using a descriptive message following [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: add support for custom output paths
   fix: correct archive page date grouping
   docs: update CLI reference
   ```
5. Push your branch and open a pull request against `main`.
6. Fill in the pull request template and link any related issues.

## Coding Guidelines

- Follow standard Go conventions (`gofmt`, `go vet`)
- Write godoc comments for all exported identifiers
- Keep functions small and focused; prefer clear naming over comments
- Add or update tests for every change to core logic
- Maintain test coverage at 80% or higher overall, 90%+ for parser/renderer
