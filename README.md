# gohan

[![CI](https://github.com/bmf-san/gohan/actions/workflows/ci.yml/badge.svg)](https://github.com/bmf-san/gohan/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmf-san/gohan.svg)](https://pkg.go.dev/github.com/bmf-san/gohan)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A simple, fast static site generator written in Go — featuring incremental builds, syntax highlighting, Mermaid diagrams, and a live-reload dev server.

> 日本語版: [README.ja.md](README.ja.md)

<img src="./docs/assets/icon.png" alt="gohan" title="gohan" width="100px">

This logo was created by [gopherize.me](https://gopherize.me/gopher/f64aa0974e77fef33a2c2fe234c8fc478c08d013).

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

| Guide | Description |
|---|---|
| [Getting Started](docs/guide/getting-started.md) | Installation, first site, build & preview |
| [Configuration](docs/guide/configuration.md) | All `config.yaml` fields and Front Matter |
| [Templates](docs/guide/templates.md) | Theme templates, variables, built-in functions |
| [Taxonomy](docs/guide/taxonomy.md) | Tags, categories, and archive pages |
| [CLI Reference](docs/guide/cli.md) | All commands and flags |

---

## Design

For architecture and design decisions see [docs/DESIGN_DOC.md](docs/DESIGN_DOC.md).

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

---

## Sponsor

If you'd like to support my work, please consider sponsoring me!

[GitHub Sponsors – bmf-san](https://github.com/sponsors/bmf-san)

Or simply giving ⭐ on GitHub is greatly appreciated—it keeps me motivated to maintain and improve the project! :D

---

## License

[MIT](LICENSE)
