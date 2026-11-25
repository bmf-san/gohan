---
title: "プラグインシステム"
description: "外部プラグインで gohan を拡張する。"
slug: "plugin-system"
categories:
  - features
translation_key: "plugin-system"
---

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

#### glossary

全記事の `glossary:` フロントマターを集約し、ロケールごとに用語集ページを 1 件生成します。各用語から定義元の記事へリンクするため、用語ごとの薄いページを量産せず、詳しい記事への索引として機能します。

**config.yaml:**
```yaml
plugins:
  glossary:
    enabled: true
    path: "glossary"   # 省略可: URL のパス（デフォルト "glossary"）
```

**記事フロントマター:**
```yaml
glossary:
  - term: "PEP"
    definition: "認可の判定を強制する構成要素。"
  - term: "認可"
    reading: "にんか"        # 省略可: 読み方として表示し、漢字などの用語の並べ替えにも使う
    definition: "アクセスを許可するかどうかを決めること。"
    id: "authorization"     # 省略可: 用語集ページでのアンカー ID
```

`term` と `definition` は必須です。どちらかが欠けている場合や、`glossary:` がマッピングのリストでない場合は、記事名を示してビルドが失敗します。定義はプレーンテキストです。

ページは用語が 1 件以上あるロケールにだけ生成されます。テーマには `glossary.html` が必要です。

**生成 URL:**
- デフォルトロケール (en): `/glossary/`
- 非デフォルトロケール (ja): `/ja/glossary/`

**並べ替えとグループ分け:**
- 数字やラテン文字で始まる用語は `0-9` と `A`〜`Z` に分けます。全角文字は半角として扱います。
- かなで始まる用語は五十音の行（`あ`、`か`、…、`わ`）に分けます。カタカナはひらがなとして並べ、濁音・半濁音・小書きのかなは元の行に入れます。
- 漢字などで始まる用語は、`reading` があればそれで並べてグループに分け、なければ `#` に入れます。ラテン文字の用語は、かなの読みがあっても英字のグループのままです。
- グループの順序は `0-9`、`A`〜`Z`、かなの行、`#` です。

**重複:** 大文字・小文字と全角・半角の違いを除いて一致する用語は 1 件にまとめます。用語と定義は最も新しい記事のものを使い、読み方や ID が空なら古い記事の値で補います。定義元の記事はすべて列挙します。

**アンカー ID:** `id` を指定した場合はそのまま使います。指定しない場合は `term-` の後に、ラテン文字を小文字にし、それ以外の記号をハイフンにまとめた用語を続けます（例: `OAuth 2.0` → `term-oauth-2-0`、`認可` → `term-認可`）。ID が重なった場合は `-2`、`-3`、… を付けます。

**テンプレートデータ構造** (`.VirtualPageData`):
```
.VirtualPageData["terms"] → []Term   # 表示順の全用語
  Term.ID          string        # アンカー ID
  Term.Term        string
  Term.Reading     string        # 指定がなければ空
  Term.Definition  string
  Term.Articles    []ArticleRef  # 定義元の記事（新しい順）
    ArticleRef.Title  string
    ArticleRef.URL    string
    ArticleRef.Date   time.Time

.VirtualPageData["groups"] → []Group  # 見出しごとにグループ化した用語
  Group.Key    string   # "0-9"、"A"〜"Z"、"あ"〜"わ"、または "#"
  Group.ID     string   # アンカー ID。例: "glossary-a"、"glossary-kana-ka"、"glossary-other"
  Group.Terms  []Term

.VirtualPageData["url"]          → string             # この用語集ページの URL
.VirtualPageData["translations"] → map[string]string  # ロケール → 他ロケールの用語集 URL
```

**サイト全体のデータ** (`.SiteData`): すべてのページに、ロケールごとの用語集 URL と用語数を渡します。用語集があるときだけリンクを表示できます。
```
.SiteData["glossary"] → どの記事にも用語がなければ nil。あれば次の値
  ["urls"]     map[string]string     # ロケール → 用語集 URL
  ["counts"]   map[string]int        # ロケール → 用語数
  ["articles"] map[string][]TermLink # 記事 URL → その記事で定義している用語（用語集の並び順）
                                     # TermLink: Term, URL（"/glossary/#term-jwt"）
```

**テンプレート例:**
```html
<nav>
  {{range index .VirtualPageData "groups"}}<a href="#{{.ID}}">{{.Key}}</a>{{end}}
</nav>
{{range index .VirtualPageData "groups"}}
  <h2 id="{{.ID}}">{{.Key}}</h2>
  <dl>
    {{range .Terms}}
      <dt id="{{.ID}}">{{.Term}}{{if .Reading}}（{{.Reading}}）{{end}}</dt>
      <dd>
        {{.Definition}}
        {{range .Articles}}<a href="{{.URL}}">{{.Title}}</a>{{end}}
      </dd>
    {{end}}
  </dl>
{{end}}
```

**ナビゲーションのリンク（任意のテンプレート）:**
```html
{{with index .SiteData "glossary"}}{{with index (index . "urls") $.CurrentLocale}}
  <a href="{{.}}">用語集</a>
{{end}}{{end}}
```

**記事で定義している用語（記事テンプレート）:**
```html
{{range .Articles}}{{$url := .URL}}{{with index $.SiteData "glossary"}}{{with index (index . "articles") $url}}
  <p>用語集: {{range .}}<a href="{{.URL}}">{{.Term}}</a> {{end}}</p>
{{end}}{{end}}{{end}}
```

### 新しい SitePlugin の追加方法

1. `internal/plugin/<name>/<name>.go` を作成し `plugin.SitePlugin` を実装
2. コンパイル時インターフェースチェックを追加: `var _ plugin.SitePlugin = (*MyPlugin)(nil)`
3. `internal/plugin/registry.go` の `DefaultRegistry()` の `sitePlugins` に登録
4. `.VirtualPageData` を読み取るテーマテンプレートを作成
5. 本セクションにドキュメントを追記

---

## AssetPlugin — ビルド時のバイナリアセット

`Plugin` がテンプレートデータを提供し、`SitePlugin` がHTMLページを生成するのに対し、**`AssetPlugin`** は出力ディレクトリに直接書き出す**ビルド時のバイナリ成果物**（OGP画像など）を生成する。テンプレートはこれらのファイルを規約に基づくURLで参照し、gohanコアはその中身を不透明なものとして扱う。

```
cmd/gohan/build.go
  └── plugin.DefaultRegistry().GenerateAssets(site, outDir, changeSet)  ← "assets" ビルドフェーズ（レンダリング後）
        └── 有効な各 AssetPlugin について:
              AssetPlugin.GenerateAssets(site, outDir, changeSet, cfg) → outDir 配下にファイルを書き出す
```

### AssetPlugin インターフェース

`internal/plugin/plugin.go` で定義：

```go
type AssetPlugin interface {
    Name() string
    Enabled(cfg map[string]interface{}) bool
    GenerateAssets(site *model.Site, outDir string, changeSet *model.ChangeSet, cfg map[string]interface{}) error
}
```

- **`Name()`** — `config.yaml` の `plugins.<name>` で使う一意のキー
- **`Enabled()`** — プラグインを実行するかどうかを制御
- **`GenerateAssets()`** — `outDir` に成果物を書き出す。`changeSet` はフルビルド時は `nil`、差分ビルド時は変更されたファイルの集合であり、変更のない出力をスキップできる

### 組み込み AssetPlugin

#### ogp

各記事の OGP 画像を `<output>/ogp/{slug}.png` に生成する。スラッグをシードとした決定的なグラデーションデザインを用いる。詳細は [OGP](ogp.ja.md) を参照。

**config.yaml:**
```yaml
plugins:
  ogp:
    enabled: true
    logo_file: "assets/images/logo.png"   # オプション
    width: 1200
    height: 630
```

### 新しい AssetPlugin の追加方法

1. `internal/plugin/<name>/<name>.go` を作成し `plugin.AssetPlugin` を実装
2. コンパイル時インターフェースチェックを追加（構造的、または `var _ plugin.AssetPlugin = (*MyPlugin)(nil)`）
3. `internal/plugin/registry.go` の `DefaultRegistry()` の `assetPlugins` に登録
4. 生成されたファイルをテーマテンプレートから参照
5. 本セクションにドキュメントを追記

## スコープ

- `plugin` パッケージによる動的ロードは意図的にスコープ外 — OS制約が多く、静的サイトジェネレーターには不要な複雑性をもたらすため
- プラグインはHTMLを生成しない。データをテーマに提供するのみで、UIはテーマ側が完全制御する
- プラグインから見た記事データは読み取り専用
- VirtualPage は Atom フィードおよびインデックスの記事一覧には含まれない
