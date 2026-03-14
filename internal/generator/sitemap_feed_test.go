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
	if err := GenerateSitemap(dir, "https://example.com", makeArticles(), model.Config{}); err != nil {
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
	if err := GenerateSitemap(dir, "https://example.com", nil, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap empty: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	if !strings.Contains(string(data), "urlset") {
		t.Errorf("expected empty urlset")
	}
}

func TestGenerateSitemap_WellFormedXML(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", makeArticles(), model.Config{}); err != nil {
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
	if err := GenerateFeeds(dir, "https://example.com", "My Blog", makeArticles(), model.Config{}); err != nil {
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
	if err := GenerateFeeds(dir, "https://example.com", "Blog", makeArticles(), model.Config{}); err != nil {
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
	if err := GenerateFeeds(dir, "https://example.com", "Blog", makeArticles(), model.Config{}); err != nil {
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
	if err := GenerateFeeds(dir, "https://example.com", "Blog", articles, model.Config{}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "feed.xml"))
	if !strings.Contains(string(data), "hello-world") {
		t.Errorf("expected slugified title in feed:\n%s", data)
	}
}

func TestGenerateSitemap_HreflangAlternates(t *testing.T) {
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{
			Article:      model.Article{FrontMatter: model.FrontMatter{Slug: "hello", Date: newer}},
			Locale:       "en",
			URL:          "/posts/hello/",
			Translations: []model.LocaleRef{{Locale: "ja", URL: "/ja/posts/hello/"}},
		},
		{
			Article:      model.Article{FrontMatter: model.FrontMatter{Slug: "hello", Date: newer}},
			Locale:       "ja",
			URL:          "/ja/posts/hello/",
			Translations: []model.LocaleRef{{Locale: "en", URL: "/posts/hello/"}},
		},
	}
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", articles, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	if !strings.Contains(s, `xmlns:xhtml=`) {
		t.Error("expected xhtml namespace in sitemap")
	}
	if !strings.Contains(s, `hreflang="en"`) || !strings.Contains(s, `hreflang="ja"`) {
		t.Errorf("missing hreflang attributes:\n%s", s)
	}
	if !strings.Contains(s, "/ja/posts/hello/") {
		t.Errorf("missing ja URL in sitemap:\n%s", s)
	}
}

func TestGenerateSitemap_I18nIndexPages(t *testing.T) {
	dir := t.TempDir()
	cfg := model.Config{}
	cfg.I18n.DefaultLocale = "en"
	cfg.I18n.Locales = []string{"en", "ja"}
	if err := GenerateSitemap(dir, "https://example.com", nil, cfg); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	if !strings.Contains(s, "https://example.com/") {
		t.Errorf("expected EN index URL https://example.com/ in sitemap:\n%s", s)
	}
	if !strings.Contains(s, "https://example.com/ja/") {
		t.Errorf("expected JA index URL https://example.com/ja/ in sitemap:\n%s", s)
	}
}

func TestGenerateSitemap_UsesPrecomputedURL(t *testing.T) {
	dir := t.TempDir()
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Slug: "slug", Date: time.Now()}},
			URL:     "/ja/posts/my-url/",
		},
	}
	if err := GenerateSitemap(dir, "https://example.com", articles, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	if !strings.Contains(s, "/ja/posts/my-url/") {
		t.Errorf("sitemap should use pre-computed URL:\n%s", s)
	}
	if strings.Contains(s, "/posts/slug/") {
		t.Errorf("sitemap should NOT use slug when URL is set:\n%s", s)
	}
}

func TestGenerateFeeds_I18nUsesPrecomputedURL(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Title: "Japanese Post",
				Slug:  "ja-post",
				Date:  date,
			}},
			URL:    "/ja/posts/ja-post/",
			Locale: "ja",
		},
	}
	if err := GenerateFeeds(dir, "https://example.com", "Blog", articles, model.Config{}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		s := string(data)
		if !strings.Contains(s, "/ja/posts/ja-post/") {
			t.Errorf("%s: expected locale-aware URL /ja/posts/ja-post/:\n%s", name, s)
		}
		if strings.Contains(s, "https://example.com/posts/ja-post/") {
			t.Errorf("%s: should NOT fall back to /posts/ when URL is set:\n%s", name, s)
		}
	}
}

func TestGenerateFeeds_NoURLFallsBackToSlug(t *testing.T) {
	dir := t.TempDir()
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Title: "Plain Post",
				Slug:  "plain-post",
				Date:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}},
			// URL is empty (no i18n)
		},
	}
	if err := GenerateFeeds(dir, "https://example.com", "Blog", articles, model.Config{}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		if !strings.Contains(string(data), "/posts/plain-post/") {
			t.Errorf("%s: expected /posts/plain-post/ fallback:\n%s", name, data)
		}
	}
}

// TestGenerateSitemap_XDefault_DefaultLocale verifies that x-default points to
// the self URL when the article's locale is the site's default locale.
func TestGenerateSitemap_XDefault_DefaultLocale(t *testing.T) {
	cfg := model.Config{}
	cfg.I18n.DefaultLocale = "en"
	articles := []*model.ProcessedArticle{
		{
			Article:      model.Article{FrontMatter: model.FrontMatter{Slug: "hello", Date: time.Now()}},
			Locale:       "en",
			URL:          "/posts/hello/",
			Translations: []model.LocaleRef{{Locale: "ja", URL: "/ja/posts/hello/"}},
		},
	}
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", articles, cfg); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	want := `hreflang="x-default" href="https://example.com/posts/hello/"`
	if !strings.Contains(s, want) {
		t.Errorf("expected x-default pointing to EN (self) URL\nwant substring: %s\ngot:\n%s", want, s)
	}
}

// TestGenerateSitemap_XDefault_NonDefaultLocale verifies that x-default points
// to the default-locale translation URL when the article itself is non-default.
func TestGenerateSitemap_XDefault_NonDefaultLocale(t *testing.T) {
	cfg := model.Config{}
	cfg.I18n.DefaultLocale = "en"
	articles := []*model.ProcessedArticle{
		{
			Article:      model.Article{FrontMatter: model.FrontMatter{Slug: "hello", Date: time.Now()}},
			Locale:       "ja",
			URL:          "/ja/posts/hello/",
			Translations: []model.LocaleRef{{Locale: "en", URL: "/posts/hello/"}},
		},
	}
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", articles, cfg); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	want := `hreflang="x-default" href="https://example.com/posts/hello/"`
	if !strings.Contains(s, want) {
		t.Errorf("expected x-default pointing to EN translation URL\nwant substring: %s\ngot:\n%s", want, s)
	}
}

// TestGenerateSitemap_XDefault_NotEmittedWithoutConfig verifies that no
// x-default link is emitted when DefaultLocale is not configured.
func TestGenerateSitemap_XDefault_NotEmittedWithoutConfig(t *testing.T) {
	articles := []*model.ProcessedArticle{
		{
			Article:      model.Article{FrontMatter: model.FrontMatter{Slug: "hello", Date: time.Now()}},
			Locale:       "en",
			URL:          "/posts/hello/",
			Translations: []model.LocaleRef{{Locale: "ja", URL: "/ja/posts/hello/"}},
		},
	}
	dir := t.TempDir()
	// model.Config{} has an empty DefaultLocale
	if err := GenerateSitemap(dir, "https://example.com", articles, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	if strings.Contains(string(data), "x-default") {
		t.Errorf("x-default should NOT be emitted when DefaultLocale is not configured:\n%s", data)
	}
}

// TestGenerateFeeds_I18n_LocaleFilter verifies that when i18n is configured:
//   - root feed.xml / atom.xml contain only default-locale articles
//   - locale subdirectory feeds contain only that locale's articles
func TestGenerateFeeds_I18n_LocaleFilter(t *testing.T) {
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cfg := model.Config{}
	cfg.I18n.DefaultLocale = "en"
	cfg.I18n.Locales = []string{"en", "ja"}

	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Title: "EN Post", Slug: "en-post", Date: date}},
			URL:     "/posts/en-post/",
			Locale:  "en",
		},
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Title: "JA Post", Slug: "ja-post", Date: date}},
			URL:     "/ja/posts/ja-post/",
			Locale:  "ja",
		},
	}
	dir := t.TempDir()
	if err := GenerateFeeds(dir, "https://example.com", "Blog", articles, cfg); err != nil {
		t.Fatalf("GenerateFeeds: %v", err)
	}

	// Root feeds (EN only)
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		s := string(data)
		if !strings.Contains(s, "EN Post") {
			t.Errorf("root %s: expected EN Post", name)
		}
		if strings.Contains(s, "JA Post") {
			t.Errorf("root %s: must NOT contain JA Post", name)
		}
	}

	// JA locale feeds
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, "ja", name))
		s := string(data)
		if !strings.Contains(s, "JA Post") {
			t.Errorf("ja/%s: expected JA Post", name)
		}
		if strings.Contains(s, "EN Post") {
			t.Errorf("ja/%s: must NOT contain EN Post", name)
		}
		// Article link must not double the locale prefix
		if strings.Contains(s, "/ja/ja/") {
			t.Errorf("ja/%s: double locale prefix /ja/ja/ detected:\n%s", name, s)
		}
		// Channel link should point to locale root (with trailing slash)
		if !strings.Contains(s, "https://example.com/ja/") {
			t.Errorf("ja/%s: channel link should be https://example.com/ja/", name)
		}
	}
}

// TestGenerateFeeds_I18n_NoDefaultLocale verifies backward compatibility:
// when no locales are configured, a single combined feed is written.
func TestGenerateFeeds_I18n_NoDefaultLocale(t *testing.T) {
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Title: "EN Post", Slug: "en-post", Date: date}},
			Locale:  "en",
		},
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Title: "JA Post", Slug: "ja-post", Date: date}},
			Locale:  "ja",
		},
	}
	dir := t.TempDir()
	if err := GenerateFeeds(dir, "https://example.com", "Blog", articles, model.Config{}); err != nil {
		t.Fatalf("GenerateFeeds: %v", err)
	}
	// Both articles must appear in root feed
	for _, name := range []string{"feed.xml", "atom.xml"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		s := string(data)
		if !strings.Contains(s, "EN Post") || !strings.Contains(s, "JA Post") {
			t.Errorf("%s: expected both EN and JA posts in non-i18n feed:\n%s", name, s)
		}
	}
	// No locale subdirectory should exist
	if _, err := os.Stat(filepath.Join(dir, "ja")); err == nil {
		t.Error("ja/ subdirectory should NOT be created when i18n is not configured")
	}
}
