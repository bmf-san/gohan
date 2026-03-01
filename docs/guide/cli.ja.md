# CLI リファレンス

> English version: [cli.md](cli.md)

---

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `gohan build` | サイトをビルド（デフォルトで差分ビルド） |
| `gohan build --full` | フルビルドを強制実行 |
| `gohan build --dry-run` | ファイルを書き出さずにビルドをシミュレート |
| `gohan new post --slug=<s> --title=<t>` | 新規記事スケルトンを作成 |
| `gohan new page --slug=<s> --title=<t>` | 新規ページスケルトンを作成 |
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

---

## `gohan new`

Front Matter が事前入力されたコンテンツスケルトンを作成します。

**サブコマンド**

| サブコマンド | 説明 |
|---|---|
| `gohan new post --slug=<s> [--title=<t>]` | `content/posts/<slug>.md` を作成 |
| `gohan new page --slug=<s> [--title=<t>]` | `content/pages/<slug>.md` を作成 |

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
