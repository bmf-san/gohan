# TODO

## 開発方針

### PR 運用ルール

- Issue ごとに 1 PR を作成する
- PR の base branch は **前の PR のブランチ** に設定する（main ではない）
  - こうすることで各 PR の diff がそのフェーズの変更のみになりレビューしやすい
- マージは必ず **順番通り** に行う（base branch が存在する状態でマージする）
- PR のラベル: 既存の `enhancement` を使用
- PR のアサイン: `bmf-san`
- PR のマージは自分（bmf-san）が行う

### ブランチ命名規則

```
feat/phase-{フェーズ番号}-{概要}
```

### コミットメッセージ

```
feat: {概要} (#{issue番号})
```

---

## PR チェーン（全フェーズ）

| フェーズ | Issue | ブランチ | Base ブランチ | PR | 状態 |
|---|---|---|---|---|---|
| 0-1 | #4  | `feat/phase-0-1-go-module`        | `main`                            | #26 | ✅ マージ済み |
| 0-2 | #6  | `feat/phase-0-2-ci`               | `feat/phase-0-1-go-module`        | #27 | ✅ マージ済み |
| 1   | #7  | `feat/phase-1-core-interfaces`    | `feat/phase-0-2-ci`               | #28 | 🔄 レビュー待ち |
| 2   | #8  | `feat/phase-2-config-loader`      | `feat/phase-1-core-interfaces`    | -   | ⏳ 未着手 |
| 3-1 | #9  | `feat/phase-3-1-markdown-parser`  | `feat/phase-2-config-loader`      | -   | ⏳ 未着手 |
| 3-2 | #10 | `feat/phase-3-2-frontmatter-parser` | `feat/phase-3-1-markdown-parser` | -   | ⏳ 未着手 |
| 4   | #11 | `feat/phase-4-template-engine`    | `feat/phase-3-2-frontmatter-parser` | - | ⏳ 未着手 |
| 5-1 | #12 | `feat/phase-5-1-dependency-graph` | `feat/phase-4-template-engine`    | -   | ⏳ 未着手 |
| 5-2 | #13 | `feat/phase-5-2-taxonomy`         | `feat/phase-5-1-dependency-graph` | -   | ⏳ 未着手 |
| 6-1 | #14 | `feat/phase-6-1-html-generator`   | `feat/phase-5-2-taxonomy`         | -   | ⏳ 未着手 |
| 6-2 | #15 | `feat/phase-6-2-sitemap-feed`     | `feat/phase-6-1-html-generator`   | -   | ⏳ 未着手 |
| 7-1 | #16 | `feat/phase-7-1-git-diff`         | `feat/phase-6-2-sitemap-feed`     | -   | ⏳ 未着手 |
| 7-2 | #17 | `feat/phase-7-2-cache`            | `feat/phase-7-1-git-diff`         | -   | ⏳ 未着手 |
| 8-1 | #18 | `feat/phase-8-1-build-command`    | `feat/phase-7-2-cache`            | -   | ⏳ 未着手 |
| 8-2 | #19 | `feat/phase-8-2-new-command`      | `feat/phase-8-1-build-command`    | -   | ⏳ 未着手 |
| 8-3 | #20 | `feat/phase-8-3-serve-command`    | `feat/phase-8-2-new-command`      | -   | ⏳ 未着手 |
| 9   | #21 | `feat/phase-9-dev-server`         | `feat/phase-8-3-serve-command`    | -   | ⏳ 未着手 |
| 10-1 | #22 | `feat/phase-10-1-syntax-highlight` | `feat/phase-9-dev-server`        | -   | ⏳ 未着手 |
| 10-2 | #23 | `feat/phase-10-2-mermaid`         | `feat/phase-10-1-syntax-highlight` | -  | ⏳ 未着手 |
| 11  | #24 | `feat/phase-11-goreleaser`        | `feat/phase-10-2-mermaid`         | -   | ⏳ 未着手 |
| 12  | #25 | `feat/phase-12-test-infra`        | `feat/phase-11-goreleaser`        | -   | ⏳ 未着手 |

---

## 各フェーズの概要

| フェーズ | 内容 |
|---|---|
| 0-1 | Go module 初期化・ディレクトリ構造・`.gitignore` |
| 0-2 | CI (GitHub Actions): golangci-lint + go test -race + カバレッジ 80% チェック |
| 1   | コアインターフェース定義・`internal/model` パッケージ（全共有データ型） |
| 2   | Config loader: `config.yaml` の読み込み・バリデーション |
| 3-1 | Markdown パーサー: goldmark ベースの HTML 変換 |
| 3-2 | Front Matter パーサー: YAML メタデータ抽出 |
| 4   | テンプレートエンジン: `html/template` ベース |
| 5-1 | 依存グラフ構築・プロセッサー実装 |
| 5-2 | タクソノミーシステム (tags / categories) |
| 6-1 | HTML 出力ジェネレーター |
| 6-2 | sitemap.xml・atom.xml 生成 |
| 7-1 | git diff ベースの差分検出 |
| 7-2 | ビルドキャッシュ (`.gohan/cache/manifest.json`) |
| 8-1 | `gohan build` コマンド |
| 8-2 | `gohan new` コマンド |
| 8-3 | `gohan serve` コマンド |
| 9   | Dev サーバー: fsnotify + SSE ライブリロード |
| 10-1 | シンタックスハイライト (Chroma) |
| 10-2 | Mermaid ダイアグラムサポート |
| 11  | GoReleaser によるリリース自動化 |
| 12  | テストインフラ整備 |
