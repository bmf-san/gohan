---
title: "検索インデックス"
description: "クライアントサイド検索用の JSON 検索インデックスを生成する。"
slug: "search-index"
categories:
  - features
translation_key: "search-index"
---

gohan はビルド時に `search-index.json` を生成します。外部ツールやサービスを使わずに、テーマへクライアントサイドの記事検索を追加できます。

> English version: [search-index.md](search-index.md)

---

## 動作の仕組み

ビルド時、gohan は「feeds」フェーズで `sitemap.xml` や `atom.xml` と並べて `search-index.json` を書き出します。設定は不要で、自動的に生成されます。

各エントリーは **メタデータのみ**（本文テキストは含まない）を保持するため、記事が数百件あるサイトでもインデックスは小さく保たれます:

| フィールド | 値の例 | 説明 |
|---|---|---|
| `title` | `Hello, World!` | 記事タイトル |
| `url` | `https://example.com/posts/hello/` | 絶対 URL（フィードと同じ規則） |
| `description` | `概要文` | meta description |
| `summary` | `自動生成された抜粋…` | 記事のサマリー |
| `tags` | `["go", "ssg"]` | タグ（空のときは省略） |
| `categories` | `["features"]` | カテゴリー（空のときは省略） |
| `date` | `2026-03-14T00:00:00Z` | 公開日（RFC 3339） |
| `locale` | `ja` | ロケールコード（i18n 利用時のみ） |

記事は **新しい順** に並びます。下書きや未来日付の記事は、`--draft` や `--future` を指定しない限り除外されます。

### 出力の形

```json
{
  "generated": "2026-06-27T04:57:01Z",
  "count": 2,
  "articles": [
    {
      "title": "i18n（多言語対応）",
      "url": "https://bmf-san.github.io/gohan/ja/features/i18n/",
      "description": "ディレクトリベースのロケール構造で多言語サイトを構築。",
      "summary": "## 概要",
      "categories": ["features"],
      "locale": "ja"
    }
  ]
}
```

---

## i18n

`i18n.locales` を設定すると、フィードやサイトマップと同じ規則でインデックスがロケール別に分割されます:

- ルートの `search-index.json` には **デフォルトロケール** の記事のみが含まれます。
- デフォルト以外の各ロケールは、それぞれ `{locale}/search-index.json`（例: `ja/search-index.json`）を持ちます。

これにより、各言語のページは自分の言語のエントリーだけを読み込めます。

---

## テーマでの使用

インデックスは静的な JSON ファイルなので、依存ライブラリなしの数行の素の JavaScript で十分です:

```html
<input id="q" type="search" placeholder="検索…">
<ul id="results"></ul>

<script>
  const out = document.getElementById("results");
  fetch("/search-index.json")
    .then((r) => r.json())
    .then(({ articles }) => {
      document.getElementById("q").addEventListener("input", (e) => {
        const q = e.target.value.toLowerCase().trim();
        out.replaceChildren();
        if (!q) return;
        for (const a of articles) {
          const hay = [a.title, a.description, (a.tags || []).join(" ")]
            .join(" ")
            .toLowerCase();
          if (!hay.includes(q)) continue;
          const li = document.createElement("li");
          const link = document.createElement("a");
          link.href = a.url;          // プロパティとして設定し HTML 注入を防ぐ
          link.textContent = a.title; // textContent でタイトルを安全にエスケープ
          li.appendChild(link);
          out.appendChild(li);
        }
      });
    });
</script>
```

あいまい検索やスコアリングが必要な場合は、`articles` を [MiniSearch](https://github.com/lucaong/minisearch) や [Fuse.js](https://www.fusejs.io/) などの軽量なクライアントサイドライブラリに渡してください。

> サブパス配信（例: GitHub Pages の `/gohan`）の場合は、fetch する URL に `base_path` を付けてください。

---

## 実装の詳細

ロジックは `internal/generator/searchindex.go` にあり、フィードやサイトマップの生成と同じ仕組みです。

| ファイル | 変更内容 |
|---|---|
| `internal/generator/searchindex.go` | `GenerateSearchIndex`: メタデータのみのエントリーを新しい順に書き出し、i18n に応じて分割 |
| `cmd/gohan/build.go` | ビルドの「feeds」フェーズで `GenerateSearchIndex` を呼び出し |
| `internal/generator/searchindex_test.go` | ユニットテスト（正常出力・空サイト・フィールド対応・i18n 分割） |
