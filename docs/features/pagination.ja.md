# ページネーション

## 概要

gohan は、主要なSSG（Hugo、Jekyllなど）の慣習に準拠し、静的ホスティング（Cloudflare Pages、GitHub Pagesなど）と互換性のあるパスベースURLスキームでページネーション済み一覧ページを生成する。

## URLスキーム

```
/              → 1ページ目 (public/index.html)
/page/2/       → 2ページ目 (public/page/2/index.html)
/page/3/       → 3ページ目 (public/page/3/index.html)
```

- 1ページ目は常にルートの `index.html` から配信される。`/page/1/` エイリアスは生成しない
- タグ・カテゴリ一覧ページも同じパターンに従う：
  - `/tags/go/` → 1ページ目
  - `/tags/go/page/2/` → 2ページ目
  - `/categories/architecture/` → 1ページ目
  - `/categories/architecture/page/2/` → 2ページ目

## 設定

`config.yaml` の `BuildConfig` に `per_page` を追加する：

```yaml
build:
  per_page: 10   # 1ページあたりの記事数。0または省略でページネーション無効
```

## データモデル

`model.go` に `Pagination` 構造体と `PerPage` フィールドを追加する：

```go
// Pagination は一覧ページ用のページング메타情報を保持する。
type Pagination struct {
    CurrentPage int
    TotalPages  int
    PerPage     int
    TotalItems  int
    PrevURL     string // 前のページがない場合は空文字列
    NextURL     string // 次のページがない場合は空文字列
}
```

`Site` に `Pagination` フィールドを追加する：

```go
type Site struct {
    // ... 既存フィールド ...
    Pagination *Pagination // ページネーション無効または一覧ページでない場合はnil
}
```

`BuildConfig` に `PerPage` を追加する：

```go
type BuildConfig struct {
    // ... 既存フィールド ...
    PerPage int `yaml:"per_page"`
}
```

## 実装

**`internal/model/model.go`**
- `Pagination` 構造体を追加
- `Site` に `Pagination *Pagination` を追加
- `BuildConfig` に `PerPage int` を追加

**`internal/generator/html.go`**
- `buildJobs()` が `cfg.Build.PerPage` のサイズで記事を分割
- 1ページ目 → `{outDir}/index.html`
- Nページ目（N ≥ 2）→ `{outDir}/page/N/index.html`
- 指定した記事スライスと `Pagination` を持つ `Site` のシャローコピーを返す `siteWithPagination(site, articles, pagination)` ヘルパーを追加
- タグ・カテゴリ一覧ページにも同じ分割ロジックを適用

**テンプレート使用例（ユーザー側）**：

```html
{{if .Pagination}}
  {{if .Pagination.PrevURL}}
    <link rel="prev" href="{{.Pagination.PrevURL}}">
  {{end}}
  {{if .Pagination.NextURL}}
    <link rel="next" href="{{.Pagination.NextURL}}">
  {{end}}

  {{if .Pagination.PrevURL}}
    <a href="{{.Pagination.PrevURL}}">← 前へ</a>
  {{end}}
  <span>{{.Pagination.CurrentPage}} / {{.Pagination.TotalPages}}</span>
  {{if .Pagination.NextURL}}
    <a href="{{.Pagination.NextURL}}">次へ →</a>
  {{end}}
{{end}}
```

## SEO 考慮事項

- `<link rel="prev">` / `<link rel="next">` により、Googleがページネーションチェーンを発見・クロール可能になる
- 各ページネーションページは独自の正規URLを持つ（`/page/2/` など）
- クエリパラメータ方式（`?page=2`）からの移行時はリダイレクト不要（ページネーション一覧ページはほとんどインデックスされていないため）
