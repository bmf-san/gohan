# Plugin System

## Overview

gohan ships with a **built-in plugin system** that allows optional features to be enabled per-project via `config.yaml` without requiring users to write Go code.

Plugins are compiled into the gohan binary. Enabling or disabling a plugin is a configuration change only — no recompilation is needed by the end user.

## Architecture

```
cmd/gohan/build.go
  └── plugin.DefaultRegistry().Enrich(site)   ← called between Process() and Generate()
        └── for each enabled plugin:
              plugin.TemplateData(article, cfg) → stored in article.PluginData["<name>"]
```

Template access pattern:
```html
{{with index .PluginData "amazon_books"}}
  {{range .books}}
    <a href="{{.LinkURL}}">{{.Title}}</a>
  {{end}}
{{end}}
```

## Plugin Interface

Defined in `internal/plugin/plugin.go`:

```go
type Plugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    TemplateData(article *model.ProcessedArticle, cfg map[string]interface{}) (map[string]interface{}, error)
}
```

- **`Name()`** — unique key used in `config.yaml` under `plugins.<name>` and as the key in `ProcessedArticle.PluginData`
- **`Enabled()`** — receives the plugin's config sub-map; controls whether the plugin runs
- **`TemplateData()`** — returns arbitrary data exposed to the theme template

## FrontMatter Extension

Plugins read per-article data from `FrontMatter.Extra`, which captures all unknown YAML keys via `yaml:",inline"`:

```yaml
---
title: My Article
tags: [go]
# Plugin-specific keys:
books:
  - asin: "4873119464"
    title: "入門 Go"
---
```

## Built-in Plugins

| Plugin | Package | Purpose |
|--------|---------|-----|
| `amazon_books` | `internal/plugin/amazonbooks` | Amazon book cards with affiliate tracking |

#### amazon_books

Generates book card data (image URL, product URL, title) from ASIN values declared in the article's front-matter.

**config.yaml:**
```yaml
plugins:
  amazon_books:
    enabled: true
    tag: "your-associate-tag-22"
```

**Article front-matter:**
```yaml
books:
  - asin: "4873119464"
    title: "入門 Go"  # optional; used for alt text
```

**Template data shape:**
```
.PluginData["amazon_books"].books → []BookCard
  BookCard.ASIN      string
  BookCard.Title     string
  BookCard.ImageURL  string   # images-na.ssl-images-amazon.com CDN
  BookCard.LinkURL   string   # amazon.co.jp/dp/{ASIN}?tag={tag}
```

## Adding a New Plugin

1. Create `internal/plugin/<name>/<name>.go` implementing `plugin.Plugin`
2. Add a compile-time interface check: `var _ plugin.Plugin = (*MyPlugin)(nil)`
3. Register in `internal/plugin/registry.go` → `DefaultRegistry()`
4. Document in this section

---

## SitePlugin — Cross-article page generation

While `Plugin` operates on individual articles, **`SitePlugin`** operates on the full site and generates **VirtualPages** — pages with no corresponding Markdown source file.

```
cmd/gohan/build.go
  └── plugin.DefaultRegistry().EnrichVirtual(site)  ← called after Enrich()
        └── for each enabled SitePlugin:
              SitePlugin.VirtualPages(site, cfg) → appended to site.VirtualPages
                                                   ↓
                                    HTMLGenerator.buildJobs() renders them
```

Template access pattern (the page-specific data is at `.VirtualPageData`):
```html
{{range index .VirtualPageData "books"}}
  <a href="{{.LinkURL}}" target="_blank" rel="noopener">
    <img src="{{.ImageURL}}" alt="{{.Title}}">
  </a>
  {{if .ArticleURL}}<a href="{{.ArticleURL}}">{{.ArticleTitle}}</a>{{end}}
{{end}}
```

### SitePlugin Interface

Defined in `internal/plugin/plugin.go`:

```go
type SitePlugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    VirtualPages(site *model.Site, cfg map[string]interface{}) ([]*model.VirtualPage, error)
}
```

- **`Name()`** — unique key used in `config.yaml` under `plugins.<name>`
- **`Enabled()`** — controls whether the plugin runs
- **`VirtualPages()`** — inspects the full site and returns zero or more `VirtualPage` values

### VirtualPage fields

```
VirtualPage.OutputPath  string   // file path relative to output dir, e.g. "bookshelf/index.html"
VirtualPage.URL         string   // canonical URL path, e.g. "/bookshelf/"
VirtualPage.Template    string   // theme template filename, e.g. "bookshelf.html"
VirtualPage.Locale      string   // locale code, e.g. "en" or "ja"
VirtualPage.Data        map[string]interface{}  // exposed as .VirtualPageData in templates
```

### Built-in SitePlugins

#### bookshelf

Aggregates all book entries from every article's `books:` front-matter and generates one bookshelf page per locale.

**config.yaml:**
```yaml
plugins:
  bookshelf:
    enabled: true
    tag: "your-associate-tag-22"   # Amazon Associates tracking tag
```

**Generated URLs:**
- Default locale (en): `/bookshelf/`
- Non-default locale (ja): `/ja/bookshelf/`

**Template data shape** (`.VirtualPageData`):**
```
.VirtualPageData["books"] → []BookEntry
  BookEntry.ASIN          string
  BookEntry.Title         string
  BookEntry.ImageURL      string   # images-na.ssl-images-amazon.com CDN
  BookEntry.LinkURL       string   # amazon.co.jp/dp/{ASIN}?tag={tag}
  BookEntry.ArticleSlug   string   # slug of the source article (book review)
  BookEntry.ArticleTitle  string   # title of the source article
  BookEntry.ArticleURL    string   # canonical URL of the source article
  BookEntry.Date          time.Time
```

Entries are sorted by date descending (newest first).

### Adding a New SitePlugin

1. Create `internal/plugin/<name>/<name>.go` implementing `plugin.SitePlugin`
2. Add a compile-time interface check: `var _ plugin.SitePlugin = (*MyPlugin)(nil)`
3. Register in `internal/plugin/registry.go` → `DefaultRegistry()` under `sitePlugins`
4. Create a theme template that reads `.VirtualPageData`
5. Document in this section

## Scope

- Dynamic plugin loading (`plugin` package) is intentionally out of scope — it adds OS constraints and complexity that are unnecessary for a static site generator
- Plugins do not generate HTML; they supply data to themes, keeping UI fully under the theme's control
- Per-article data is read-only from the plugin's perspective
- VirtualPages are not included in Atom feeds or the index article listing
