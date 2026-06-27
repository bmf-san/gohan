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
- **Project scaffolding** — `gohan init` bootstraps `config.yaml`, content folders, and archetype templates
- **Content linter** — `gohan check` validates duplicate slugs, missing front matter, and orphan translation keys
- **Archetype templates** — `gohan new --archetype=<name>` renders custom front-matter skeletons from `archetypes/<name>.md`
- **Markdown + Front Matter** — GitHub Flavored Markdown with YAML metadata
- **TOC / WordCount / ReadingTime** — Auto-derived per article and exposed to templates
- **Scheduled posts** — Future-dated articles are skipped by default; opt in with `gohan build --future`
- **Build observability** — `--stats` prints per-phase timing; `--explain` shows what triggered a rebuild
- **Syntax highlighting** — Code blocks styled with [chroma](https://github.com/alecthomas/chroma)
- **Mermaid diagrams** — Fenced `mermaid` blocks render as interactive diagrams
- **Taxonomy** — Tag and category pages generated automatically
- **Atom feed & sitemap** — `atom.xml` and `sitemap.xml` generated automatically
- **Search index** — `search-index.json` generated automatically (per locale) for client-side search
- **Live-reload dev server** — `gohan serve` watches files and reloads the browser (CSS-only changes hot-swap stylesheets without a full reload)
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
# 1. Scaffold a new project (creates config.yaml, content/, archetypes/, README.md)
gohan init myblog && cd myblog

# 2. Create your first article (uses archetypes/post.md)
gohan new --title="Hello, World!" hello-world

# 3. Validate the content (optional but recommended in CI)
gohan check

# 4. Build the site
gohan build

# 5. Preview locally with live reload
gohan serve   # open http://127.0.0.1:1313
```

See [Configuration](https://bmf-san.github.io/gohan/guide/configuration/) for all `config.yaml` options.

---

## Documentation

Full documentation site: **https://bmf-san.github.io/gohan/**

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
