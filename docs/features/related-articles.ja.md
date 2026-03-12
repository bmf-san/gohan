# 関連記事

gohan は各記事ページに自動的に関連記事リストを生成します。テーマに「関連記事」セクションを追加するのが簡単になります。

> English version: [related-articles.md](related-articles.md)

---

## 動作の仕組み

記事ページを生成する際、gohan は以下のロジックで `Site.RelatedArticles` を算出します:

1. 現在の記事と **1 つ以上のカテゴリーを共有** する記事を収集する。
2. 現在の記事と **同じロケール** に絞り込む（英語記事は英語記事のみ、日本語は日本語のみ）。
3. **現在の記事自身** を除外する。
4. **日付の降順**（新しい順）に並べ替える。
5. 最大 **5 件** を返す。

`RelatedArticles` は記事ページ以外（`index.html`・`tag.html`・`category.html`・`archive.html`）では `nil` になります。

---

## テンプレートでの使用

`$.RelatedArticles` は `article.html` で利用できます。gohan の記事テンプレートは通常 `{{range .Articles}}` を使うため、ルートの `Site` 値である `$` を使ってアクセスしてください。

```html
{{if $.RelatedArticles}}
<section>
  <h2>関連記事</h2>
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

### 関連記事の各フィールド

`$.RelatedArticles` の各要素は `*ProcessedArticle` です。よく使うフィールド:

| フィールド | 値の例 | 説明 |
|---|---|---|
| `.URL` | `/posts/hello/` | 正規 URL パス |
| `.FrontMatter.Title` | `Hello, World!` | 記事タイトル |
| `.FrontMatter.Date` | `time.Time` | 公開日 |
| `.FrontMatter.Categories` | `[]string{"go"}` | カテゴリー |
| `.FrontMatter.Description` | `概要文` | meta description |

---

## 実装の詳細

ロジックは `internal/generator/html.go` の `relatedArticles()` 関数にあります:

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

記事ページのジョブループ内（`buildJobs()`）で 1 記事ごとに呼び出されます:

```go
d.RelatedArticles = relatedArticles(site.Articles, a, 5)
```

### 変更ファイル一覧

| ファイル | 変更内容 |
|---|---|
| `internal/model/site.go` | `Site` に `RelatedArticles []*ProcessedArticle` を追加 |
| `internal/generator/html.go` | `relatedArticles()` ヘルパー追加・記事ジョブループで呼び出し |
| `internal/generator/html_test.go` | `TestRelatedArticles` ユニットテスト追加 |
