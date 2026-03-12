# gohan

[![GitHub release](https://img.shields.io/github/release/bmf-san/gohan.svg)](https://github.com/bmf-san/gohan/releases)
[![CI](https://github.com/bmf-san/gohan/actions/workflows/ci.yml/badge.svg)](https://github.com/bmf-san/gohan/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/bmf-san/gohan/branch/main/graph/badge.svg)](https://codecov.io/gh/bmf-san/gohan)
[![CodeQL](https://github.com/bmf-san/gohan/actions/workflows/codeql.yml/badge.svg)](https://github.com/bmf-san/gohan/actions/workflows/codeql.yml)
[![Dependabot Updates](https://github.com/bmf-san/gohan/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/bmf-san/gohan/actions/workflows/dependabot/dependabot-updates)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmf-san/gohan)](https://goreportcard.com/report/github.com/bmf-san/gohan)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmf-san/gohan.svg)](https://pkg.go.dev/github.com/bmf-san/gohan)
[![Sourcegraph](https://sourcegraph.com/github.com/bmf-san/gohan/-/badge.svg)](https://sourcegraph.com/github.com/bmf-san/gohan?badge)
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
- **Mermaid diagrams** — Fenced `mermaid` blocks render as interactive diagrams
- **Taxonomy** — Tag and category pages generated automatically
- **Atom feed & sitemap** — `atom.xml` and `sitemap.xml` generated automatically
- **Live-reload dev server** — `gohan serve` watches files and reloads the browser
- **Customisable themes** — Full control via Go `html/template`
- **Plugin system** — Built-in plugins enabled per-project via `config.yaml` (no Go code required)
- **i18n** — Multi-locale content with per-article translation links and `hreflang` support
- **OGP image generation** — Build-time `1200×630` Open Graph images, one per article
- **Pagination** — Configurable `per_page` with automatic next/previous page links
- **GitHub source link** — Per-article link to the source file on GitHub for easy editing
- **Related articles** — Automatic same-category article recommendations on article pages

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
gohan new --title="Hello, World!" hello-world

# 4. Build the site
gohan build

# 5. Preview locally with live reload
gohan serve   # open http://127.0.0.1:1313
```

---

## Plugins

Plugins are compiled into gohan and toggled via `config.yaml`. No Go code is required to use them.

### amazon_books

Generates Amazon book card data (cover image, product URL, title) from ASIN values in an article's front-matter. Designed for affiliate link integration.

**config.yaml:**
```yaml
plugins:
  amazon_books:
    enabled: true
    tag: "your-associate-tag-22"   # Amazon Associates tracking tag
```

**Article front-matter:**
```yaml
books:
  - asin: "4873119464"
    title: "Learning Go"   # optional; used for alt text
```

**Template usage** (in your theme's `article.html`):
```html
{{with index .PluginData "amazon_books"}}
{{if .books}}
<section class="book-cards">
  {{range .books}}
  <a href="{{.LinkURL}}" target="_blank" rel="noopener">
    <img src="{{.ImageURL}}" alt="{{.Title}}">
    <span>{{.Title}}</span>
  </a>
  {{end}}
</section>
{{end}}
{{end}}
```

See [docs/DESIGN_DOC.md §20](docs/DESIGN_DOC.md) for the full plugin architecture.

---

## User Guide

| Guide | Description |
|---|---|
| [Getting Started](docs/guide/getting-started.md) | Installation, first site, build & preview |
| [Configuration](docs/guide/configuration.md) | All `config.yaml` fields and Front Matter |
| [Templates](docs/guide/templates.md) | Theme templates, variables, built-in functions |
| [Taxonomy](docs/guide/taxonomy.md) | Tags, categories, and archive pages |
| [CLI Reference](docs/guide/cli.md) | All commands and flags |

## Feature Documentation

| Feature | Description |
|---|---|
| [i18n](docs/features/i18n.md) | Multi-locale content with translation links and `hreflang` |
| [OGP Image Generation](docs/features/ogp.md) | Build-time Open Graph images per article |
| [Pagination](docs/features/pagination.md) | Configurable `per_page` and page navigation |
| [GitHub Source Link](docs/features/github-source-link.md) | Per-article link to the source file on GitHub |
| [Plugin System](docs/features/plugin-system.md) | Built-in plugins (`amazon_books`, …) via `config.yaml` |
| [Related Articles](docs/features/related-articles.md) | Same-category article recommendations on article pages |

---

## Design

For architecture and design decisions see [docs/DESIGN_DOC.md](docs/DESIGN_DOC.md).

---

## Sites Built with gohan

| Site | Description |
|---|---|
| [bmf-tech.com](https://bmf-tech.com) ([source](https://github.com/bmf-san/bmf-tech)) | Personal tech blog — i18n (EN/JA), 700+ articles, Cloudflare Pages |

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
