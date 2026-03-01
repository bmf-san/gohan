# Getting Started

> This guide walks you through installing gohan, creating your first site, and previewing it locally.

> 日本語版: [getting-started.ja.md](getting-started.ja.md)

---

## Prerequisites

- Go 1.21 or later
- Git (required for incremental builds)

---

## Installation

```bash
go install github.com/bmf-san/gohan/cmd/gohan@latest
```

Verify the installation:

```bash
gohan version
# gohan v1.0.0 (commit: abc1234, built: 2024-01-01T00:00:00Z)
```

### Build from source

```bash
git clone https://github.com/bmf-san/gohan.git
cd gohan
make install
```

### Download a binary

Download a pre-built binary for your platform from [GitHub Releases](https://github.com/bmf-san/gohan/releases).

---

## Create Your First Site

### Step 1: Create a project directory

```bash
mkdir myblog
cd myblog
```

### Step 2: Create `config.yaml`

```yaml
site:
  title: My Blog
  description: A simple personal blog
  base_url: https://myblog.example.com
  language: en

build:
  content_dir: content
  output_dir: public
  assets_dir: assets
  parallelism: 4

theme:
  name: default

syntax_highlight:
  theme: github
  line_numbers: false
```

See [Configuration](configuration.md) for all available fields.

### Step 3: Create your first article

```bash
gohan new post --slug=hello-world --title="Hello, World!"
```

This creates `content/posts/hello-world.md`. Edit it to add body content:

```markdown
---
title: Hello, World!
date: 2024-01-15
slug: hello-world
tags:
  - go
  - blog
categories:
  - tech
draft: false
description: My first gohan post
---

# Hello, World!

Welcome to my blog powered by **gohan**!

## Features

- Incremental builds for fast generation
- Customizable themes with Go html/template
- Mermaid diagrams and syntax highlighting

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, gohan!")
}
```
```

### Step 4: Create theme templates

```bash
mkdir -p themes/default/templates
```

`themes/default/templates/index.html`:

```html
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{{.Config.Site.Description}}">
  <title>{{.Config.Site.Title}}</title>
  <link rel="stylesheet" href="/assets/style.css">
  <link rel="alternate" type="application/atom+xml" href="/atom.xml">
</head>
<body>
  <header>
    <h1><a href="/">{{.Config.Site.Title}}</a></h1>
  </header>
  <main>
    <ul>
      {{range .Articles}}
      <li>
        <span>{{formatDate "2006-01-02" .FrontMatter.Date}}</span>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

`themes/default/templates/article.html`:

```html
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>{{(index .Articles 0).FrontMatter.Title}} — {{.Config.Site.Title}}</title>
  <link rel="stylesheet" href="/assets/style.css">
</head>
<body>
  <header><a href="/">← {{.Config.Site.Title}}</a></header>
  <main>
    {{with (index .Articles 0)}}
    <article>
      <h1>{{.FrontMatter.Title}}</h1>
      <time>{{formatDate "2006-01-02" .FrontMatter.Date}}</time>
      {{if .FrontMatter.Tags}}
      <ul class="tags">
        {{range .FrontMatter.Tags}}
        <li><a href="{{tagURL .}}">{{.}}</a></li>
        {{end}}
      </ul>
      {{end}}
      <div class="content">{{.HTMLContent}}</div>
    </article>
    {{end}}
  </main>
</body>
</html>
```

See [Templates](templates.md) for the complete template reference.

### Step 5: Add static assets (optional)

```bash
mkdir -p assets
cat > assets/style.css << 'EOF'
body { font-family: sans-serif; max-width: 800px; margin: 0 auto; padding: 1rem; }
a { color: #0066cc; }
EOF
```

### Step 6: Build the site

```bash
gohan build
```

The site is generated in `public/`:

```
public/
├── index.html
├── sitemap.xml
├── atom.xml
├── posts/
│   └── hello-world/
│       └── index.html
└── assets/
    └── style.css
```

### Step 7: Preview with the development server

```bash
gohan serve
# Open http://127.0.0.1:1313
```

The browser reloads automatically whenever you save a file.

---

## Common Operations

### Draft articles

Articles with `draft: true` are excluded from builds:

```yaml
---
title: Work in progress
draft: true
---
```

### Incremental builds

After the first build, subsequent builds are automatically incremental (only changed files are regenerated):

```bash
gohan build        # First run: full build
# edit content/
gohan build        # Second run: only changed files
gohan build --full # Force a full rebuild
```

### Recommended `.gitignore`

```gitignore
public/
.gohan/
gohan
```

---

## Next Steps

- [Configuration](configuration.md) — all `config.yaml` options
- [Templates](templates.md) — customize your theme
- [Taxonomy](taxonomy.md) — manage tags and categories
