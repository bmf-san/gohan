package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

func writeCheckArticle(t *testing.T, root, rel, frontmatter, body string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "---\n" + frontmatter + "\n---\n" + body
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestLintArticles_NoIssues(t *testing.T) {
	contentDir := t.TempDir()
	articles := []*model.Article{{
		FilePath: filepath.Join(contentDir, "posts", "hello.md"),
		FrontMatter: model.FrontMatter{
			Title: "Hello",
			Date:  time.Now(),
			Slug:  "hello",
		},
	}}
	issues := lintArticles(articles, contentDir)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d: %#v", len(issues), issues)
	}
}

func TestLintArticles_MissingTitleAndDate(t *testing.T) {
	contentDir := t.TempDir()
	articles := []*model.Article{{
		FilePath:    filepath.Join(contentDir, "posts", "incomplete.md"),
		FrontMatter: model.FrontMatter{},
	}}
	issues := lintArticles(articles, contentDir)
	kinds := map[string]bool{}
	for _, it := range issues {
		kinds[it.Kind] = true
	}
	if !kinds["missing-title"] {
		t.Errorf("missing-title issue not reported")
	}
	if !kinds["missing-date"] {
		t.Errorf("missing-date issue not reported")
	}
}

func TestLintArticles_DuplicateSlug(t *testing.T) {
	contentDir := t.TempDir()
	now := time.Now()
	articles := []*model.Article{
		{
			FilePath:    filepath.Join(contentDir, "posts", "a.md"),
			FrontMatter: model.FrontMatter{Title: "A", Date: now, Slug: "dup"},
		},
		{
			FilePath:    filepath.Join(contentDir, "posts", "b.md"),
			FrontMatter: model.FrontMatter{Title: "B", Date: now, Slug: "dup"},
		},
	}
	issues := lintArticles(articles, contentDir)
	found := false
	for _, it := range issues {
		if it.Kind == "duplicate-slug" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected duplicate-slug issue, got %#v", issues)
	}
}

func TestLintArticles_OrphanTranslationKey(t *testing.T) {
	contentDir := t.TempDir()
	now := time.Now()
	articles := []*model.Article{{
		FilePath: filepath.Join(contentDir, "en", "posts", "hello.md"),
		FrontMatter: model.FrontMatter{
			Title:          "Hello",
			Date:           now,
			Slug:           "hello",
			TranslationKey: "hello-key",
		},
	}}
	issues := lintArticles(articles, contentDir)
	found := false
	for _, it := range issues {
		if it.Kind == "orphan-translation-key" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected orphan-translation-key issue, got %#v", issues)
	}
}

func TestLintArticles_PairedTranslationKey_NoIssue(t *testing.T) {
	contentDir := t.TempDir()
	now := time.Now()
	articles := []*model.Article{
		{
			FilePath: filepath.Join(contentDir, "en", "posts", "hello.md"),
			FrontMatter: model.FrontMatter{
				Title: "Hello", Date: now, Slug: "hello",
				TranslationKey: "hello-key",
			},
		},
		{
			FilePath: filepath.Join(contentDir, "ja", "posts", "hello.md"),
			FrontMatter: model.FrontMatter{
				Title: "ハロー", Date: now, Slug: "hello",
				TranslationKey: "hello-key",
			},
		},
	}
	issues := lintArticles(articles, contentDir)
	for _, it := range issues {
		if it.Kind == "orphan-translation-key" {
			t.Errorf("paired translation_key flagged as orphan: %#v", it)
		}
	}
}

func TestWriteCheckReport_NoIssues(t *testing.T) {
	var buf bytes.Buffer
	writeCheckReport(&buf, nil)
	if !strings.Contains(buf.String(), "no issues found") {
		t.Errorf("expected ok message, got %q", buf.String())
	}
}

func TestWriteCheckReport_FormatsIssues(t *testing.T) {
	var buf bytes.Buffer
	writeCheckReport(&buf, []checkIssue{
		{File: "posts/a.md", Kind: "missing-title", Message: "no title"},
	})
	out := buf.String()
	if !strings.Contains(out, "posts/a.md") || !strings.Contains(out, "missing-title") {
		t.Errorf("expected formatted issue, got %q", out)
	}
	if !strings.Contains(out, "1 issue(s)") {
		t.Errorf("expected count summary, got %q", out)
	}
}

func TestRunCheck_EndToEnd(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte(
		"site:\n  title: Test\n  base_url: https://example.com\n  language: en\nbuild:\n  content_dir: content\n  output_dir: public\n  assets_dir: assets\n",
	), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	// Two posts with the same slug.
	writeCheckArticle(t, root, "content/posts/a.md", "title: A\ndate: 2024-01-01\nslug: dup", "body")
	writeCheckArticle(t, root, "content/posts/b.md", "title: B\ndate: 2024-01-02\nslug: dup", "body")

	oldWd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	err := runCheck(nil)
	if err == nil {
		t.Fatalf("expected error for duplicate slug, got nil")
	}
	if !strings.Contains(err.Error(), "issue") {
		t.Errorf("expected issue error, got %v", err)
	}
}
