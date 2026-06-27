package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

// decodeSearchIndex reads and parses a search-index.json file, failing the test
// on any error.
func decodeSearchIndex(t *testing.T, path string) searchIndex {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var idx searchIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("invalid JSON in %s: %v\n%s", path, err, data)
	}
	return idx
}

func TestGenerateSearchIndex_Valid(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateSearchIndex(dir, "https://example.com", makeArticles(), model.Config{}); err != nil {
		t.Fatalf("GenerateSearchIndex: %v", err)
	}

	idx := decodeSearchIndex(t, filepath.Join(dir, "search-index.json"))
	if idx.Count != 2 || len(idx.Articles) != 2 {
		t.Fatalf("expected 2 articles, got count=%d len=%d", idx.Count, len(idx.Articles))
	}
	// makeArticles returns "New Post" (2024-06) and "Old Post" (2024-01);
	// the index must be newest-first.
	if idx.Articles[0].Title != "New Post" || idx.Articles[1].Title != "Old Post" {
		t.Errorf("expected newest-first ordering, got %q then %q",
			idx.Articles[0].Title, idx.Articles[1].Title)
	}
	if idx.Generated == "" {
		t.Error("expected non-empty generated timestamp")
	}
}

func TestGenerateSearchIndex_Empty(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateSearchIndex(dir, "https://example.com", nil, model.Config{}); err != nil {
		t.Fatalf("GenerateSearchIndex empty: %v", err)
	}
	idx := decodeSearchIndex(t, filepath.Join(dir, "search-index.json"))
	if idx.Count != 0 {
		t.Errorf("expected count 0, got %d", idx.Count)
	}
	if idx.Articles == nil {
		t.Error("expected non-nil (empty) articles array")
	}
}

func TestGenerateSearchIndex_Fields(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Title:       "Hello",
				Slug:        "hello",
				Date:        date,
				Description: "a greeting",
				Tags:        []string{"go", "ssg"},
				Categories:  []string{"Programming"},
			}},
			Summary: "summary text",
			URL:     "/posts/hello/",
			Locale:  "en",
		},
	}
	if err := GenerateSearchIndex(dir, "https://example.com", articles, model.Config{}); err != nil {
		t.Fatalf("GenerateSearchIndex: %v", err)
	}

	idx := decodeSearchIndex(t, filepath.Join(dir, "search-index.json"))
	if len(idx.Articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(idx.Articles))
	}
	e := idx.Articles[0]
	if e.Title != "Hello" {
		t.Errorf("title = %q", e.Title)
	}
	if e.URL != "https://example.com/posts/hello/" {
		t.Errorf("url = %q, want absolute URL", e.URL)
	}
	if e.Description != "a greeting" {
		t.Errorf("description = %q", e.Description)
	}
	if e.Summary != "summary text" {
		t.Errorf("summary = %q", e.Summary)
	}
	if len(e.Tags) != 2 || e.Tags[0] != "go" {
		t.Errorf("tags = %v", e.Tags)
	}
	if len(e.Categories) != 1 || e.Categories[0] != "Programming" {
		t.Errorf("categories = %v", e.Categories)
	}
	if e.Locale != "en" {
		t.Errorf("locale = %q", e.Locale)
	}
	if e.Date != "2024-03-14T00:00:00Z" {
		t.Errorf("date = %q", e.Date)
	}
}

func TestGenerateSearchIndex_I18n(t *testing.T) {
	dir := t.TempDir()
	en := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	ja := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Title: "Hello", Slug: "hello", Date: en}},
			URL:     "/posts/hello/",
			Locale:  "en",
		},
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Title: "こんにちは", Slug: "hello", Date: ja}},
			URL:     "/ja/posts/hello/",
			Locale:  "ja",
		},
	}
	cfg := model.Config{}
	cfg.I18n.Locales = []string{"en", "ja"}
	cfg.I18n.DefaultLocale = "en"

	if err := GenerateSearchIndex(dir, "https://example.com", articles, cfg); err != nil {
		t.Fatalf("GenerateSearchIndex: %v", err)
	}

	// Root index = default locale (en) only.
	root := decodeSearchIndex(t, filepath.Join(dir, "search-index.json"))
	if root.Count != 1 || root.Articles[0].Locale != "en" {
		t.Errorf("root index should contain only en articles, got %+v", root.Articles)
	}

	// Per-locale index for ja.
	jaIdx := decodeSearchIndex(t, filepath.Join(dir, "ja", "search-index.json"))
	if jaIdx.Count != 1 || jaIdx.Articles[0].Locale != "ja" {
		t.Errorf("ja index should contain only ja articles, got %+v", jaIdx.Articles)
	}
	if jaIdx.Articles[0].URL != "https://example.com/ja/posts/hello/" {
		t.Errorf("ja url = %q", jaIdx.Articles[0].URL)
	}
}
