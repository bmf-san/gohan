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
	if err := GenerateSitemap(dir, "https://example.com", makeArticles(), nil, nil, model.Config{}); err != nil {
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
	if err := GenerateSitemap(dir, "https://example.com", nil, nil, nil, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap empty: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	if !strings.Contains(string(data), "urlset") {
		t.Errorf("expected empty urlset")
	}
}

func TestGenerateSitemap_WellFormedXML(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", makeArticles(), nil, nil, model.Config{}); err != nil {
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

func TestGenerateSitemap_VirtualPages(t *testing.T) {
	dir := t.TempDir()
	vps := []*model.VirtualPage{
		{URL: "/bookshelf/"},
		{URL: "/ja/bookshelf/"},
	}
	if err := GenerateSitemap(dir, "https://example.com", nil, vps, nil, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	if !strings.Contains(s, "https://example.com/bookshelf/") {
		t.Errorf("expected /bookshelf/ in sitemap:\n%s", s)
	}
	if !strings.Contains(s, "https://example.com/ja/bookshelf/") {
		t.Errorf("expected /ja/bookshelf/ in sitemap:\n%s", s)
	}
}

func TestGenerateSitemap_LastmodFromField(t *testing.T) {
	date := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	lastmod := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Slug:    "my-post",
				Date:    date,
				LastMod: lastmod,
			}},
			URL: "/posts/my-post/",
		},
	}
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", articles, nil, nil, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	if !strings.Contains(s, "<lastmod>2026-03-15</lastmod>") {
		t.Errorf("expected lastmod field to be used, got:\n%s", s)
	}
	if strings.Contains(s, "<lastmod>2020-01-01</lastmod>") {
		t.Errorf("expected date not to be used when LastMod is set, got:\n%s", s)
	}
}

func TestGenerateSitemap_LastmodFallsBackToDate(t *testing.T) {
	date := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{Slug: "no-lastmod", Date: date}},
			URL:     "/posts/no-lastmod/",
		},
	}
	dir := t.TempDir()
	if err := GenerateSitemap(dir, "https://example.com", articles, nil, nil, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	if !strings.Contains(string(data), "<lastmod>2020-06-15</lastmod>") {
		t.Errorf("expected fallback to date, got:\n%s", data)
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
	if err := GenerateSitemap(dir, "https://example.com", articles, nil, nil, model.Config{}); err != nil {
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
	if err := GenerateSitemap(dir, "https://example.com", nil, nil, nil, cfg); err != nil {
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
	if err := GenerateSitemap(dir, "https://example.com", articles, nil, nil, model.Config{}); err != nil {
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
	if err := GenerateSitemap(dir, "https://example.com", articles, nil, nil, cfg); err != nil {
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
	if err := GenerateSitemap(dir, "https://example.com", articles, nil, nil, cfg); err != nil {
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
	if err := GenerateSitemap(dir, "https://example.com", articles, nil, nil, model.Config{}); err != nil {
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

func TestGenerateSitemap_ExtraURLs(t *testing.T) {
	dir := t.TempDir()
	extra := []string{"/tags/go/", "/categories/architecture/", "/archives/2024/", "/archives/2024/01/"}
	if err := GenerateSitemap(dir, "https://example.com", nil, nil, extra, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	for _, u := range extra {
		want := "https://example.com" + u
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in sitemap:\n%s", want, s)
		}
	}
}

func TestTaxonomyURLs_I18n(t *testing.T) {
	cfg := model.Config{}
	cfg.I18n.DefaultLocale = "en"
	cfg.I18n.Locales = []string{"en", "ja"}

	date := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	site := &model.Site{
		Config:     cfg,
		Tags:       []model.Taxonomy{{Name: "Go"}},
		Categories: []model.Taxonomy{{Name: "Architecture"}},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Slug:       "en-post",
					Date:       date,
					Tags:       []string{"Go"},
					Categories: []string{"Architecture"},
				}},
				Locale: "en",
			},
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Slug:       "ja-post",
					Date:       date,
					Tags:       []string{"Go"},
					Categories: []string{"Architecture"},
				}},
				Locale: "ja",
			},
		},
	}

	urls := TaxonomyURLs(site, cfg)
	wantContains := []string{
		"/tags/go/",
		"/ja/tags/go/",
		"/categories/architecture/",
		"/ja/categories/architecture/",
		"/archives/2024/",
		"/ja/archives/2024/",
		"/archives/2024/03/",
		"/ja/archives/2024/03/",
	}
	urlSet := map[string]bool{}
	for _, u := range urls {
		urlSet[u] = true
	}
	for _, want := range wantContains {
		if !urlSet[want] {
			t.Errorf("TaxonomyURLs missing %q; got: %v", want, urls)
		}
	}
}

func writePage(t *testing.T, outDir, urlPath, head, body string) {
	t.Helper()
	dir := filepath.Join(outDir, filepath.FromSlash(strings.Trim(urlPath, "/")))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := "<!DOCTYPE html><html><head><title>t</title>" + head + "</head><body>" + body + "</body></html>"
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateSitemap_SkipsNoindexPages(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "/tags/thin/", `<meta name="robots" content="noindex,follow">`, "")
	writePage(t, dir, "/tags/go/", `<meta name="description" content="noindex is mentioned here">`, "")
	writePage(t, dir, "/categories/mentions/", "", "<p>this page talks about noindex</p>")
	writePage(t, dir, "/bookshelf/", `<META NAME="ROBOTS" CONTENT="NONE">`, "")
	writePage(t, dir, "/posts/hidden/", `<meta content="noindex" name="robots"/>`, "")
	writePage(t, dir, "/posts/visible/", `<meta name="robots" content="index,follow">`, "")

	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Slug: "hidden", Date: date}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{Slug: "visible", Date: date}}},
	}
	vps := []*model.VirtualPage{{URL: "/bookshelf/"}}
	extra := []string{"/tags/thin/", "/tags/go/", "/categories/mentions/", "/archives/2024/"}
	if err := GenerateSitemap(dir, "https://example.com", articles, vps, extra, model.Config{}); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	for _, u := range []string{"/tags/thin/", "/bookshelf/", "/posts/hidden/"} {
		if strings.Contains(s, "https://example.com"+u+"<") {
			t.Errorf("noindex page %q must not be in the sitemap:\n%s", u, s)
		}
	}
	// Pages without a robots noindex directive stay, including pages that
	// have not been rendered (missing files are treated as indexable).
	for _, u := range []string{"/tags/go/", "/categories/mentions/", "/archives/2024/", "/posts/visible/"} {
		if !strings.Contains(s, "https://example.com"+u+"<") {
			t.Errorf("expected %q in the sitemap:\n%s", u, s)
		}
	}
}

func TestGenerateSitemap_NoindexTranslationIsNotAnAlternate(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "/posts/hello/", "", "")
	writePage(t, dir, "/ja/posts/hello/", `<meta name="googlebot" content="noindex">`, "")

	cfg := model.Config{}
	cfg.I18n.DefaultLocale = "en"
	cfg.I18n.Locales = []string{"en", "ja"}
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	en := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{Slug: "hello", Date: date}}, Locale: "en", URL: "/posts/hello/"}
	ja := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{Slug: "hello", Date: date}}, Locale: "ja", URL: "/ja/posts/hello/"}
	en.Translations = []model.LocaleRef{{Locale: "ja", URL: "/ja/posts/hello/"}}
	ja.Translations = []model.LocaleRef{{Locale: "en", URL: "/posts/hello/"}}
	if err := GenerateSitemap(dir, "https://example.com", []*model.ProcessedArticle{en, ja}, nil, nil, cfg); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	s := string(data)
	if !strings.Contains(s, "<loc>https://example.com/posts/hello/</loc>") {
		t.Errorf("indexable article missing:\n%s", s)
	}
	if strings.Contains(s, "https://example.com/ja/posts/hello/") {
		t.Errorf("noindex translation must not appear as a URL or hreflang alternate:\n%s", s)
	}
}

func TestPageIsNoindex(t *testing.T) {
	cases := []struct {
		name string
		head string
		body string
		want bool
	}{
		{"robots noindex", `<meta name="robots" content="noindex">`, "", true},
		{"robots noindex follow", `<meta name="robots" content="noindex, follow">`, "", true},
		{"attribute order reversed", `<meta content="noindex,nofollow" name="robots">`, "", true},
		{"uppercase none", `<META NAME=ROBOTS CONTENT="NONE">`, "", true},
		{"googlebot", `<meta name='googlebot' content='noindex'>`, "", true},
		{"robots index", `<meta name="robots" content="index,follow">`, "", false},
		{"other meta mentions noindex", `<meta name="description" content="about noindex">`, "", false},
		{"robots-like name", `<meta name="robotsx" content="noindex">`, "", false},
		{"meta in body is ignored", "", `<meta name="robots" content="noindex">`, false},
		{"no meta", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writePage(t, dir, "/p/", tc.head, tc.body)
			if got := pageIsNoindex(dir, "/p/"); got != tc.want {
				t.Errorf("pageIsNoindex = %v, want %v", got, tc.want)
			}
		})
	}
	if pageIsNoindex(t.TempDir(), "/missing/") {
		t.Error("a missing page must be treated as indexable")
	}
}

func TestPageIsNoindex_ResolvesRootAndFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><head><meta name="robots" content="noindex"></head></html>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "page.html"), []byte(`<html><head><meta name="robots" content="noindex"></head></html>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if !pageIsNoindex(dir, "/") {
		t.Error("expected / to resolve to index.html")
	}
	if !pageIsNoindex(dir, "/page.html?x=1#frag") {
		t.Error("expected /page.html to resolve to the file, ignoring query and fragment")
	}
}
