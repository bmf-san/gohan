# OGP画像生成

## 概要

gohan はビルド時に純粋なGoでOGP（Open Graph Protocol）サムネイル画像を生成する。各記事のタイトルから一意の `og:image` を生成するため、手動での画像作成が不要になる。

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

`config.yaml` に `ogp` ブロックを追加する：

```yaml
ogp:
  enabled: true
  background_color: "#1e1e2e"
  text_color: "#cdd6f4"
  font_file: "assets/fonts/NotoSansJP-Bold.ttf"   # TTF/OTF、CJK対応に必要
  logo_file: "assets/images/logo.png"              # オプションのロゴオーバーレイ
  width: 1200
  height: 630
```

フォントファイルはユーザーが用意する。TTF/OTF形式であれば任意のフォントが利用可能。

## データモデル

`model.go` に `OGPConfig` を追加する：

```go
// OGPConfig はビルド時OGP画像生成の設定を保持する。
type OGPConfig struct {
    Enabled         bool   `yaml:"enabled"`
    BackgroundColor string `yaml:"background_color"`
    TextColor       string `yaml:"text_color"`
    FontFile        string `yaml:"font_file"`
    LogoFile        string `yaml:"logo_file"` // 空文字列はロゴなし
    Width           int    `yaml:"width"`
    Height          int    `yaml:"height"`
}
```

`Config` に追加する：

```go
type Config struct {
    // ... 既存フィールド ...
    OGP OGPConfig `yaml:"ogp"`
}
```

## 実装

**`internal/generator/ogp.go`**（新規ファイル）
- `OutputGenerator` を満たす `OGPGenerator` を実装
- 標準ライブラリ `image`・`image/color`・`image/png`・`image/draw` を使用
- TrueTypeレンダリングに `golang.org/x/image/font` と `golang.org/x/image/font/opentype` を使用
- `golang.org/x/image/math/fixed` でフォントレンダリング用の固定小数点演算を行う
- レンダリングパイプライン：背景塗りつぶし → ロゴ描画（設定時） → タイトルテキストを縦横中央に折り返して描画
- `ogp.enabled: false` の場合は生成をスキップ
- 出力 `.png` が存在し、かつ元の記事が変更されていない場合はスキップ（`ChangeSet` によるキャッシュ対応）

**`internal/generator/generator.go`**
- `cfg.OGP.Enabled` が true の場合、ビルドパイプライン内で `OGPGenerator.Generate()` を呼び出す

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
| `golang.org/x/image/font` | フォントフェースインターフェースとテキスト描画 |
| `golang.org/x/image/font/opentype` | TTF/OTFフォントファイルの読み込み |
| `golang.org/x/image/math/fixed` | フォントレンダリング用固定小数点演算 |

`golang.org/x/image` は多くのGoプロジェクトで推移的に取り込まれており、ビルドオーバーヘッドは無視できる。
