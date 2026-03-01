package generator

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

func makeArticles() []*model.ProcessedArticle {
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Old Post", Slug: "old-post", Date: older}}, Summary: "old"},
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "New Post", Slug: "new-post", Date: newer}}, Summary: "new"},
	}
}

func TestGenerateSitemap_Valid(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", makeArticles()); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	if !strings.Contains(s, "new-post") || !strings.Contains(s, "old-post") {
		t.Errorf("missing slugs:\n%s", s)
	}
	if !strings.Contains(s, "urlset") {
		t.Errorf("missing urlset:\n%s", s)
	}
	if strings.Index(s, "new-post") > strings.Index(s, "old-post") {
		t.Errorf("expected newest-first ordering")
	}
}

func TestGenerateSitemap_Empty(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", nil); err != nil {
		t.Fatalf("GenerateSitemap empty: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	if !strings.Contains(string(data), "urlset") {
		t.Errorf("expected empty urlset")
	}
}

func TestGenerateSitemap_WellFormedXML(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", makeArticles()); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	var v interface{}
	if err := xml.Unmarshal(data, &v); err != nil {
		t.Errorf("invalid XML: %v\n%s", err, data)
	}
}

func TestGenerateFeeds_Valid(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateFeeds(dir, "https://example.com", "My Blog", makeArticles()); err != nil {
		t.Fatalf("GenerateFeeds: %v", err)
	}
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		s := string(data)
		if !strings.Contains(s, "New Post") || !strings.Contains(s, "Old Post") {
			t.Errorf("%s missing titles:\n%s", name, s)
		}
		if !strings.Contains(s, "My Blog") {
			t.Errorf("%s missing site title:\n%s", name, s)
		}
	}
}

func TestGenerateFeeds_NewestFirst(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateFeeds(dir, "https://example.com", "Blog", makeArticles()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		s := string(data)
		if strings.Index(s, "new-post") > strings.Index(s, "old-post") {
			t.Errorf("%s: expected newest-first ordering", name)
		}
	}
}

func TestGenerateFeeds_WellFormedXML(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateFeeds(dir, "https://example.com", "Blog", makeArticles()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		var v interface{}
		if err := xml.Unmarshal(data, &v); err != nil {
			t.Errorf("%s invalid XML: %v", name, err)
		}
	}
}

func TestGenerateFeeds_SlugifiesTitle(t *testing.T) {
	dir := t.TempDir()
	articles := []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Hello World", Date: time.Now()}}},
	}
	if err := GenerateFeeds(dir, "https://example.com", "Blog", articles); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "feed.xml"))
	if !strings.Contains(string(data), "hello-world") {
		t.Errorf("expected slugified title in feed:\n%s", data)
	}
}
