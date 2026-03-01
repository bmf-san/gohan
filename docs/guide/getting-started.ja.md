# Getting Started — gohan 入門ガイド

> gohan のインストールから最初のサイト作成・ローカルプレビューまでを説明します。

> English version: [getting-started.md](getting-started.md)

---

## 前提条件

- Go 1.21 以上
- Git（差分ビルド機能を使う場合）

---

## インストール

```bash
go install github.com/bmf-san/gohan/cmd/gohan@latest
```

インストールを確認します:

```bash
gohan version
# gohan v1.0.0 (commit: abc1234, built: 2024-01-01T00:00:00Z)
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

## 最初のサイトを作る

### ステップ 1: プロジェクトを作成する

```bash
mkdir myblog
cd myblog
```

### ステップ 2: `config.yaml` を作成する

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

全フィールドの詳細は [Configuration](configuration.ja.md) を参照してください。

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
```

### ステップ 4: テーマテンプレートを作成する

```bash
mkdir -p themes/default/templates
```

`themes/default/templates/index.html`:

```html
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>{{.Config.Site.Title}}</title>
  <link rel="stylesheet" href="/assets/style.css">
  <link rel="alternate" type="application/atom+xml" href="/atom.xml">
</head>
<body>
  <header>
    <h1><a href="/">{{.Config.Site.Title}}</a></h1>
  </header>
  <main>
    <ul>
      {{range .Articles}}
      <li>
        <span>{{formatDate "2006-01-02" .FrontMatter.Date}}</span>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

テンプレートの詳細は [テンプレートガイド](templates.ja.md) を参照してください。

### ステップ 5: アセットを追加する（任意）

```bash
mkdir -p assets
cat > assets/style.css << 'EOF'
body { font-family: sans-serif; max-width: 800px; margin: 0 auto; padding: 1rem; }
a { color: #0066cc; }
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

ファイルを保存するたびにブラウザが自動でリロードされます。

---

## よくある操作

### 記事の下書き

`draft: true` を設定した記事はビルドに含まれません:

```yaml
---
title: 作成中の記事
draft: true
---
```

### 差分ビルド

2 回目以降のビルドは自動的に差分ビルドになります:

```bash
gohan build          # 初回: フルビルド
# content/ を編集
gohan build          # 2 回目: 変更分のみ再生成
gohan build --full   # 強制的なフルビルド
```

### 推奨 `.gitignore`

```gitignore
public/
.gohan/
gohan
```

---

## 次のステップ

- [設定リファレンス](configuration.ja.md) — すべての config.yaml オプション
- [テンプレートガイド](templates.ja.md) — テーマのカスタマイズ
- [タクソノミーガイド](taxonomy.ja.md) — タグ・カテゴリーの管理
