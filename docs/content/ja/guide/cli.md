---
title: "CLI リファレンス"
description: "gohan CLI のコマンドリファレンス。"
slug: "cli"
categories:
  - guide
translation_key: "cli"
---

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `gohan build` | サイトをビルド（デフォルトで差分ビルド） |
| `gohan build --full` | フルビルドを強制実行 |
| `gohan build --dry-run` | ファイルを書き出さずにビルドをシミュレート |
| `gohan build --draft` | `draft: true` の記事をビルドに含める |
| `gohan build --future` | `date` が未来の記事をビルドに含める |
| `gohan build --stats` | ビルド後にフェーズごとの所要時間と件数を表示 |
| `gohan build --explain` | 再ビルド対象のファイルとその理由を表示 |
| `gohan init [--force] [<dir>]` | 新規プロジェクト（config / content / archetypes）をスキャフォールド |
| `gohan new [--type=post] [--title=<t>] <slug>` | 新規記事スケルトンを作成 |
| `gohan new --type=<section> --archetype=<name> <slug>` | archetype テンプレートを使ってカスタムセクションのコンテンツを作成 |
| `gohan new --type=page [--title=<t>] <slug>` | 新規ページスケルトンを作成 |
| `gohan serve` | ライブリロード付き開発サーバーを起動 |
| `gohan version` | バージョン情報を表示 |

---

## `gohan build`

`content/` をスキャンして、設定された出力ディレクトリにサイトを生成します。

デフォルトでは前回のビルドから変更されたファイルのみを再生成します（差分ビルド）。すべてを再生成するには `--full` を使用します。

**フラグ**

| フラグ | 説明 |
|---|---|
| `--full` | 差分検出をスキップしてすべてのページを再生成 |
| `--dry-run` | ファイルを書き出さずに生成対象を表示 |
| `--draft` | Front Matter に `draft: true` を持つ記事をビルドに含める。デフォルトではドラフトは除外される。 |
| `--future` | `date` が現在時刻よりも未来の記事をビルドに含める。デフォルトでは未来日付の記事は除外されるため、`date` を未来に設定すれば記事を「予約公開」できる。 |
| `--stats` | フェーズごと（parse / diff / process / plugins / render / feeds / manifest）の所要時間と合計時間をビルド完了後に表示する。 |
| `--explain` | 再ビルドのトリガーとなったコンテンツファイルを表示する。フルビルドの場合は理由（`--full` フラグ、設定ハッシュの変化、マニフェスト未生成など）も併せて表示する。 |

---

---

## `gohan init`

指定したディレクトリ（省略時はカレントディレクトリ）に新規 gohan プロジェクトをスキャフォールドします。

**使い方**

```sh
gohan init [--force] [<dir>]
```

最小構成の `config.yaml`、標準の `content/` および `archetypes/` ディレクトリ、スタータ用の `README.md` を生成します。対象ディレクトリが空でない場合は `--force` を付けない限り処理を中止します。既存ファイルは上書きされません（`--force` は新規ファイルの追加のみを許可します）。

**フラグ**

| フラグ | 説明 |
|---|---|
| `--force` | 空でないディレクトリへのスキャフォールドを許可する。既存ファイルは保持される。 |

---

## `gohan new`

Front Matter が事前入力されたコンテンツスケルトンを作成します。

**フラグ**

| フラグ | 説明 |
|---|---|
| `--type` | コンテンツセクション: `post`（デフォルト → `content/posts/`）、`page`（→ `content/pages/`）、その他任意の名前（→ `content/<name>/`）。 |
| `--archetype` | `archetypes/` 配下の archetype テンプレート名。省略時は `--type` の値を使用。 |
| `--title` | 記事タイトル（省略時はスラッグをタイトルケースに変換） |

**使用例**

| 使用例 | 説明 |
|---|---|
| `gohan new [--type=post] [--title=<t>] <slug>` | `content/posts/<slug>.md` を作成 |
| `gohan new --type=page [--title=<t>] <slug>` | `content/pages/<slug>.md` を作成 |
| `gohan new --type=tutorial intro` | `archetypes/tutorial.md` があれば、それを用いて `content/tutorial/intro.md` を作成 |
| `gohan new --archetype=news <slug>` | `archetypes/news.md` を用いて `content/posts/<slug>.md` を作成 |

### Archetypes（カスタムテンプレート）

プロジェクトルートに `archetypes/<name>.md` が存在する場合、`gohan new` は Go 標準 `text/template` でそれをレンダリングします。利用可能な変数:

| 変数 | 説明 |
|---|---|
| `{{ .Title }}` | 解決されたタイトル（`--title` or スラッグから生成） |
| `{{ .Date }}` | 当日日付（`YYYY-MM-DD`） |
| `{{ .Slug }}` | slug 引数 |
| `{{ .Type }}` | `--type` の値 |

Archetype ファイルが無い場合は、`post` と `page` についてのみ組み込みテンプレートが使用されます。それ以外の `--type` 値では、対応する archetype ファイルが必須となります。

`--title` を省略した場合、スラッグからタイトルが自動生成されます（例: `my-post` → `My Post`）。

---

## `gohan serve`

ライブリロード機能付きのローカル HTTP 開発サーバーを起動します。

- デフォルトアドレス: `http://127.0.0.1:1313`
- `content/`・`themes/`・`assets/` のファイル変更を監視
- ファイル変更時に自動で再ビルドしてブラウザをリロード

---

## `gohan version`

インストールされたバイナリのバージョン・コミットハッシュ・ビルド日時を表示します。
