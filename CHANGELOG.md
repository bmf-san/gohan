# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.1.0] - 2025-07-xx

Initial public release of gohan — a simple, fast static site generator written in Go.

### Added

#### Core

- Markdown + Front Matter parsing (GitHub Flavored Markdown via `goldmark`)
- Incremental builds: diff-based change detection regenerates only affected pages
- Live-reload development server (`gohan serve`) with file watch and browser auto-reload
- CLI commands: `gohan build`, `gohan serve`, `gohan new post`, `gohan new page`
- `--draft` flag for `gohan build` to include draft articles in preview builds
- `--full` flag to skip diff detection and regenerate all pages
- `exclude_files` configuration to skip specific source files during build

#### Content

- **Taxonomy**: tag and category pages generated automatically from Front Matter
- **Archive pages**: monthly archives at `/archives/{year}/{month}/`
- **Pagination**: path-based pagination (`/page/2/`) for index, tag, and category listings
- **Atom feed** (`atom.xml`) and **RSS feed** (`feed.xml`) with locale-aware article URLs
- **Sitemap** (`sitemap.xml`) generated automatically
- **OGP image** generation from article metadata
- **Mermaid diagrams**: fenced `mermaid` code blocks rendered as interactive SVG
- **Syntax highlighting**: code blocks styled via [chroma](https://github.com/alecthomas/chroma)
- **Draft filtering**: articles with `draft: true` excluded from production builds
- **Custom templates**: per-article template override via `template` Front Matter field
- **Pages routing**: `content/pages/` articles written to `public/pages/` (not `public/posts/`)

#### i18n

- Multi-language support: locale detected from content path (`content/{locale}/posts/`)
- Per-locale index, tag, category, and archive pages
- Translation linking via `translation_key` Front Matter field
- Language-switcher data (`Translations`) available in templates

#### Plugins

- Built-in plugin system enabled per-project via `config.yaml` (no Go code required)
- `amazon_books` plugin: renders Amazon product cards from ISBNs in Front Matter

#### Themes

- Full template customisation via Go `html/template`
- `themes/<name>/layouts/` with support for base, article, list, and partial templates

#### GitHub Integration

- GitHub source link support: edit/view links generated from `ContentPath` metadata

#### Taxonomy Validation

- `tags.yaml` / `categories.yaml` registry files validate article Front Matter tags and categories at build time

### Changed

- `OutputGenerator` interface reduced to a single `Generate` method; `GenerateSitemap` and `GenerateFeed` exposed as package-level functions only
- Model split into per-concern files (`article.go`, `config.go`, `site.go`, etc.)
- Processor implementation moved to `processor_impl.go` for clarity

### Fixed

- Feed article URLs now use the pre-computed locale-aware `URL` field instead of hardcoded `/posts/{slug}/` paths
- Article output paths use `ProcessedArticle.OutputPath` (calculated by the processor), fixing routing for `content/pages/` articles
- `ValidateArticleTaxonomies` now wired into the build pipeline when taxonomy registry files are present
- Archive generation skips articles with a zero `Date` value

[v0.1.0]: https://github.com/bmf-san/gohan/releases/tag/v0.1.0
