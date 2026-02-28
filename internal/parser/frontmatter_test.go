package parser

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	return path
}

func TestFileParser_Parse_WithFrontMatter(t *testing.T) {
	dir := t.TempDir()
	src := "---\ntitle: Hello World\ndraft: false\ntags:\n  - go\n  - ssg\n---\n# Hello\n\nBody text.\n"
	path := writeFile(t, dir, "post.md", src)
	p := NewFileParser()
	a, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.FrontMatter.Title != "Hello World" {
		t.Errorf("Title: got %q", a.FrontMatter.Title)
	}
	if a.FrontMatter.Draft {
		t.Error("Draft: want false")
	}
	if len(a.FrontMatter.Tags) != 2 || a.FrontMatter.Tags[0] != "go" {
		t.Errorf("Tags: got %v", a.FrontMatter.Tags)
	}
	if a.FilePath != path {
		t.Errorf("FilePath mismatch: got %q", a.FilePath)
	}
	if a.RawContent == "" {
		t.Error("RawContent should not be empty")
	}
	if a.LastModified.IsZero() {
		t.Error("LastModified should not be zero")
	}
}

func TestFileParser_Parse_NoFrontMatter(t *testing.T) {
	dir := t.TempDir()
	body := "# No Front Matter\n\nContent.\n"
	path := writeFile(t, dir, "plain.md", body)
	p := NewFileParser()
	a, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.FrontMatter.Title != "" {
		t.Errorf("expected empty title, got %q", a.FrontMatter.Title)
	}
	if a.RawContent != body {
		t.Errorf("RawContent mismatch: got %q, want %q", a.RawContent, body)
	}
}

func TestFileParser_Parse_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty.md", "")
	p := NewFileParser()
	a, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.FrontMatter.Title != "" {
		t.Error("expected empty title for empty file")
	}
}

func TestFileParser_Parse_NoClosingDelimiter(t *testing.T) {
	dir := t.TempDir()
	content := "---\ntitle: Incomplete\n# No closing"
	path := writeFile(t, dir, "incomplete.md", content)
	p := NewFileParser()
	a, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.FrontMatter.Title != "" {
		t.Errorf("expected empty title, got %q", a.FrontMatter.Title)
	}
	if string(a.RawContent) != content {
		t.Error("RawContent should be entire file content")
	}
}

func TestFileParser_Parse_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.md", "---\ntitle: [unclosed\n---\nbody\n")
	p := NewFileParser()
	_, err := p.Parse(path)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestFileParser_Parse_FileNotFound(t *testing.T) {
	p := NewFileParser()
	_, err := p.Parse("/nonexistent/path/post.md")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestFileParser_Parse_AllFrontMatterFields(t *testing.T) {
	dir := t.TempDir()
	src := "---\ntitle: Full Post\ndate: 2024-01-15T00:00:00Z\ndraft: true\ntags: [a, b]\ncategories: [news]\ndescription: A description\nauthor: Alice\nslug: full-post\ntemplate: custom\n---\ncontent\n"
	path := writeFile(t, dir, "full.md", src)
	p := NewFileParser()
	a, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm := a.FrontMatter
	if fm.Title != "Full Post" {
		t.Errorf("Title: got %q", fm.Title)
	}
	want := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !fm.Date.Equal(want) {
		t.Errorf("Date: got %v, want %v", fm.Date, want)
	}
	if !fm.Draft {
		t.Error("Draft: want true")
	}
	if len(fm.Tags) != 2 {
		t.Errorf("Tags: got %v", fm.Tags)
	}
	if len(fm.Categories) != 1 || fm.Categories[0] != "news" {
		t.Errorf("Categories: got %v", fm.Categories)
	}
	if fm.Description != "A description" {
		t.Errorf("Description: got %q", fm.Description)
	}
	if fm.Author != "Alice" {
		t.Errorf("Author: got %q", fm.Author)
	}
	if fm.Slug != "full-post" {
		t.Errorf("Slug: got %q", fm.Slug)
	}
	if fm.Template != "custom" {
		t.Errorf("Template: got %q", fm.Template)
	}
}

func TestFileParser_ParseAll_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.md", "---\ntitle: A\n---\nbody a\n")
	writeFile(t, dir, "b.md", "---\ntitle: B\n---\nbody b\n")
	writeFile(t, dir, "c.markdown", "---\ntitle: C\n---\nbody c\n")
	writeFile(t, dir, "ignored.txt", "not markdown")
	p := NewFileParser()
	articles, err := p.ParseAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 3 {
		t.Errorf("expected 3 articles, got %d", len(articles))
	}
}

func TestFileParser_ParseAll_SubDirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "root.md", "# Root\n")
	writeFile(t, sub, "nested.md", "# Nested\n")
	p := NewFileParser()
	articles, err := p.ParseAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(articles))
	}
}

func TestFileParser_ParseAll_DirNotFound(t *testing.T) {
	p := NewFileParser()
	_, err := p.ParseAll("/nonexistent/content/dir")
	if err == nil {
		t.Error("expected error for missing directory, got nil")
	}
}

func TestFileParser_ParseAll_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	p := NewFileParser()
	articles, err := p.ParseAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 0 {
		t.Errorf("expected 0 articles, got %d", len(articles))
	}
}
