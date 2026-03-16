package bookshelf_test

import (
	"strings"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/plugin/bookshelf"
)

func cfg(enabled bool, tag string) map[string]interface{} {
	return map[string]interface{}{"enabled": enabled, "tag": tag}
}

func TestBookshelf_Name(t *testing.T) {
	b := bookshelf.New()
	if got := b.Name(); got != "bookshelf" {
		t.Errorf("Name() = %q, want %q", got, "bookshelf")
	}
}

func TestBookshelf_Enabled(t *testing.T) {
	b := bookshelf.New()
	if b.Enabled(cfg(true, "")) == false {
		t.Error("Enabled should return true when enabled: true")
	}
	if b.Enabled(cfg(false, "")) == true {
		t.Error("Enabled should return false when enabled: false")
	}
	if b.Enabled(map[string]interface{}{}) == true {
		t.Error("Enabled should return false when key is absent")
	}
}

func TestBookshelf_NoBooks(t *testing.T) {
	b := bookshelf.New()
	site := &model.Site{
		Config: model.Config{
			Site: model.SiteConfig{Language: "en"},
			I18n: model.I18nConfig{DefaultLocale: "en"},
		},
		Articles: []*model.ProcessedArticle{
			{Article: model.Article{FrontMatter: model.FrontMatter{Title: "No books"}}},
		},
	}

	pages, err := b.VirtualPages(site, cfg(true, "test-22"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 0 {
		t.Errorf("expected 0 pages for site with no books, got %d", len(pages))
	}
}

func TestBookshelf_SingleLocale(t *testing.T) {
	b := bookshelf.New()
	date := time.Date(2024, 5, 4, 0, 0, 0, 0, time.UTC)
	site := &model.Site{
		Config: model.Config{
			Site: model.SiteConfig{Language: "en"},
			I18n: model.I18nConfig{DefaultLocale: "en"},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "Perfect Ruby",
					Slug:  "perfect-ruby",
					Date:  date,
					Extra: map[string]interface{}{
						"books": []interface{}{
							map[string]interface{}{"asin": "4774189774", "title": "パーフェクトRuby"},
						},
					},
				}},
				Locale: "en",
				URL:    "/posts/perfect-ruby/",
			},
		},
	}

	pages, err := b.VirtualPages(site, cfg(true, "test-22"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}

	p := pages[0]
	if p.OutputPath != "bookshelf/index.html" {
		t.Errorf("OutputPath = %q, want %q", p.OutputPath, "bookshelf/index.html")
	}
	if p.URL != "/bookshelf/" {
		t.Errorf("URL = %q, want %q", p.URL, "/bookshelf/")
	}
	if p.Template != "bookshelf.html" {
		t.Errorf("Template = %q, want %q", p.Template, "bookshelf.html")
	}
	if p.Locale != "en" {
		t.Errorf("Locale = %q, want %q", p.Locale, "en")
	}

	books, ok := p.Data["books"].([]bookshelf.BookEntry)
	if !ok {
		t.Fatalf("Data[\"books\"] type = %T, want []bookshelf.BookEntry", p.Data["books"])
	}
	if len(books) != 1 {
		t.Fatalf("expected 1 book entry, got %d", len(books))
	}
	e := books[0]
	if e.ASIN != "4774189774" {
		t.Errorf("ASIN = %q", e.ASIN)
	}
	if e.Title != "パーフェクトRuby" {
		t.Errorf("Title = %q", e.Title)
	}
	if !strings.Contains(e.ImageURL, "4774189774") {
		t.Errorf("ImageURL %q does not contain ASIN", e.ImageURL)
	}
	if !strings.Contains(e.LinkURL, "4774189774") {
		t.Errorf("LinkURL %q does not contain ASIN", e.LinkURL)
	}
	if !strings.Contains(e.LinkURL, "test-22") {
		t.Errorf("LinkURL %q does not contain affiliate tag", e.LinkURL)
	}
	if e.ArticleSlug != "perfect-ruby" {
		t.Errorf("ArticleSlug = %q", e.ArticleSlug)
	}
	if e.ArticleURL != "/posts/perfect-ruby/" {
		t.Errorf("ArticleURL = %q", e.ArticleURL)
	}
}

func TestBookshelf_MultiLocale_NonDefaultGetsPrefix(t *testing.T) {
	b := bookshelf.New()
	date := time.Date(2024, 5, 4, 0, 0, 0, 0, time.UTC)
	site := &model.Site{
		Config: model.Config{
			Site: model.SiteConfig{Language: "en"},
			I18n: model.I18nConfig{DefaultLocale: "en", Locales: []string{"en", "ja"}},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "Book EN",
					Slug:  "book-en",
					Date:  date,
					Extra: map[string]interface{}{
						"books": []interface{}{
							map[string]interface{}{"asin": "AAA", "title": "Book A"},
						},
					},
				}},
				Locale: "en",
				URL:    "/posts/book-en/",
			},
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "本 JA",
					Slug:  "book-ja",
					Date:  date,
					Extra: map[string]interface{}{
						"books": []interface{}{
							map[string]interface{}{"asin": "BBB", "title": "本 B"},
						},
					},
				}},
				Locale: "ja",
				URL:    "/ja/posts/book-ja/",
			},
		},
	}

	pages, err := b.VirtualPages(site, cfg(true, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages (one per locale), got %d", len(pages))
	}

	byLocale := map[string]*model.VirtualPage{}
	for _, p := range pages {
		byLocale[p.Locale] = p
	}

	enPage := byLocale["en"]
	if enPage == nil {
		t.Fatal("missing en page")
	}
	if enPage.OutputPath != "bookshelf/index.html" {
		t.Errorf("en OutputPath = %q", enPage.OutputPath)
	}
	if enPage.URL != "/bookshelf/" {
		t.Errorf("en URL = %q", enPage.URL)
	}

	jaPage := byLocale["ja"]
	if jaPage == nil {
		t.Fatal("missing ja page")
	}
	if jaPage.OutputPath != "ja/bookshelf/index.html" {
		t.Errorf("ja OutputPath = %q, want ja/bookshelf/index.html", jaPage.OutputPath)
	}
	if jaPage.URL != "/ja/bookshelf/" {
		t.Errorf("ja URL = %q, want /ja/bookshelf/", jaPage.URL)
	}
}

func TestBookshelf_SortedByDateDesc(t *testing.T) {
	b := bookshelf.New()
	site := &model.Site{
		Config: model.Config{
			Site: model.SiteConfig{Language: "en"},
			I18n: model.I18nConfig{DefaultLocale: "en"},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "Old",
					Slug:  "old",
					Date:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					Extra: map[string]interface{}{
						"books": []interface{}{
							map[string]interface{}{"asin": "OLD", "title": "Old Book"},
						},
					},
				}},
				Locale: "en",
			},
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "New",
					Slug:  "new",
					Date:  time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
					Extra: map[string]interface{}{
						"books": []interface{}{
							map[string]interface{}{"asin": "NEW", "title": "New Book"},
						},
					},
				}},
				Locale: "en",
			},
		},
	}

	pages, err := b.VirtualPages(site, cfg(true, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	books := pages[0].Data["books"].([]bookshelf.BookEntry)
	if books[0].ASIN != "NEW" {
		t.Errorf("first entry should be newest: got ASIN=%q", books[0].ASIN)
	}
	if books[1].ASIN != "OLD" {
		t.Errorf("second entry should be oldest: got ASIN=%q", books[1].ASIN)
	}
}

func TestBookshelf_InvalidBooksField(t *testing.T) {
	b := bookshelf.New()
	site := &model.Site{
		Config: model.Config{
			Site: model.SiteConfig{Language: "en"},
			I18n: model.I18nConfig{DefaultLocale: "en"},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "Bad",
					Extra: map[string]interface{}{
						"books": "not-a-list",
					},
				}},
				Locale: "en",
			},
		},
	}

	_, err := b.VirtualPages(site, cfg(true, ""))
	if err == nil {
		t.Error("expected error for invalid books field type, got nil")
	}
}
