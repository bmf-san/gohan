# gohan

[![CI](https://github.com/bmf-san/gohan/actions/workflows/ci.yml/badge.svg)](https://github.com/bmf-san/gohan/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmf-san/gohan.svg)](https://pkg.go.dev/github.com/bmf-san/gohan)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**gohan** は Go で実装されたシンプルな静的サイトジェネレーター（SSG）です。  
個人ブログや小規模サイト向けに、差分ビルドによる高速な HTML 生成を提供します。

> A simple, fast static site generator written in Go — featuring incremental builds, syntax highlighting, Mermaid diagrams, and a live-reload dev server.

---

## 特徴 / Features

- **差分ビルド** — 変更されたファイルのみを再生成し、ビルド時間を最小化
- **Markdown + Front Matter** — GFM (GitHub Flavored Markdown) 対応
- **シンタックスハイライト** — [chroma](https://github.com/alecthomas/chroma) ベースのコードブロックとカラーテーマ
- **Mermaid 図** — ` ```mermaid ``` ` フェンスコードブロックをインタラクティブな図に変換
- **タクソノミー** — タグ・カテゴリーによる記事分類とページ生成
- **Atom フィード / サイトマップ** — `atom.xml`・`sitemap.xml` を自動生成
- **開発サーバー** — ファイル変更を検知してブラウザをライブリロード (`gohan serve`)
- **テーマ** — Go `html/template` による完全カスタマイズ可能なテーマ

---

## インストール / Installation

### go install（推奨）

```bash
go install github.com/bmf-san/gohan/cmd/gohan@latest
```

### ソースからビルド

```bash
git clone https://github.com/bmf-san/gohan.git
cd gohan
make install
```

### バイナリダウンロード

[GitHub Releases](https://github.com/bmf-san/gohan/releases) から各プラットフォーム向けのビルド済みバイナリをダウンロードできます。

---

## クイックスタート / Quick Start

### 1. プロジェクトディレクトリの作成

```bash
mkdir myblog && cd myblog
```

### 2. `config.yaml` の作成

```yaml
site:
  title: My Blog
  description: A simple blog
  base_url: https://example.com
  language: ja

build:
  content_dir: content
  output_dir: public
  assets_dir: assets
  parallelism: 4

theme:
  name: default

syntax_highlight:
  theme: github
  line_numbers: false
```

### 3. 記事の作成

```bash
gohan new post --slug=hello-world --title="Hello World"
# または手動で content/posts/hello-world.md を作成
```

Front Matter の例:

```markdown
---
title: Hello World
date: 2024-01-01
slug: hello-world
tags:
  - go
  - blog
categories:
  - tech
draft: false
---

# Hello World

記事の本文を Markdown で書きます。
```

### 4. テーマテンプレートの配置

`themes/default/templates/` ディレクトリに以下のテンプレートファイルを作成します:

```
themes/
└── default/
    └── templates/
        ├── index.html      # サイトトップページ
        ├── article.html    # 記事ページ
        ├── tag.html        # タグ一覧ページ
        ├── category.html   # カテゴリー一覧ページ
        └── archive.html    # アーカイブページ
```

テンプレートには Go の `html/template` 構文が使えます。テンプレートに渡されるデータは [`model.Site`](internal/model/model.go) 型です。

```html
<!-- themes/default/templates/index.html -->
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>{{.Config.Site.Title}}</title>
</head>
<body>
  <h1>{{.Config.Site.Title}}</h1>
  <ul>
    {{range .Articles}}
    <li>
      <a href="/posts/{{.FrontMatter.Slug}}/">
        {{.FrontMatter.Title}}
      </a>
      <span>{{formatDate "2006-01-02" .FrontMatter.Date}}</span>
    </li>
    {{end}}
  </ul>
</body>
</html>
```

### 5. ビルド

```bash
gohan build
```

出力先は `public/` ディレクトリです。

### 6. 開発サーバーの起動

```bash
gohan serve
# http://localhost:1313 でライブプレビュー
```

ファイルを保存するたびにブラウザが自動でリロードされます。

---

## ディレクトリ構成 / Directory Structure

```
.
├── config.yaml           # サイト設定
├── content/
│   ├── posts/            # ブログ記事（タグ・カテゴリー・アーカイブ対象）
│   │   └── my-post.md
│   └── pages/            # 静的ページ（About, Contact など）
│       └── about.md
├── assets/               # CSS・画像などの静的ファイル
│   └── style.css
├── themes/
│   └── default/
│       └── templates/    # Go html/template テンプレート
│           ├── index.html
│           ├── article.html
│           ├── tag.html
│           ├── category.html
│           └── archive.html
└── public/               # ビルド出力（.gitignore 推奨）
```

---

## 設定リファレンス / Configuration Reference

`config.yaml` の全フィールド:

```yaml
site:
  title: "サイトタイトル"          # required: サイト名
  description: "サイトの説明"      # optional: メタ description
  base_url: "https://example.com"  # required: ベース URL（末尾スラッシュなし）
  language: "ja"                   # optional: サイト言語 (default: "en")

build:
  content_dir: "content"           # optional: コンテンツディレクトリ (default: "content")
  output_dir: "public"             # optional: 出力ディレクトリ (default: "public")
  assets_dir: "assets"             # optional: アセットディレクトリ (default: "assets")
  exclude_files: []                # optional: ビルドから除外するファイルパターン
  parallelism: 4                   # optional: 並列処理数 (default: 4)

theme:
  name: "default"                  # optional: テーマ名 (default: "default")
  dir: "themes/default"            # optional: テーマディレクトリ (default: "themes/<name>")
  params:                          # optional: テーマカスタムパラメーター
    primary_color: "#0066cc"

syntax_highlight:
  theme: "github"                  # optional: chroma テーマ名 (default: "github")
  line_numbers: false              # optional: 行番号表示 (default: false)
```

### chroma テーマ一覧

よく使われるテーマ: `github`, `monokai`, `dracula`, `solarized-dark`, `nord`, `vs`

全テーマは [chroma styles](https://xyproto.github.io/splash/docs/) を参照してください。

---

## CLI コマンドリファレンス / CLI Reference

### `gohan build`

サイトをビルドします。

```
gohan build [flags]

Flags:
  --config string       設定ファイルのパス (default: "config.yaml")
  --output string       出力ディレクトリの上書き
  --full                差分ビルドをスキップして全記事を再生成
  --parallel int        並列処理数の上書き (0 = config 値を使用)
  --dry-run             ファイルを書き出さずにビルドをシミュレート
  --log-format string   ログフォーマット: text または json (default: "text")
```

### `gohan new`

新しい記事・ページのスケルトンを作成します。

```
gohan new <type> [flags]

Types:
  post   content/posts/<slug>.md を作成
  page   content/pages/<slug>.md を作成

Flags:
  --slug string    記事のスラッグ (required)
  --title string   記事のタイトル（省略時は slug から自動生成）
```

例:

```bash
gohan new post --slug=my-first-post --title="はじめての投稿"
gohan new page --slug=about --title="このサイトについて"
```

### `gohan serve`

差分ビルド + ライブリロード付き開発サーバーを起動します。

```
gohan serve [flags]

Flags:
  --config string   設定ファイルのパス (default: "config.yaml")
  --port int        listen ポート (default: 1313)
  --host string     bind アドレス (default: "127.0.0.1")
```

### `gohan version`

バージョン情報を表示します。

```bash
$ gohan version
gohan v1.0.0 (commit: abc1234, built: 2024-01-01T00:00:00Z)
```

---

## テンプレート変数 / Template Variables

テンプレートには `model.Site` 型の値が渡されます。

| 変数 | 型 | 説明 |
|---|---|---|
| `.Config.Site.Title` | `string` | サイトタイトル |
| `.Config.Site.Description` | `string` | サイト説明 |
| `.Config.Site.BaseURL` | `string` | ベース URL |
| `.Config.Site.Language` | `string` | サイト言語 |
| `.Config.Theme.Params` | `map[string]string` | テーマカスタムパラメーター |
| `.Articles` | `[]*ProcessedArticle` | 記事一覧（ページにより絞り込み済み） |
| `.Tags` | `[]Taxonomy` | タグ一覧 |
| `.Categories` | `[]Taxonomy` | カテゴリー一覧 |

### 記事フィールド (`.Articles` の各要素)

| フィールド | 型 | 説明 |
|---|---|---|
| `.FrontMatter.Title` | `string` | 記事タイトル |
| `.FrontMatter.Date` | `time.Time` | 公開日 |
| `.FrontMatter.Slug` | `string` | URL スラッグ |
| `.FrontMatter.Tags` | `[]string` | タグ一覧 |
| `.FrontMatter.Categories` | `[]string` | カテゴリー一覧 |
| `.FrontMatter.Description` | `string` | 記事の説明 |
| `.FrontMatter.Draft` | `bool` | 下書きフラグ |
| `.HTMLContent` | `template.HTML` | レンダリング済み HTML |
| `.Summary` | `string` | 記事の要約（先頭 200 文字） |

### 組み込みテンプレート関数

| 関数 | シグネチャ | 説明 |
|---|---|---|
| `formatDate` | `formatDate layout time` | 日付フォーマット（例: `formatDate "2006-01-02" .FrontMatter.Date`） |
| `tagURL` | `tagURL name` | タグページの URL を生成 |
| `categoryURL` | `categoryURL name` | カテゴリーページの URL を生成 |
| `markdownify` | `markdownify str` | Markdown 文字列を HTML に変換 |

---

## 開発者向け / For Developers

```bash
# テスト実行
make test

# カバレッジ確認
make coverage

# リント
make lint

# クリーン
make clean
```

詳細な設計ドキュメントは [docs/DESIGN_DOC.md](docs/DESIGN_DOC.md) を参照してください。

---

## ライセンス / License

[MIT License](LICENSE) — Copyright (c) bmf-san