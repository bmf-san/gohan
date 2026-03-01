# gohan User Guide

Official documentation for using, configuring, and customizing gohan.

> 日本語版: [README.ja.md](README.ja.md)

---

## Guides

| Document | Description |
|---|---|
| [Getting Started](getting-started.md) | Install gohan, create your first site, build & preview |
| [Configuration](configuration.md) | Full `config.yaml` field reference and Front Matter |
| [Templates](templates.md) | Create themes, template variables, and built-in functions |
| [Taxonomy](taxonomy.md) | Manage tags, categories, and archive pages |

---

## Quick Reference

### Common commands

```bash
# Create a new post
gohan new post --slug=my-post --title="My Post Title"

# Build the site
gohan build

# Force a full rebuild (skip diff detection)
gohan build --full

# Simulate a build without writing any files
gohan build --dry-run

# Start the development server (http://127.0.0.1:1313)
gohan serve

# Print version information
gohan version
```

### Makefile targets

```bash
make build     # Compile the gohan binary
make test      # Run all tests with the race detector
make coverage  # Run tests and report coverage
make lint      # Run golangci-lint
make clean     # Remove build artifacts
make help      # List available targets
```

---

## Design Document

For gohan's internal architecture and design decisions, see [DESIGN_DOC.md](../DESIGN_DOC.md).
