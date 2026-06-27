---
title: "Search Index"
description: "Generate a JSON search index for client-side search."
slug: "search-index"
categories:
  - features
translation_key: "search-index"
---

gohan generates a `search-index.json` at build time so you can add client-side article search to your theme without any external tooling or services.

> 日本語版: [search-index.ja.md](search-index.ja.md)

---

## How it works

During the build, gohan writes a `search-index.json` alongside `sitemap.xml` and `atom.xml` in the "feeds" phase. It is generated automatically — no configuration is required.

Each entry holds **metadata only** (no full body text), so the index stays small even on sites with hundreds of articles:

| Field | Example | Description |
|---|---|---|
| `title` | `Hello, World!` | Article title |
| `url` | `https://example.com/posts/hello/` | Absolute URL (same rule as feeds) |
| `description` | `A short intro` | Meta description |
| `summary` | `An auto-generated excerpt…` | Article summary |
| `tags` | `["go", "ssg"]` | Tags (omitted when empty) |
| `categories` | `["features"]` | Categories (omitted when empty) |
| `date` | `2026-03-14T00:00:00Z` | Publish date (RFC 3339) |
| `locale` | `en` | Locale code (present only with i18n) |

Articles are sorted **newest-first**. Draft and future-dated articles are excluded unless you pass `--draft` or `--future`.

### Output shape

```json
{
  "generated": "2026-06-27T04:57:01Z",
  "count": 2,
  "articles": [
    {
      "title": "i18n",
      "url": "https://bmf-san.github.io/gohan/features/i18n/",
      "description": "Build multi-language sites with directory-based locales.",
      "summary": "## Overview",
      "categories": ["features"],
      "locale": "en"
    }
  ]
}
```

---

## i18n

When `i18n.locales` is configured, the index is split per locale — the same convention used by feeds and sitemap:

- The root `search-index.json` contains **default-locale** articles only.
- Each non-default locale gets its own `{locale}/search-index.json` (e.g. `ja/search-index.json`).

Each localized page can then load only the entries for its own language.

---

## Theme usage

The index is a static JSON file, so a few lines of vanilla JavaScript are enough — no dependencies required:

```html
<input id="q" type="search" placeholder="Search…">
<ul id="results"></ul>

<script>
  const out = document.getElementById("results");
  fetch("/search-index.json")
    .then((r) => r.json())
    .then(({ articles }) => {
      document.getElementById("q").addEventListener("input", (e) => {
        const q = e.target.value.toLowerCase().trim();
        out.replaceChildren();
        if (!q) return;
        for (const a of articles) {
          const hay = [a.title, a.description, (a.tags || []).join(" ")]
            .join(" ")
            .toLowerCase();
          if (!hay.includes(q)) continue;
          const li = document.createElement("li");
          const link = document.createElement("a");
          link.href = a.url;          // set as a property — avoids HTML injection
          link.textContent = a.title; // textContent escapes the title safely
          li.appendChild(link);
          out.appendChild(li);
        }
      });
    });
</script>
```

For fuzzy matching or ranking, pass `articles` to a small client-side library such as [MiniSearch](https://github.com/lucaong/minisearch) or [Fuse.js](https://www.fusejs.io/).

> When your site is served from a sub-path (e.g. GitHub Pages at `/gohan`), prefix the fetch URL with your `base_path`.

---

## Internals

The logic lives in `internal/generator/searchindex.go`, mirroring how feeds and sitemap are generated.

| File | Change |
|---|---|
| `internal/generator/searchindex.go` | `GenerateSearchIndex`: writes metadata-only entries, newest-first, with an i18n-aware split |
| `cmd/gohan/build.go` | Calls `GenerateSearchIndex` in the build "feeds" phase |
| `internal/generator/searchindex_test.go` | Unit tests: valid output, empty site, field mapping, i18n split |
