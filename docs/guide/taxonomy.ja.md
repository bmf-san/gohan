# タクソノミーガイド

タクソノミーとは、記事を **タグ** や **カテゴリー** で分類する仕組みです。
gohan は各タグ・カテゴリーに対応したページを自動生成します。

> English version: [taxonomy.md](taxonomy.md)

---

## 記事への設定

Front Matter でタグとカテゴリーを指定します:

```markdown
---
title: Go の並行処理を理解する
date: 2024-03-01
slug: go-concurrency
tags:
  - go
  - concurrency
  - goroutine
categories:
  - tech
  - programming
---
```

- `tags`: 記事のキーワード。複数指定可。1 タグ 1 ページが生成されます
- `categories`: 記事の分類。複数指定可。1 カテゴリー 1 ページが生成されます

---

## 生成されるページ

上記の記事がある場合、以下のページが生成されます:

```
public/
├── tags/
│   ├── go/index.html
│   ├── concurrency/index.html
│   └── goroutine/index.html
└── categories/
    ├── tech/index.html
    └── programming/index.html
```

---

## テンプレートでのアクセス

### タグ一覧ページ (`tag.html`)

タグページでは `.Articles` に **そのタグを持つ記事** が絞り込まれて渡されます:

```html
<!-- themes/default/templates/tag.html -->
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>タグ別記事 — {{.Config.Site.Title}}</title>
</head>
<body>
  <header><nav><a href="/">← {{.Config.Site.Title}}</a></nav></header>
  <main>
    <h2>記事 ({{len .Articles}} 件)</h2>
    <ul>
      {{range .Articles}}
      <li>
        <time>{{formatDate "2006-01-02" .FrontMatter.Date}}</time>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

### カテゴリー一覧ページ (`category.html`)

カテゴリーページでは `.Articles` に **そのカテゴリーを持つ記事** が渡されます:

```html
<!-- themes/default/templates/category.html -->
<!DOCTYPE html>
<html lang="{{.Config.Site.Language}}">
<head>
  <meta charset="UTF-8">
  <title>カテゴリー別記事 — {{.Config.Site.Title}}</title>
</head>
<body>
  <main>
    <h2>記事一覧 ({{len .Articles}} 件)</h2>
    <ul>
      {{range .Articles}}
      <li>
        <a href="/posts/{{.FrontMatter.Slug}}/">{{.FrontMatter.Title}}</a>
      </li>
      {{end}}
    </ul>
  </main>
</body>
</html>
```

### タグ・カテゴリーの全一覧

`.Tags` と `.Categories` を使うと全タグ・全カテゴリーに動的リンクを生成できます:

```html
<section>
  <h3>タグ</h3>
  <ul>
    {{range .Tags}}
    <li><a href="{{tagURL .Name}}">{{.Name}}</a></li>
    {{end}}
  </ul>

  <h3>カテゴリー</h3>
  <ul>
    {{range .Categories}}
    <li><a href="{{categoryURL .Name}}">{{.Name}}</a></li>
    {{end}}
  </ul>
</section>
```

---

## URL の生成

gohan はタグ名・カテゴリー名を URL スラッグに変換します:

- スペース → ハイフン
- 大文字 → 小文字
- 例: `"Go Language"` → `/tags/go-language/`

---

## アーカイブページ

公開日 (`date`) から年別のアーカイブページが自動生成されます:

```
public/
└── archive/
    ├── 2024/index.html
    └── 2023/index.html
```

`archive.html` テンプレートでは `.Articles` に **その年の記事** が渡されます。

---

## タクソノミーの設計指針

- **タグ**: 記事の具体的なキーワード（`go`, `docker`, `postgresql` など）。数が多くても構いません
- **カテゴリー**: 大分類（`tech`, `life`, `book` など）。少め（5〜10 種類程度）を推奨します

---

## 関連ページ

- [テンプレートガイド](templates.ja.md)
- [設定リファレンス](configuration.ja.md)
