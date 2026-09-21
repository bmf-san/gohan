---
title: "Plugin System"
description: "Extend gohan with external plugins."
slug: "plugin-system"
categories:
  - features
translation_key: "plugin-system"
---

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

**Article front-matter:**
```yaml
books:
  - asin: "4873119464"          # Amazon ASIN — generates cover image + Amazon link
    title: "入門 Go"
  - url: "https://booth.pm/..."  # non-Amazon: direct sales URL (no cover image)
    title: "My Zine"
```

When `url:` is provided instead of `asin:`, `ImageURL` is empty and `LinkURL` is the direct URL.

**Generated URLs:**
- Default locale (en): `/bookshelf/`
- Non-default locale (ja): `/ja/bookshelf/`

**Template data shape** (`.VirtualPageData`):**
```
.VirtualPageData["books"] → []BookEntry
  BookEntry.ASIN          string
  BookEntry.Title         string
  BookEntry.ImageURL      string   # images-na.ssl-images-amazon.com CDN; empty for url:-only entries
  BookEntry.LinkURL       string   # amazon.co.jp/dp/{ASIN}?tag={tag}, or the direct url: value
  BookEntry.ArticleSlug   string   # slug of the source article (book review)
  BookEntry.ArticleTitle  string   # title of the source article
  BookEntry.ArticleURL    string   # canonical URL of the source article
  BookEntry.Categories    []string # categories of the source article
  BookEntry.Date          time.Time

.VirtualPageData["categories"] → []CategoryGroup  # entries grouped by category
  CategoryGroup.Name      string       # category name; empty string = uncategorised
  CategoryGroup.Books     []BookEntry
```

`books` is sorted by date descending (newest first). `categories` is sorted alphabetically by category name; uncategorised entries appear last.

**Template example (category grouping):**
```html
{{range index .VirtualPageData "categories"}}
  <h2>{{if .Name}}{{.Name}}{{else}}Uncategorised{{end}}</h2>
  {{range .Books}}
    <a href="{{.LinkURL}}" target="_blank" rel="noopener">
      {{if .ImageURL}}<img src="{{.ImageURL}}" alt="{{.Title}}">{{else}}{{.Title}}{{end}}
    </a>
  {{end}}
{{end}}
```

### Adding a New SitePlugin

1. Create `internal/plugin/<name>/<name>.go` implementing `plugin.SitePlugin`
2. Add a compile-time interface check: `var _ plugin.SitePlugin = (*MyPlugin)(nil)`
3. Register in `internal/plugin/registry.go` → `DefaultRegistry()` under `sitePlugins`
4. Create a theme template that reads `.VirtualPageData`
5. Document in this section

---

## AssetPlugin — Build-time binary assets

While `Plugin` supplies template data and `SitePlugin` generates HTML pages, **`AssetPlugin`** generates **build-time binary artifacts** (such as OGP images) written directly into the output directory. Templates reference these files by a conventional URL; gohan core treats their contents as opaque.

```
cmd/gohan/build.go
  └── plugin.DefaultRegistry().GenerateAssets(site, outDir, changeSet)  ← the "assets" build phase (after render)
        └── for each enabled AssetPlugin:
              AssetPlugin.GenerateAssets(site, outDir, changeSet, cfg) → writes files under outDir
```

### AssetPlugin Interface

Defined in `internal/plugin/plugin.go`:

```go
type AssetPlugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    GenerateAssets(site *model.Site, outDir string, changeSet *model.ChangeSet, cfg map[string]interface{}) error
}
```

- **`Name()`** — unique key used in `config.yaml` under `plugins.<name>`
- **`Enabled()`** — controls whether the plugin runs
- **`GenerateAssets()`** — writes artifacts into `outDir`. `changeSet` is `nil` for full builds, or the set of changed files for incremental builds so the plugin can skip unchanged outputs

### Built-in AssetPlugins

#### ogp

Generates one Open Graph (OGP) image per article at `<output>/ogp/{slug}.png`, using a deterministic slug-seeded gradient design. See [OGP](ogp.md) for details.

**config.yaml:**
```yaml
plugins:
  ogp:
    enabled: true
    logo_file: "assets/images/logo.png"   # optional
    width: 1200
    height: 630
```

### Adding a New AssetPlugin

1. Create `internal/plugin/<name>/<name>.go` implementing `plugin.AssetPlugin`
2. Add a compile-time interface check (structural, or `var _ plugin.AssetPlugin = (*MyPlugin)(nil)`)
3. Register in `internal/plugin/registry.go` → `DefaultRegistry()` under `assetPlugins`
4. Reference the generated files from a theme template
5. Document in this section

## Scope

- Dynamic plugin loading (`plugin` package) is intentionally out of scope — it adds OS constraints and complexity that are unnecessary for a static site generator
- Plugins do not generate HTML; they supply data to themes, keeping UI fully under the theme's control
- Per-article data is read-only from the plugin's perspective
- VirtualPages are not included in Atom feeds or the index article listing
