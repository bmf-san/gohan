---
title: "OGP"
description: "Open Graph / Twitter Card メタデータを自動生成。"
slug: "ogp"
categories:
  - features
translation_key: "ogp"
---

## 概要

gohan はビルド時に純粋なGoでOGP（Open Graph Protocol）サムネイル画像を、組み込みの `ogp` **アセットプラグイン**で生成する。各記事のスラッグから決定的なグラデーション画像を生成して一意の `og:image` を割り当てるため、手動での画像作成が不要になる。

## 出力

```
public/
└── ogp/
    └── {slug}.png    # 記事ごとに1200×630pxのOGP画像
```

各記事ページの `og:image` タグには以下のURLを使用する：

```
{{.Config.Site.BaseURL}}/ogp/{{.Article.FrontMatter.Slug}}.png
```

一覧ページ（インデックス、タグ、カテゴリ）はユーザーが用意したデフォルト画像にフォールバックする：

```
{{.Config.Site.BaseURL}}/assets/images/ogp-default.png
```

## 設定

OGP生成は組み込みの**アセットプラグイン**である。`config.yaml` の `plugins.ogp` で有効化する：

```yaml
plugins:
  ogp:
    enabled: true
    logo_file: "assets/images/logo.png"   # オプション: 左上に重ねるロゴ
    width: 1200                            # オプション: デフォルト1200
    height: 630                            # オプション: デフォルト630
```

各画像は記事のスラッグをシードとした決定的なデザイン（斜めのグラデーションと幾何学模様）であり、**フォントファイルは不要**で、同じ記事からは常に同じ画像が生成される。

| キー | 型 | デフォルト | 説明 |
|---|---|---|---|
| `enabled` | bool | `false` | ビルド時にOGP画像を生成する |
| `logo_file` | string | `""` | オプション: 左上に描画するPNG/JPEGロゴ |
| `width` | int | `1200` | 画像の幅（ピクセル） |
| `height` | int | `630` | 画像の高さ（ピクセル） |

## 実装

OGP生成は **`internal/plugin/ogp`** パッケージにあり、`plugin.AssetPlugin` インターフェースを実装する：

```go
type AssetPlugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    GenerateAssets(site *model.Site, outDir string, changeSet *model.ChangeSet, cfg map[string]interface{}) error
}
```

- `internal/plugin/registry.go` に登録され、ビルドパイプラインの `assets` フェーズ（HTMLレンダリング後）で呼び出される。
- 設定は `plugins.ogp` マップから読み取る。トップレベルの `ogp:` 設定セクションは廃止された。
- 標準ライブラリ `image`・`image/color`・`image/png`・`image/draw` に加え、ロゴのスケーリングに `golang.org/x/image/draw` を使用する。
- レンダリングパイプライン：スラッグをシードとした斜めのグラデーション背景 → 幾何学的なアクセント図形 → オプションのロゴオーバーレイ（左上）。
- 出力 `.png` が既に存在し、かつ元の記事が変更されていない場合は記事ごとの生成をスキップする（`ChangeSet` による差分対応）。

**テンプレート使用例（ユーザー側）**：

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

## 依存関係

| パッケージ | 用途 |
|---|---|
| `image`, `image/png`, `image/draw` | 標準ライブラリ — キャンバス生成とPNGエンコード |
| `image/jpeg` | 標準ライブラリ — JPEGロゴファイルのデコード |
| `golang.org/x/image/draw` | オプションのロゴオーバーレイの高品質スケーリング |

`golang.org/x/image` のビルドオーバーヘッドは無視できる。
