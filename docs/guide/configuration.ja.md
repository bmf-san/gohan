# 設定リファレンス

`config.yaml` はプロジェクトルートに置く唯一の設定ファイルです。

> English version: [configuration.md](configuration.md)

---

## 完全な設定例

```yaml
site:
  title: "My Blog"
  description: "技術系個人ブログ"
  base_url: "https://myblog.example.com"
  language: "ja"
  github_repo: "https://github.com/owner/repo"  # 省略可: 「このページを編集」リンクを有効にする
  github_branch: "main"                          # 省略可: 編集リンクに使うブランチ（デフォルト: main）

build:
  content_dir: "content"
  output_dir: "public"
  assets_dir: "assets"
  static_dir: "static"   # 省略可: output_dir のルートにそのままコピーされる静的ファイルディレクトリ
  exclude_files:
    - "*.draft.md"
    - "_*"
  parallelism: 4
  per_page: 20           # 省略可: ページネーション一覧の記事数（0 = ページネーション無効）

theme:
  name: "default"
  dir: "themes/default"
  params:
    primary_color: "#0066cc"
    footer_text: "© 2024 My Blog"

syntax_highlight:
  theme: "github"
  line_numbers: false

ogp:
  enabled: false         # 省略可: ビルド時に OGP 画像を生成する
  logo_file: ""          # 省略可: 生成画像に埋め込むロゴファイルのパス（プロジェクトルートからの相対）
  width: 1200
  height: 630

i18n:
  locales: [en, ja]      # 省略可: ロケールコードのリスト。空 = シングル言語モード
  default_locale: en     # 省略可: ルート URL で配信するロケール（デフォルト: site.language）

plugins:               # 省略可: プラグイン設定（キー = プラグイン名）
  amazon_books: {}
```

---

## `site` セクション

サイト全体のメタデータを設定します。

| フィールド | 型 | デフォルト | 説明 |
|---|---|---|---|
| `title` | string | *(required)* | サイトタイトル。テンプレートで `.Config.Site.Title` として参照 |
| `description` | string | `""` | サイトの説明。メタタグや Atom フィードに使用 |
| `base_url` | string | *(required)* | サイトのベース URL。末尾スラッシュなし（例: `https://example.com`） |
| `language` | string | `"en"` | BCP 47 言語コード。`<html lang="">` に使用 |
| `github_repo` | string | `""` | GitHub リポジトリのベース URL（例: `https://github.com/owner/repo`）。設定すると `.ContentPath` を使って「このページを編集」リンクを生成できる |
| `github_branch` | string | `"main"` | 編集リンクの生成に使うブランチ |

> `base_url` は `sitemap.xml` と `atom.xml` の URL 生成に使われます。末尾にスラッシュを付けないでください。

---

## `build` セクション

ファイルパスとビルド動作を設定します。

| フィールド | 型 | デフォルト | 説明 |
|---|---|---|---|
| `content_dir` | string | `"content"` | Markdown コンテンツのディレクトリ（プロジェクトルートからの相対パス） |
| `output_dir` | string | `"public"` | HTML 出力先ディレクトリ |
| `assets_dir` | string | `"assets"` | 処理済みアセットのディレクトリ（CSS、画像など） |
| `static_dir` | string | `""` | 出力ルートにそのままコピーされる静的ファイルのディレクトリ（例: `static/404.html` → `public/404.html`） |
| `exclude_files` | []string | `[]` | ビルドから除外するファイルのグロブパターン |
| `parallelism` | int | `4` | HTML 生成の並列数 |
| `per_page` | int | `0` | ページネーション一覧の記事数。`0` でページネーション無効 |

### `exclude_files` の例

```yaml
build:
  exclude_files:
    - "*.draft.md"      # .draft.md で終わるファイルを除外
    - "_*"              # _ で始まるファイルを除外
    - "templates/*"     # templates/ 配下を除外
```

---

## `theme` セクション

使用するテーマとカスタムパラメーターを設定します。

| フィールド | 型 | デフォルト | 説明 |
|---|---|---|---|
| `name` | string | `"default"` | テーマ名。`dir` が未設定の場合 `themes/<name>` が使われる |
| `dir` | string | `"themes/<name>"` | テーマディレクトリのパス（プロジェクトルートからの相対パス） |
| `params` | map[string]string | `{}` | テンプレートから `.Config.Theme.Params.<key>` でアクセスできる任意のパラメーター |

### テンプレートからのアクセス

```html
<style>
  :root { --primary: {{.Config.Theme.Params.primary_color}}; }
</style>
<footer>{{.Config.Theme.Params.footer_text}}</footer>
```

### テーマディレクトリ構成

```
themes/
└── <name>/
    └── templates/      ← テンプレートファイルを置くディレクトリ
        ├── index.html
        ├── article.html
        ├── tag.html
        ├── category.html
        └── archive.html
```

---

## `syntax_highlight` セクション

コードブロックのシンタックスハイライトを設定します（[chroma](https://github.com/alecthomas/chroma) 使用）。

| フィールド | 型 | デフォルト | 説明 |
|---|---|---|---|
| `theme` | string | `"github"` | chroma のカラーテーマ名 |
| `line_numbers` | bool | `false` | 行番号を表示するか |

### 利用可能なテーマ

| テーマ名 | 特徴 |
|---|---|
| `github` | GitHub の明るいテーマ（デフォルト） |
| `monokai` | 暗い背景に鮮やかな色 |
| `dracula` | ダークテーマ |
| `solarized-dark` | Solarized ダーク |
| `solarized-light` | Solarized ライト |
| `nord` | 北欧風ダークテーマ |
| `vs` | Visual Studio 風 |
| `pygments` | Python pygments スタイル |

全テーマのプレビュー: https://xyproto.github.io/splash/docs/

### ハイライトを無効にする

```yaml
syntax_highlight:
  theme: ""   # 空文字列でハイライト無効
```

---

## Front Matter リファレンス

各 Markdown ファイルの先頭に YAML Front Matter を記述します。

```yaml
---
title: "記事タイトル"              # required: 記事タイトル
date: 2024-01-15                  # required: 公開日 (YYYY-MM-DD)
slug: "my-post"                   # optional: URL スラッグ（省略時はタイトルから生成）
draft: false                      # optional: true の場合ビルドから除外 (default: false)
tags:                             # optional: タグ一覧
  - go
  - blog
categories:                       # optional: カテゴリー一覧
  - tech
description: "記事の説明"          # optional: メタ description・フィードの概要
author: "Your Name"               # optional: 著者名
template: "article.html"          # optional: 使用するテンプレートファイル名
---
```

### `slug` の自動生成

`slug` を省略した場合、`title` から自動生成されます:

- スペース → ハイフン
- 大文字 → 小文字
- 例: `"Hello World"` → `hello-world`

---

## `ogp` セクション

ビルド時の OGP 画像生成を設定します。

| フィールド | 型 | デフォルト | 説明 |
|---|---|---|---|
| `enabled` | bool | `false` | ビルド時に OGP 画像を生成する |
| `logo_file` | string | `""` | 生成画像に埋め込むロゴファイルのパス（プロジェクトルートからの相対）。空 = ロゴなし |
| `width` | int | `1200` | 生成画像の幅（ピクセル） |
| `height` | int | `630` | 生成画像の高さ（ピクセル） |

詳細は [docs/features/ogp.ja.md](../features/ogp.ja.md) を参照してください。

---

## `i18n` セクション

多言語サイトの設定です。

| フィールド | 型 | デフォルト | 説明 |
|---|---|---|---|
| `locales` | []string | `[]` | コンテンツディレクトリに存在するロケールコードのリスト。空 = シングル言語モード |
| `default_locale` | string | `site.language` | ルート URL で配信するロケール |

詳細は [docs/features/i18n.ja.md](../features/i18n.ja.md) を参照してください。

---

## `plugins` セクション

プラグインの設定です。キーはプラグイン名、値はプラグイン固有の設定です。

```yaml
plugins:
  amazon_books: {}
```

詳細は [docs/features/plugin-system.ja.md](../features/plugin-system.ja.md) を参照してください。
