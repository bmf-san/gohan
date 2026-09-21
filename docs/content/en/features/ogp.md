---
title: "OGP"
description: "Auto-generate Open Graph and Twitter Card metadata."
slug: "ogp"
categories:
  - features
translation_key: "ogp"
---

## Overview

gohan generates OGP (Open Graph Protocol) thumbnail images at build time using pure Go, via the built-in `ogp` **asset plugin**. Each article gets a unique `og:image` — a deterministic gradient-and-shape design derived from its slug — eliminating the need for manual image creation.

## Output

```
public/
└── ogp/
    └── {slug}.png    # 1200×630px OGP image per article
```

The `og:image` tag in each article page resolves to:

```
{{.Config.Site.BaseURL}}/ogp/{{.Article.FrontMatter.Slug}}.png
```

Listing pages (index, tag, category) fall back to a user-supplied default image:

```
{{.Config.Site.BaseURL}}/assets/images/ogp-default.png
```

## Configuration

OGP generation is a built-in **asset plugin**. Enable it under `plugins.ogp` in `config.yaml`:

```yaml
plugins:
  ogp:
    enabled: true
    logo_file: "assets/images/logo.png"   # optional top-left logo overlay
    width: 1200                            # optional; default 1200
    height: 630                            # optional; default 630
```

Each image is a deterministic design (a diagonal gradient with geometric decorations) seeded from the article slug, so **no font file is required** and the same article always yields the same image.

| Key | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `false` | Generate OGP images during build |
| `logo_file` | string | `""` | Optional PNG/JPEG logo drawn in the top-left corner |
| `width` | int | `1200` | Image width in pixels |
| `height` | int | `630` | Image height in pixels |

## Implementation

OGP generation lives in the **`internal/plugin/ogp`** package and implements the `plugin.AssetPlugin` interface:

```go
type AssetPlugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    GenerateAssets(site *model.Site, outDir string, changeSet *model.ChangeSet, cfg map[string]interface{}) error
}
```

- Registered in `internal/plugin/registry.go` and invoked by the build pipeline as the `assets` phase (after HTML rendering).
- Reads its configuration from the `plugins.ogp` map — there is no longer a top-level `ogp:` config section.
- Uses stdlib `image`, `image/color`, `image/png`, `image/draw`, plus `golang.org/x/image/draw` for logo scaling.
- Rendering pipeline: slug-seeded diagonal gradient background → geometric accent shapes → optional logo overlay (top-left).
- Skips per-article generation when the output `.png` already exists and the source article is unchanged (incremental via `ChangeSet`).

**Template usage (user-side)**:

```html
<!-- article.html -->
<meta property="og:image"
  content="{{.Config.Site.BaseURL}}/ogp/{{.Article.FrontMatter.Slug}}.png">
<meta name="twitter:image"
  content="{{.Config.Site.BaseURL}}/ogp/{{.Article.FrontMatter.Slug}}.png">

<!-- index.html, tag.html, category.html -->
<meta property="og:image"
  content="{{.Config.Site.BaseURL}}/assets/images/ogp-default.png">
```

## Dependencies

| Package | Purpose |
|---|---|
| `image`, `image/png`, `image/draw` | stdlib — canvas creation and PNG encoding |
| `image/jpeg` | stdlib — decode JPEG logo files |
| `golang.org/x/image/draw` | High-quality scaling for the optional logo overlay |

`golang.org/x/image` adds negligible build overhead.
