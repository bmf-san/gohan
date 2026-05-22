#!/usr/bin/env python3
"""Migrate docs/{guide,features}/*.md and DESIGN_DOC.md into docs/content/{en,ja}/.

Strips leading H1 + immediate blockquote intro + cross-locale link + initial '---'
separator, and prepends YAML front matter.
"""
from __future__ import annotations
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
SRC = ROOT / "docs"
DEST_EN = SRC / "content" / "en"
DEST_JA = SRC / "content" / "ja"

# (src_relative, locale, category, slug, title, description, translation_key)
MIGRATIONS: list[tuple[str, str, str, str, str, str, str]] = [
    # guide
    ("guide/cli.md", "en", "guide", "cli", "CLI Reference",
     "Command reference for the gohan CLI.", "cli"),
    ("guide/cli.ja.md", "ja", "guide", "cli", "CLI リファレンス",
     "gohan CLI のコマンドリファレンス。", "cli"),
    ("guide/configuration.md", "en", "guide", "configuration", "Configuration",
     "Configure gohan via config.yaml: site, theme, i18n, and more.", "configuration"),
    ("guide/configuration.ja.md", "ja", "guide", "configuration", "設定",
     "config.yaml で gohan のサイト・テーマ・i18n などを設定する方法。", "configuration"),
    ("guide/templates.md", "en", "guide", "templates", "Templates",
     "Author and customize templates for layouts, articles, and taxonomies.", "templates"),
    ("guide/templates.ja.md", "ja", "guide", "templates", "テンプレート",
     "レイアウト・記事・タクソノミーのテンプレートを書く方法。", "templates"),
    ("guide/taxonomy.md", "en", "guide", "taxonomy", "Taxonomy",
     "Organize content with categories and tags.", "taxonomy"),
    ("guide/taxonomy.ja.md", "ja", "guide", "taxonomy", "タクソノミー",
     "カテゴリーとタグでコンテンツを整理する。", "taxonomy"),
    # features (i18n is already migrated in Phase 1; skip here)
    ("features/github-source-link.md", "en", "features", "github-source-link",
     "GitHub Source Link", "Link articles back to their source on GitHub.", "github-source-link"),
    ("features/github-source-link.ja.md", "ja", "features", "github-source-link",
     "GitHub ソースリンク", "記事から GitHub のソースへのリンクを生成する。", "github-source-link"),
    ("features/ogp.md", "en", "features", "ogp", "OGP",
     "Auto-generate Open Graph and Twitter Card metadata.", "ogp"),
    ("features/ogp.ja.md", "ja", "features", "ogp", "OGP",
     "Open Graph / Twitter Card メタデータを自動生成。", "ogp"),
    ("features/pagination.md", "en", "features", "pagination", "Pagination",
     "Split listings across pages.", "pagination"),
    ("features/pagination.ja.md", "ja", "features", "pagination", "ページネーション",
     "一覧ページをページ送りで分割する。", "pagination"),
    ("features/plugin-system.md", "en", "features", "plugin-system", "Plugin System",
     "Extend gohan with external plugins.", "plugin-system"),
    ("features/plugin-system.ja.md", "ja", "features", "plugin-system", "プラグインシステム",
     "外部プラグインで gohan を拡張する。", "plugin-system"),
    ("features/related-articles.md", "en", "features", "related-articles", "Related Articles",
     "Surface related articles automatically.", "related-articles"),
    ("features/related-articles.ja.md", "ja", "features", "related-articles", "関連記事",
     "関連記事を自動で表示する。", "related-articles"),
    # design doc
    ("DESIGN_DOC.md", "en", "guide", "design", "Design Document",
     "Architecture and design decisions behind gohan.", "design"),
    ("DESIGN_DOC.ja.md", "ja", "guide", "design", "設計ドキュメント",
     "gohan のアーキテクチャと設計判断。", "design"),
]


def strip_intro(text: str) -> str:
    """Remove leading H1 + immediate blockquote intro + cross-locale link + '---'."""
    lines = text.splitlines(keepends=True)
    i = 0
    n = len(lines)

    # Skip leading blank lines
    while i < n and lines[i].strip() == "":
        i += 1
    # Skip first H1 if present
    if i < n and re.match(r"^# ", lines[i]):
        i += 1
    # Skip a run of blank lines and blockquote-style intro lines
    while i < n and (lines[i].strip() == "" or lines[i].startswith("> ") or lines[i].rstrip() == ">"):
        i += 1
    # Optionally skip a single horizontal-rule separator and trailing blanks
    if i < n and lines[i].rstrip() == "---":
        i += 1
        while i < n and lines[i].strip() == "":
            i += 1
    return "".join(lines[i:])


def front_matter(title: str, description: str, slug: str, category: str, translation_key: str) -> str:
    return (
        "---\n"
        f'title: "{title}"\n'
        f'description: "{description}"\n'
        f'slug: "{slug}"\n'
        "categories:\n"
        f"  - {category}\n"
        f'translation_key: "{translation_key}"\n'
        "---\n\n"
    )


def main() -> int:
    for src_rel, locale, cat, slug, title, desc, tk in MIGRATIONS:
        src = SRC / src_rel
        if not src.exists():
            print(f"SKIP missing: {src_rel}", file=sys.stderr)
            continue
        body = strip_intro(src.read_text(encoding="utf-8"))
        out_dir = (DEST_EN if locale == "en" else DEST_JA) / cat
        out_dir.mkdir(parents=True, exist_ok=True)
        out = out_dir / f"{slug}.md"
        out.write_text(front_matter(title, desc, slug, cat, tk) + body, encoding="utf-8")
        print(f"wrote {out.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
