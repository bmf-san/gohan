# gohan ユーザーガイド / User Guide

gohan のすべての使い方・設定・カスタマイズについての公式ドキュメントです。

---

## ガイド一覧

| ドキュメント | 内容 |
|---|---|
| [Getting Started](getting-started.md) | インストール、最初のサイト作成、ビルド、開発サーバー |
| [Configuration](configuration.md) | `config.yaml` の全フィールドリファレンス、Front Matter |
| [Templates](templates.md) | テーマテンプレートの作成・変数・組み込み関数 |
| [Taxonomy](taxonomy.md) | タグ・カテゴリー・アーカイブの管理と活用 |

---

## クイックリファレンス

### よく使うコマンド

```bash
# 新しい記事を作成
gohan new post --slug=my-post --title="記事タイトル"

# サイトをビルド
gohan build

# 差分ビルドをスキップして全記事を再生成
gohan build --full

# ファイルを一切書き出さずにビルドをシミュレート
gohan build --dry-run

# 開発サーバーを起動 (http://127.0.0.1:1313)
gohan serve

# バージョン確認
gohan version
```

### Makefile ターゲット

```bash
make build     # バイナリをビルド
make test      # テストを実行
make coverage  # カバレッジを計測・表示
make lint      # golangci-lint を実行
make clean     # ビルド成果物を削除
make help      # 利用可能なターゲットを表示
```

---

## 設計ドキュメント

gohan の内部設計・アーキテクチャについては [DESIGN_DOC.md](../DESIGN_DOC.md) を参照してください。
