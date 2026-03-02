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

## スコープ

- `plugin` パッケージによる動的ロードは意図的にスコープ外 — OS制約が多く、静的サイトジェネレーターには不要な複雑性をもたらすため
- プラグインはHTMLを生成しない。データをテーマに提供するのみで、UIはテーマ側が完全制御する
- プラグインから見た記事データは読み取り専用
