# gohan

[![CI](https://github.com/bmf-san/gohan/actions/workflows/ci.yml/badge.svg)](https://github.com/bmf-san/gohan/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmf-san/gohan.svg)](https://pkg.go.dev/github.com/bmf-san/gohan)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Go で実装されたシンプルな静的サイトジェネレーター（SSG）。差分ビルド・シンタックスハイライト・Mermaid 図・ライブリロード開発サーバーを備えます。

> English version: [README.md](README.md)

---

## 特徴

- **差分ビルド** — 変更されたファイルのみを再生成し、ビルド時間を最小化
- **Markdown + Front Matter** — GFM (GitHub Flavored Markdown) 対応
- **シンタックスハイライト** — [chroma](https://github.com/alecthomas/chroma) によるコードブロックのスタイリング
- **Mermaid 図** — ` + "`mermaid`" + ` フェンスコードブロックをインタラクティブな図に変換
- **タクソノミー** — タグ・カテゴリーページを自動生成
- **Atom フィード / サイトマップ** — `atom.xml`・`sitemap.xml` を自動生成
- **ライブリロード開発サーバー** — `gohan serve` でファイル変更を検知してブラウザを自動リロード
- **カスタマイズ可能なテーマ** — Go `html/template` による完全制御

---

## インストール

```bash
go install github.com/bmf-san/gohan/cmd/gohan@latest
```

ソースからビルドする場合:

```bash
git clone https://github.com/bmf-san/gohan.git
cd gohan
make install
```

ビルド済みバイナリは [GitHub Releases](https://github.com/bmf-san/gohan/releases) からダウンロードできます。

---

## クイックスタート

```bash
# 1. プロジェクトディレクトリを作成
mkdir myblog && cd myblog

# 2. config.yaml を作成（全オプションは docs/guide/configuration.ja.md を参照）
cat > config.yaml << 'EOF'
site:
  title: My Blog
  base_url: https://example.com
  language: ja
build:
  content_dir: content
  output_dir: public
theme:
  name: default
EOF

# 3. 最初の記事を作成
gohan new post --slug=hello-world --title="Hello, World!"

# 4. サイトをビルド
gohan build

# 5. 開発サーバーでプレビュー
gohan serve   # http://127.0.0.1:1313 を開く
```

---

## ユーザーガイド

詳細なドキュメントは **[docs/guide/](docs/guide/README.ja.md)** を参照してください:

| ガイド | 内容 |
|---|---|
| [Getting Started](docs/guide/getting-started.ja.md) | インストール、最初のサイト作成、ビルド、プレビュー |
| [Configuration](docs/guide/configuration.ja.md) | `config.yaml` の全フィールドと Front Matter |
| [Templates](docs/guide/templates.ja.md) | テーマテンプレート・変数・組み込み関数 |
| [Taxonomy](docs/guide/taxonomy.ja.md) | タグ・カテゴリー・アーカイブページ |

---

## CLI リファレンス

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

## 開発者向け

```bash
make test      # テストを実行（race detector 有効）
make coverage  # テストを実行してカバレッジを表示
make lint      # golangci-lint を実行
make build     # gohan バイナリをビルド
make clean     # ビルド成果物を削除
```

設計ドキュメントは [docs/DESIGN_DOC.ja.md](docs/DESIGN_DOC.ja.md) を参照してください。

---

## ライセンス

[MIT](LICENSE)
