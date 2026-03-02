# OGP Image Generation

## Overview

gohan generates OGP (Open Graph Protocol) thumbnail images at build time using pure Go. Each article gets a unique `og:image` derived from its title, eliminating the need for manual image creation.

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

Add an `ogp` block to `config.yaml`:

```yaml
ogp:
  enabled: true
  background_color: "#1e1e2e"
  text_color: "#cdd6f4"
  font_file: "assets/fonts/NotoSansJP-Bold.ttf"   # TTF/OTF, required for CJK
  logo_file: "assets/images/logo.png"              # optional overlay
  width: 1200
  height: 630
```

The font file must be bundled by the user. Any TTF/OTF font is supported.

## Data Model

Add `OGPConfig` to `model.go`:

```go
// OGPConfig holds settings for build-time OGP image generation.
type OGPConfig struct {
    Enabled         bool   `yaml:"enabled"`
    BackgroundColor string `yaml:"background_color"`
    TextColor       string `yaml:"text_color"`
    FontFile        string `yaml:"font_file"`
    LogoFile        string `yaml:"logo_file"` // empty means no logo
    Width           int    `yaml:"width"`
    Height          int    `yaml:"height"`
}
```

Add to `Config`:

```go
type Config struct {
    // ... existing fields ...
    OGP OGPConfig `yaml:"ogp"`
}
```

## Implementation

**`internal/generator/ogp.go`** (new file)
- Implement `OGPGenerator` satisfying `OutputGenerator`
- Use stdlib `image`, `image/color`, `image/png`, `image/draw`
- Use `golang.org/x/image/font` and `golang.org/x/image/font/opentype` for TrueType rendering
- Use `golang.org/x/image/math/fixed` for fixed-point arithmetic
- Rendering pipeline: fill background → draw logo (if configured) → draw word-wrapped title text centered vertically and horizontally
- Skip generation if `ogp.enabled: false`
- Skip per-article generation if the output `.png` already exists and the source article is unchanged (cache-aware via `ChangeSet`)

**`internal/generator/generator.go`**
- Invoke `OGPGenerator.Generate()` as part of the build pipeline when `cfg.OGP.Enabled` is true

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
| `golang.org/x/image/font` | Font face interface and text drawing |
| `golang.org/x/image/font/opentype` | Load TTF/OTF font files |
| `golang.org/x/image/math/fixed` | Fixed-point arithmetic for font rendering |

`golang.org/x/image` is already transitively pulled in by many Go projects and adds negligible build overhead.
