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

build:
  content_dir: "content"
  output_dir: "public"
  assets_dir: "assets"
  exclude_files:
    - "*.draft.md"
    - "_*"
  parallelism: 4

theme:
  name: "default"
  dir: "themes/default"
  params:
    primary_color: "#0066cc"
    footer_text: "© 2024 My Blog"

syntax_highlight:
  theme: "github"
  line_numbers: false
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

> `base_url` is used to generate absolute URLs in `sitemap.xml` and `atom.xml`. Do not include a trailing slash.

---

## `build` section

File paths and build behaviour.

| Field | Type | Default | Description |
|---|---|---|---|
| `content_dir` | string | `"content"` | Markdown content directory (relative to project root) |
| `output_dir` | string | `"public"` | HTML output directory |
| `assets_dir` | string | `"assets"` | Static files directory (CSS, images, etc.) |
| `exclude_files` | []string | `[]` | Glob patterns for files to exclude from the build |
| `parallelism` | int | `4` | Number of parallel HTML generation workers |

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

### Automatic slug generation

When `slug` is omitted it is derived from `title`:

- Spaces → hyphens
- Uppercase → lowercase
- Example: `"Hello World"` → `hello-world`
