# Pagination

## Overview

gohan generates paginated listing pages using a path-based URL scheme, consistent with the conventions of major SSGs (Hugo, Jekyll, etc.) and compatible with static hosting (Cloudflare Pages, GitHub Pages).

## URL Scheme

```
/              → Page 1 (public/index.html)
/page/2/       → Page 2 (public/page/2/index.html)
/page/3/       → Page 3 (public/page/3/index.html)
```

- Page 1 is always served from the root `index.html`; no `/page/1/` alias is generated
- Tag and category listing pages follow the same pattern:
  - `/tags/go/` → Page 1
  - `/tags/go/page/2/` → Page 2
  - `/categories/architecture/` → Page 1
  - `/categories/architecture/page/2/` → Page 2

## Configuration

`per_page` is added to `BuildConfig` in `config.yaml`:

```yaml
build:
  per_page: 10   # number of articles per page. 0 or omitted disables pagination
```

## Data Model

Add `Pagination` struct and `PerPage` field to `model.go`:

```go
// Pagination holds computed paging metadata for listing pages.
type Pagination struct {
    CurrentPage int
    TotalPages  int
    PerPage     int
    TotalItems  int
    PrevURL     string // empty string if no previous page
    NextURL     string // empty string if no next page
    BaseURL     string // URL path prefix used to construct PrevURL/NextURL (e.g. "/tags/go")
}
```

`Site` gains a `Pagination` field:

```go
type Site struct {
    // ... existing fields ...
    Pagination *Pagination // nil when pagination is disabled or page is not a listing page
}
```

`BuildConfig` gains `PerPage`:

```go
type BuildConfig struct {
    // ... existing fields ...
    PerPage int `yaml:"per_page"`
}
```

## Implementation

**`internal/model/model.go`**
- Add `Pagination` struct
- Add `Pagination *Pagination` to `Site`
- Add `PerPage int` to `BuildConfig`

**`internal/generator/html.go`**
- `buildJobs()` splits articles into pages of size `cfg.Build.PerPage`
- Page 1 → `{outDir}/index.html`
- Page N (N ≥ 2) → `{outDir}/page/N/index.html`
- Add `siteWithPagination(site, articles, pagination)` helper that returns a shallow copy of `Site` with the given article slice and `Pagination`
- Apply the same split logic to tag and category listing pages

**Template usage (user-side)**:

```html
{{if .Pagination}}
  {{if .Pagination.PrevURL}}
    <link rel="prev" href="{{.Pagination.PrevURL}}">
  {{end}}
  {{if .Pagination.NextURL}}
    <link rel="next" href="{{.Pagination.NextURL}}">
  {{end}}

  {{if .Pagination.PrevURL}}
    <a href="{{.Pagination.PrevURL}}">← Prev</a>
  {{end}}
  <span>{{.Pagination.CurrentPage}} / {{.Pagination.TotalPages}}</span>
  {{if .Pagination.NextURL}}
    <a href="{{.Pagination.NextURL}}">Next →</a>
  {{end}}
{{end}}
```

## SEO Considerations

- `<link rel="prev">` / `<link rel="next">` allows Google to discover and follow the pagination chain
- Each paginated page has its own canonical URL (`/page/2/`, etc.)
- No redirect is required when migrating from query-parameter pagination (`?page=2`), since paginated listing pages are rarely indexed or externally linked
