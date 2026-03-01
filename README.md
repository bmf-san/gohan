# gohan

[![CI](https://github.com/bmf-san/gohan/actions/workflows/ci.yml/badge.svg)](https://github.com/bmf-san/gohan/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmf-san/gohan.svg)](https://pkg.go.dev/github.com/bmf-san/gohan)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A simple, fast static site generator written in Go — featuring incremental builds, syntax highlighting, Mermaid diagrams, and a live-reload dev server.

> 日本語版: [README.ja.md](README.ja.md)

---

## Features

- **Incremental builds** — Regenerate only changed files, minimising build time
- **Markdown + Front Matter** — GitHub Flavored Markdown with YAML metadata
- **Syntax highlighting** — Code blocks styled with [chroma](https://github.com/alecthomas/chroma)
- **Mermaid diagrams** — Fenced ` + "`" + `mermaid` + "`" + ` blocks render as interactive diagrams
- **Taxonomy** — Tag and category pages generated automatically
- **Atom feed & sitemap** — `atom.xml` and `sitemap.xml` generated automatically
- **Live-reload dev server** — `gohan serve` watches files and reloads the browser
- **Customisable themes** — Full control via Go `html/template`

---

## Installation

```bash
go install github.com/bmf-san/gohan/cmd/gohan@latest
```

Or build from source:

```bash
git clone https://github.com/bmf-san/gohan.git
cd gohan
make install
```

Pre-built binaries are available on [GitHub Releases](https://github.com/bmf-san/gohan/releases).

---

## Quick Start

```bash
# 1. Create a project directory
mkdir myblog && cd myblog

# 2. Add config.yaml (see docs/guide/configuration.md for all options)
cat > config.yaml << 'EOF'
site:
  title: My Blog
  base_url: https://example.com
  language: en
build:
  content_dir: content
  output_dir: public
theme:
  name: default
EOF

# 3. Create your first article
gohan new post --slug=hello-world --title="Hello, World!"

# 4. Build the site
gohan build

# 5. Preview locally with live reload
gohan serve   # open http://127.0.0.1:1313
```

---

## User Guide

For full documentation see **[docs/guide/](docs/guide/README.md)**:

| Guide | Description |
|---|---|
| [Getting Started](docs/guide/getting-started.md) | Installation, first site, build & preview |
| [Configuration](docs/guide/configuration.md) | All `config.yaml` fields and Front Matter |
| [Templates](docs/guide/templates.md) | Theme templates, variables, built-in functions |
| [Taxonomy](docs/guide/taxonomy.md) | Tags, categories, and archive pages |

---

## CLI Reference

| Command | Description |
|---|---|
| `gohan build` | Build the site (incremental by default) |
| `gohan build --full` | Force a full rebuild |
| `gohan build --dry-run` | Simulate a build without writing files |
| `gohan new post --slug=<s> --title=<t>` | Create a new post skeleton |
| `gohan new page --slug=<s> --title=<t>` | Create a new page skeleton |
| `gohan serve` | Start the live-reload development server |
| `gohan version` | Print version information |

---

## For Developers

```bash
make test      # Run all tests with the race detector
make coverage  # Run tests and report coverage percentage
make lint      # Run golangci-lint
make build     # Compile the gohan binary
make clean     # Remove build artifacts
```

For architecture and design decisions see [docs/DESIGN_DOC.md](docs/DESIGN_DOC.md).

---

## License

[MIT](LICENSE)
