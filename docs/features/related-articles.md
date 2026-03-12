# Related Articles

gohan automatically populates a list of related articles for each article page, making it easy to add a "Related Articles" section to your theme.

> 日本語版: [related-articles.ja.md](related-articles.ja.md)

---

## How it works

When generating an article page, gohan computes `Site.RelatedArticles` by:

1. Collecting all articles that share **at least one category** with the current article.
2. Filtering to the **same locale** as the current article (so English articles only relate to English articles, Japanese to Japanese, etc.).
3. Excluding the **current article itself**.
4. Sorting the candidates by **date descending** (newest first).
5. Returning at most **5** articles.

`RelatedArticles` is `nil` on all non-article pages (`index.html`, `tag.html`, `category.html`, `archive.html`).

---

## Template usage

`$.RelatedArticles` is available on `article.html`. Because gohan article templates typically use `{{range .Articles}}`, use `$` (the root `Site` value) to access it:

```html
{{if $.RelatedArticles}}
<section>
  <h2>Related Articles</h2>
  <ul>
    {{range $.RelatedArticles}}
    <li>
      <time>{{formatDate "2006-01-02" .FrontMatter.Date}}</time>
      <a href="{{.URL}}">{{.FrontMatter.Title}}</a>
      {{range .FrontMatter.Categories}}
      <span>{{.}}</span>
      {{end}}
    </li>
    {{end}}
  </ul>
</section>
{{end}}
```

### Available fields on each related article

Each item in `$.RelatedArticles` is a `*ProcessedArticle`. Commonly used fields:

| Field | Example value | Description |
|---|---|---|
| `.URL` | `/posts/hello/` | Canonical URL path |
| `.FrontMatter.Title` | `Hello, World!` | Article title |
| `.FrontMatter.Date` | `time.Time` | Publish date |
| `.FrontMatter.Categories` | `[]string{"go"}` | Categories |
| `.FrontMatter.Description` | `A short intro` | Meta description |

---

## Internals

The logic lives in `internal/generator/html.go`:

```go
func relatedArticles(all []*model.ProcessedArticle, a *model.ProcessedArticle, n int) []*model.ProcessedArticle {
    catSet := make(map[string]bool, len(a.FrontMatter.Categories))
    for _, c := range a.FrontMatter.Categories {
        catSet[c] = true
    }
    var related []*model.ProcessedArticle
    for _, candidate := range all {
        if candidate == a || candidate.Locale != a.Locale {
            continue
        }
        for _, c := range candidate.FrontMatter.Categories {
            if catSet[c] {
                related = append(related, candidate)
                break
            }
        }
    }
    sortByDateDesc(related)
    if len(related) > n {
        related = related[:n]
    }
    return related
}
```

It is called once per article page inside `buildJobs()`:

```go
d.RelatedArticles = relatedArticles(site.Articles, a, 5)
```

### Changes required

| File | Change |
|---|---|
| `internal/model/site.go` | Add `RelatedArticles []*ProcessedArticle` to `Site` |
| `internal/generator/html.go` | Add `relatedArticles()` helper; call it in the article job loop |
| `internal/generator/html_test.go` | Add `TestRelatedArticles` unit tests |
