# プラグインシステム

## 概要

gohan は **ビルトインプラグインシステム** を備えており、オプション機能をプロジェクトごとに `config.yaml` で有効化できます。ユーザーが Go コードを書く必要はありません。

プラグインは gohan バイナリにコンパイル済みで同梱されます。有効・無効の切り替えは設定変更のみで完結し、再コンパイルは不要です。

## アーキテクチャ

```
cmd/gohan/build.go
  └── plugin.DefaultRegistry().Enrich(site)   ← Process() と Generate() の間で呼ばれる
        └── 有効な各プラグインに対して:
              plugin.TemplateData(article, cfg) → article.PluginData["<name>"] に格納
```

テンプレートでの参照パターン:
```html
{{with index .PluginData "amazon_books"}}
  {{range .books}}
    <a href="{{.LinkURL}}">{{.Title}}</a>
  {{end}}
{{end}}
```

## Plugin インターフェース

`internal/plugin/plugin.go` で定義:

```go
type Plugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    TemplateData(article *model.ProcessedArticle, cfg map[string]interface{}) (map[string]interface{}, error)
}
```

- **`Name()`** — `config.yaml` の `plugins.<name>` および `ProcessedArticle.PluginData` のキーとなる一意識別子
- **`Enabled()`** — プラグインの設定サブマップを受け取り、実行可否を返す
- **`TemplateData()`** — テンプレートに公開する任意のデータを返す

## フロントマター拡張

プラグインは `FrontMatter.Extra` から記事ごとのデータを読み取ります。このフィールドは `yaml:",inline"` により未知の YAML キーをすべてキャプチャします:

```yaml
---
title: My Article
tags: [go]
# プラグイン固有のキー:
books:
  - asin: "4873119464"
    title: "入門 Go"
---
```

## ビルトインプラグイン

| プラグイン | パッケージ | 用途 |
|--------|---------|-----|
| `amazon_books` | `internal/plugin/amazonbooks` | Amazonアフィリエイト書籍カード |

#### amazon_books

記事フロントマターに記載された ASIN から書籍カードデータ（書影URL・商品URL・タイトル）を生成します。

**config.yaml:**
```yaml
plugins:
  amazon_books:
    enabled: true
    tag: "your-associate-tag-22"
```

**記事フロントマター:**
```yaml
books:
  - asin: "4873119464"
    title: "入門 Go"  # 任意。alt属性・キャプション用
```

**テンプレートデータ構造:**
```
.PluginData["amazon_books"].books → []BookCard
  BookCard.ASIN      string
  BookCard.Title     string
  BookCard.ImageURL  string   # images-na.ssl-images-amazon.com CDN
  BookCard.LinkURL   string   # amazon.co.jp/dp/{ASIN}?tag={tag}
```

## 新しいプラグインの追加方法

1. `internal/plugin/<name>/<name>.go` を作成し `plugin.Plugin` を実装
2. コンパイル時インターフェースチェックを追加: `var _ plugin.Plugin = (*MyPlugin)(nil)`
3. `internal/plugin/registry.go` の `DefaultRegistry()` に登録
4. 本セクションにドキュメントを追記

---

## SitePlugin — サイト横断のページ生成

`Plugin` が個別記事を対象とするのに対し、**`SitePlugin`** はサイト全体を対象として **VirtualPage**（Markdown ソースを持たないページ）を生成します。

```
cmd/gohan/build.go
  └── plugin.DefaultRegistry().EnrichVirtual(site)  ← Enrich() の後に呼ばれる
        └── 有効な各 SitePlugin に対して:
              SitePlugin.VirtualPages(site, cfg) → site.VirtualPages に追加
                                                   ↓
                                    HTMLGenerator.buildJobs() がレンダリング
```

テンプレートでの参照パターン（ページ固有データは `.VirtualPageData` にあります）:
```html
{{range index .VirtualPageData "categories"}}
  <h2>{{if .Name}}{{.Name}}{{else}}未分類{{end}}</h2>
  {{range .Books}}
    <a href="{{.LinkURL}}" target="_blank" rel="noopener">
      {{if .ImageURL}}<img src="{{.ImageURL}}" alt="{{.Title}}">{{else}}{{.Title}}{{end}}
    </a>
  {{end}}
{{end}}
```

### SitePlugin インターフェース

`internal/plugin/plugin.go` で定義:

```go
type SitePlugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    VirtualPages(site *model.Site, cfg map[string]interface{}) ([]*model.VirtualPage, error)
}
```

- **`Name()`** — `config.yaml` の `plugins.<name>` キーとなる一意識別子
- **`Enabled()`** — プラグインの実行可否を制御
- **`VirtualPages()`** — サイト全体を検査して 0 件以上の `VirtualPage` を返す

### VirtualPage フィールド

```
VirtualPage.OutputPath  string   // 出力ディレクトリからの相対ファイルパス（例: "bookshelf/index.html"）
VirtualPage.URL         string   // 正規 URL パス（例: "/bookshelf/"）
VirtualPage.Template    string   // テーマのテンプレートファイル名（例: "bookshelf.html"）
VirtualPage.Locale      string   // ロケールコード（例: "en" または "ja"）
VirtualPage.Data        map[string]interface{}  // テンプレートで .VirtualPageData として参照
```

### ビルトイン SitePlugin

#### bookshelf

全記事の `books:` フロントマターを集約し、ロケールごとに本棚ページを 1 件生成します。

**config.yaml:**
```yaml
plugins:
  bookshelf:
    enabled: true
    tag: "your-associate-tag-22"   # Amazon アソシエイトトラッキングタグ
```

**記事フロントマター:**
```yaml
books:
  - asin: "4873119464"          # Amazon ASIN — 書影 + Amazon リンクを生成
    title: "入門 Go"
  - url: "https://booth.pm/..."  # 非 Amazon: 直販 URL（書影なし）
    title: "技術同人誌タイトル"
```

`url:` を `asin:` の代わりに指定した場合、`ImageURL` は空になり `LinkURL` はその URL になります。

**生成 URL:**
- デフォルトロケール (en): `/bookshelf/`
- 非デフォルトロケール (ja): `/ja/bookshelf/`

**テンプレートデータ構造** (`.VirtualPageData`):
```
.VirtualPageData["books"] → []BookEntry  # 日付降順（新しい順）
  BookEntry.ASIN          string
  BookEntry.Title         string
  BookEntry.ImageURL      string   # images-na.ssl-images-amazon.com CDN。url: のみのエントリは空
  BookEntry.LinkURL       string   # amazon.co.jp/dp/{ASIN}?tag={tag}、または直販 url: の値
  BookEntry.ArticleSlug   string   # 書評記事のスラッグ
  BookEntry.ArticleTitle  string   # 書評記事のタイトル
  BookEntry.ArticleURL    string   # 書評記事の正規 URL
  BookEntry.Categories    []string # 書評記事のカテゴリー
  BookEntry.Date          time.Time

.VirtualPageData["categories"] → []CategoryGroup  # カテゴリーでグループ化（アルファベット順）
  CategoryGroup.Name      string       # カテゴリー名。空文字列 = 未分類（末尾に配置）
  CategoryGroup.Books     []BookEntry
```

### 新しい SitePlugin の追加方法

1. `internal/plugin/<name>/<name>.go` を作成し `plugin.SitePlugin` を実装
2. コンパイル時インターフェースチェックを追加: `var _ plugin.SitePlugin = (*MyPlugin)(nil)`
3. `internal/plugin/registry.go` の `DefaultRegistry()` の `sitePlugins` に登録
4. `.VirtualPageData` を読み取るテーマテンプレートを作成
5. 本セクションにドキュメントを追記

## スコープ

- `plugin` パッケージによる動的ロードは意図的にスコープ外 — OS制約が多く、静的サイトジェネレーターには不要な複雑性をもたらすため
- プラグインはHTMLを生成しない。データをテーマに提供するのみで、UIはテーマ側が完全制御する
- プラグインから見た記事データは読み取り専用
- VirtualPage は Atom フィードおよびインデックスの記事一覧には含まれない
