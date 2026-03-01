# Getting Started — gohan 入門ガイド

> This guide walks you through installing gohan, creating your first site, and publishing it as static HTML.

---

## 前提条件 / Prerequisites

- Go 1.21 以上
- Git（差分ビルド機能を使う場合）

---

## インストール / Installation

```bash
go install github.com/bmf-san/gohan/cmd/gohan@latest
```

インストールを確認します:

```bash
gohan version
# gohan v1.0.0 (commit: abc1234, built: 2024-01-01T00:00:00Z)
```

---

## 最初のサイトを作る / Create Your First Site

### ステップ 1: プロジェクトを作成する

```bash
mkdir myblog
cd myblog
```

### ステップ 2: `config.yaml` を作成する

サイト設定ファイルをプロジェクトルートに作成します。

```yaml
site:
  title: My Blog
  description: A simple personal blog
  base_url: https://myblog.example.com
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

### ステップ 3: 最初の記事を作成する

```bash
gohan new post --slug=hello-world --title="Hello, World!"
```

`content/posts/hello-world.md` が作成されます。編集して本文を追加しましょう:

```markdown
---
title: Hello, World!
date: 2024-01-15
slug: hello-world
tags:
  - go
  - blog
categories:
  - tech
draft: false
description: はじめての gohan ブログ記事
---

# Hello, World!

**gohan** でブログを始めました！

## gohan の特徴

- 差分ビルドによる高速ビルド
- Go html/template によるテーマカスタマイズ
- Mermaid 図やシンタックスハイライトのサポート

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, gohan!")
}
```
```

### ステップ 4: テーマテンプレートを作成する

`themes/default/templates/` ディレクトリを作成し、最小限のテンプレートを用意します:

```bash
mkdir -p themes/default/templates
```

`themes/default/templates/index.html`:

```html
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{{.Config.Site.Description}}">
  <title>{{.Config.Site.Title}}</title>
  <link rel="stylesheet" href="/assets/style.css">
  <link rel="alternate" type="application/atom+xml" href="/atom.xml">
</head>
<body>
  <header>
    <h1><a href="/">{{.Config.Site.Title}}</a></h1>
  </header>
  <main>
    <ul class="post-list">
      {{range .Articles}}
      <li>
        <span class="date">{{formatDate "2006-01-02" .FrontMatter.Date}}</span>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

`themes/default/templates/article.html`:

```html
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>{{(index .Articles 0).FrontMatter.Title}} — {{.Config.Site.Title}}</title>
  <link rel="stylesheet" href="/assets/style.css">
  <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
</head>
<body>
  <header>
    <a href="/">← {{.Config.Site.Title}}</a>
  </header>
  <main>
    {{with (index .Articles 0)}}
    <article>
      <h1>{{.FrontMatter.Title}}</h1>
      <time>{{formatDate "2006-01-02" .FrontMatter.Date}}</time>
      {{if .FrontMatter.Tags}}
      <ul class="tags">
        {{range .FrontMatter.Tags}}
        <li><a href="{{tagURL .}}">{{.}}</a></li>
        {{end}}
      </ul>
      {{end}}
      <div class="content">{{.HTMLContent}}</div>
    </article>
    {{end}}
  </main>
</body>
</html>
```

### ステップ 5: アセットを追加する（任意）

```bash
mkdir -p assets
cat > assets/style.css << 'EOF'
body { font-family: sans-serif; max-width: 800px; margin: 0 auto; padding: 1rem; }
a { color: #0066cc; }
.date { color: #888; margin-right: 1em; }
.tags { list-style: none; display: flex; gap: 0.5em; padding: 0; }
EOF
```

### ステップ 6: サイトをビルドする

```bash
gohan build
```

`public/` ディレクトリにサイトが生成されます:

```
public/
├── index.html
├── sitemap.xml
├── atom.xml
├── posts/
│   └── hello-world/
│       └── index.html
└── assets/
    └── style.css
```

### ステップ 7: 開発サーバーで確認する

```bash
gohan serve
# http://127.0.0.1:1313 でプレビュー
```

---

## よくある操作 / Common Operations

### 記事の下書き

`draft: true` を設定した記事はビルドに含まれません:

```yaml
---
title: 作成中の記事
draft: true
---
```

### 差分ビルド

2 回目以降のビルドは自動的に差分ビルドになります（変更されたファイルのみ再生成）:

```bash
gohan build          # 初回: フルビルド
# content/ を編集
gohan build          # 2 回目: 変更分のみ再生成
gohan build --full   # 強制的なフルビルド
```

### .gitignore の設定

```gitignore
public/
.gohan/
gohan
```

---

## 次のステップ

- [設定リファレンス](configuration.md) — すべての config.yaml オプション
- [テンプレートガイド](templates.md) — テーマのカスタマイズ
- [タクソノミーガイド](taxonomy.md) — タグ・カテゴリーの管理
