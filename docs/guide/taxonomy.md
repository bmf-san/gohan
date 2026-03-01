# Taxonomy Guide

Taxonomy is the system for classifying articles by **tags** and **categories**.
gohan automatically generates a dedicated page for each tag and category.

> 日本語版: [taxonomy.ja.md](taxonomy.ja.md)

---

## Setting up taxonomy in articles

Specify tags and categories in the Front Matter:

```markdown
---
title: Understanding Go Concurrency
date: 2024-03-01
slug: go-concurrency
tags:
  - go
  - concurrency
  - goroutine
categories:
  - tech
  - programming
---
```

- `tags` — Article keywords. Multiple values allowed. One page is generated per tag.
- `categories` — Article classification. Multiple values allowed. One page is generated per category.

---

## Generated pages

Given the article above, the following pages are generated:

```
public/
├── tags/
│   ├── go/index.html
│   ├── concurrency/index.html
│   └── goroutine/index.html
└── categories/
    ├── tech/index.html
    └── programming/index.html
```

---

## Accessing taxonomy in templates

### Tag page (`tag.html`)

`.Articles` contains only the articles that have this tag:

```html
<!-- themes/default/templates/tag.html -->
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>Tag Articles — {{.Config.Site.Title}}</title>
</head>
<body>
  <header><nav><a href="/">← {{.Config.Site.Title}}</a></nav></header>
  <main>
    <h2>Articles ({{len .Articles}})</h2>
    <ul>
      {{range .Articles}}
      <li>
        <time>{{formatDate "2006-01-02" .FrontMatter.Date}}</time>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

### Category page (`category.html`)

`.Articles` contains only the articles that belong to this category:

```html
<!-- themes/default/templates/category.html -->
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>Category Articles — {{.Config.Site.Title}}</title>
</head>
<body>
  <main>
    <h2>Articles ({{len .Articles}})</h2>
    <ul>
      {{range .Articles}}
      <li>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

### Listing all tags and categories

Use `.Tags` and `.Categories` to generate dynamic links to all tags/categories:

```html
<!-- e.g. in index.html -->
<section>
  <h3>Tags</h3>
  <ul>
    {{range .Tags}}
    <li><a href="{{tagURL .Name}}">{{.Name}}</a></li>
    {{end}}
  </ul>

  <h3>Categories</h3>
  <ul>
    {{range .Categories}}
    <li><a href="{{categoryURL .Name}}">{{.Name}}</a></li>
    {{end}}
  </ul>
</section>
```

---

## URL generation

Tag and category names are normalized to URL slugs:

- Spaces → hyphens
- Uppercase → lowercase
- Example: `"Go Language"` → `/tags/go-language/`

Template functions:

```html
<a href="{{tagURL "Go Language"}}">Go Language</a>
<!-- → <a href="/tags/go-language/">Go Language</a> -->

<a href="{{categoryURL "Web Development"}}">Web Development</a>
<!-- → <a href="/categories/web-development/">Web Development</a> -->
```

---

## Archive pages

Year-based archive pages are generated automatically from the article `date` field:

```
public/
└── archive/
    ├── 2024/index.html
    └── 2023/index.html
```

In `archive.html`, `.Articles` contains the articles published in that year:

```html
<!-- themes/default/templates/archive.html -->
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>Archive — {{.Config.Site.Title}}</title>
</head>
<body>
  <main>
    <h2>Archive ({{len .Articles}} articles)</h2>
    <ul>
      {{range .Articles}}
      <li>
        <time>{{formatDate "2006-01-02" .FrontMatter.Date}}</time>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

---

## Taxonomy design guidelines

- **Tags** — Specific keywords for the article (`go`, `docker`, `postgresql`, etc.). Having many tags is fine.
- **Categories** — Broad classifications (`tech`, `life`, `book`, etc.). Keep them few (5–10 recommended).

---

## Related pages

- [Template Guide](templates.md)
- [Configuration](configuration.md)
