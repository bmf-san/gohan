# Configuration Reference

`config.yaml` is the single configuration file placed at the project root.

> 日本語版: [configuration.ja.md](configuration.ja.md)

---

## Complete example

```yaml
site:
  title: "My Blog"
  description: "A personal tech blog"
  base_url: "https://myblog.example.com"
  language: "en"
  github_repo: "https://github.com/owner/repo"  # optional: enables "Edit this page" links
  github_branch: "main"                          # optional: branch used for edit links (default: main)

build:
  content_dir: "content"
  output_dir: "public"
  assets_dir: "assets"
  static_dir: "static"   # optional: copied verbatim to output_dir root
  exclude_files:
    - "*.draft.md"
    - "_*"
  parallelism: 4
  per_page: 20           # optional: articles per paginated listing page (0 = no pagination)

theme:
  name: "default"
  dir: "themes/default"
  params:
    primary_color: "#0066cc"
    footer_text: "© 2024 My Blog"

syntax_highlight:
  theme: "github"
  line_numbers: false

ogp:
  enabled: false         # optional: generate OGP images at build time
  logo_file: ""          # optional: path to logo file (relative to project root)
  width: 1200
  height: 630

i18n:
  locales: [en, ja]      # optional: ordered locale codes; empty = single-language mode
  default_locale: en     # optional: locale served at root URL (default: site.language)

plugins:               # optional: plugin configuration (key = plugin name)
  amazon_books: {}
```

---

## `site` section

Site-wide metadata.

| Field | Type | Default | Description |
|---|---|---|---|
| `title` | string | *(required)* | Site title. Available in templates as `.Config.Site.Title` |
| `description` | string | `""` | Site description. Used in meta tags and the Atom feed |
| `base_url` | string | *(required)* | Base URL without a trailing slash (e.g. `https://example.com`) |
| `language` | string | `"en"` | BCP 47 language code used in `<html lang="">` |
| `github_repo` | string | `""` | GitHub repository base URL (e.g. `https://github.com/owner/repo`). When set, templates can render an "Edit this page" link using `.ContentPath` |
| `github_branch` | string | `"main"` | Branch used to build the edit URL |

> `base_url` is used to generate absolute URLs in `sitemap.xml` and `atom.xml`. Do not include a trailing slash.

---

## `build` section

File paths and build behaviour.

| Field | Type | Default | Description |
|---|---|---|---|
| `content_dir` | string | `"content"` | Markdown content directory (relative to project root) |
| `output_dir` | string | `"public"` | HTML output directory |
| `assets_dir` | string | `"assets"` | Processed assets directory (CSS, images, etc.) |
| `static_dir` | string | `""` | Static files directory copied verbatim to output root (e.g. `static/404.html` → `public/404.html`) |
| `exclude_files` | []string | `[]` | Glob patterns for files to exclude from the build |
| `parallelism` | int | `4` | Number of parallel HTML generation workers |
| `per_page` | int | `0` | Articles per paginated listing page. `0` disables pagination |

### `exclude_files` examples

```yaml
build:
  exclude_files:
    - "*.draft.md"   # Exclude files ending in .draft.md
    - "_*"           # Exclude files starting with _
    - "templates/*"  # Exclude everything under templates/
```

---

## `theme` section

Active theme and custom parameters.

| Field | Type | Default | Description |
|---|---|---|---|
| `name` | string | `"default"` | Theme name. Used to resolve `dir` when `dir` is not set |
| `dir` | string | `"themes/<name>"` | Theme directory path (relative to project root) |
| `params` | map[string]string | `{}` | Arbitrary parameters accessible in templates as `.Config.Theme.Params.<key>` |

### Accessing params in templates

```html
<style>
  :root { --primary: {{.Config.Theme.Params.primary_color}}; }
</style>
<footer>{{.Config.Theme.Params.footer_text}}</footer>
```

### Theme directory layout

```
themes/
└── <name>/
    └── templates/      ← Place template files here
        ├── index.html
        ├── article.html
        ├── tag.html
        ├── category.html
        └── archive.html
```

---

## `syntax_highlight` section

Code-block syntax highlighting powered by [chroma](https://github.com/alecthomas/chroma).

| Field | Type | Default | Description |
|---|---|---|---|
| `theme` | string | `"github"` | chroma colour theme name |
| `line_numbers` | bool | `false` | Show line numbers in code blocks |

### Available themes

| Theme | Style |
|---|---|
| `github` | Light theme (default) |
| `monokai` | Dark background, vivid colours |
| `dracula` | Dark theme |
| `solarized-dark` | Solarized dark |
| `solarized-light` | Solarized light |
| `nord` | Nordic dark theme |
| `vs` | Visual Studio style |
| `pygments` | Python pygments style |

Browse all themes: https://xyproto.github.io/splash/docs/

### Disable highlighting

```yaml
syntax_highlight:
  theme: ""  # Empty string disables highlighting
```

---

## Front Matter reference

Every Markdown file begins with a YAML Front Matter block.

```yaml
---
title: "Article title"       # required: Article title
date: 2024-01-15             # required: Publication date (YYYY-MM-DD)
lastmod: 2026-03-15              # optional: Last-reviewed date. When set, overrides date in sitemap.xml <lastmod> and JSON-LD dateModified
slug: "my-post"              # optional: URL slug (auto-generated from title if omitted)
draft: false                 # optional: Exclude from build when true (default: false)
tags:                        # optional: Tag list
  - go
  - blog
categories:                  # optional: Category list
  - tech
description: "Summary"       # optional: Meta description and feed summary
author: "Your Name"          # optional: Author name
template: "article.html"     # optional: Override the template file
---
```

> **`lastmod`** (`time.Time`) overrides `date` in sitemap `<lastmod>` and JSON-LD `dateModified`. Templates access it via `.FrontMatter.LastMod`.

### Automatic slug generation

When `slug` is omitted it is derived from `title`:

- Spaces → hyphens
- Uppercase → lowercase
- Example: `"Hello World"` → `hello-world`

---

## `ogp` section

Build-time OGP image generation.

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `false` | Generate OGP images during build |
| `logo_file` | string | `""` | Path to a logo file to embed in generated images (relative to project root). Empty = no logo |
| `width` | int | `1200` | Output image width in pixels |
| `height` | int | `630` | Output image height in pixels |

See [docs/features/ogp.md](../features/ogp.md) for the full OGP guide.

---

## `i18n` section

Multi-language site configuration.

| Field | Type | Default | Description |
|---|---|---|---|
| `locales` | []string | `[]` | Ordered locale codes present under the content directory. Empty = single-language mode |
| `default_locale` | string | `site.language` | Locale served at the root URL without a language prefix |

See [docs/features/i18n.md](../features/i18n.md) for the full i18n guide.

---

## `plugins` section

Plugin configuration. Keys are plugin names; values are plugin-specific settings.

```yaml
plugins:
  amazon_books: {}
```

See [docs/features/plugin-system.md](../features/plugin-system.md) for the full plugin guide.
