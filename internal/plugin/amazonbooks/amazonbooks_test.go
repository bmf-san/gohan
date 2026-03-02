package amazonbooks_test

import (
	"testing"

	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/plugin/amazonbooks"
)

func TestAmazonBooks_Name(t *testing.T) {
	p := amazonbooks.New()
	if got := p.Name(); got != "amazon_books" {
		t.Errorf("Name() = %q, want %q", got, "amazon_books")
	}
}

func TestAmazonBooks_Enabled(t *testing.T) {
	p := amazonbooks.New()

	cases := []struct {
		cfg  map[string]interface{}
		want bool
	}{
		{map[string]interface{}{"enabled": true, "tag": "test-22"}, true},
		{map[string]interface{}{"enabled": false}, false},
		{map[string]interface{}{}, false},
		{nil, false},
	}
	for _, tc := range cases {
		if got := p.Enabled(tc.cfg); got != tc.want {
			t.Errorf("Enabled(%v) = %v, want %v", tc.cfg, got, tc.want)
		}
	}
}

func TestAmazonBooks_TemplateData_NoBooks(t *testing.T) {
	p := amazonbooks.New()
	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{Title: "Test"},
		},
	}
	cfg := map[string]interface{}{"enabled": true, "tag": "test-22"}

	data, err := p.TemplateData(article, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	books, ok := data["books"]
	if !ok {
		t.Fatal("expected 'books' key in result")
	}
	cards, ok := books.([]amazonbooks.BookCard)
	if !ok {
		t.Fatalf("expected []BookCard, got %T", books)
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
}

func TestAmazonBooks_TemplateData_WithBooks(t *testing.T) {
	p := amazonbooks.New()
	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{
				Title: "Test",
				Extra: map[string]interface{}{
					"books": []interface{}{
						map[string]interface{}{"asin": "4781920004", "title": "組織を変える5つの対話"},
						map[string]interface{}{"asin": "4873119464"},
					},
				},
			},
		},
	}
	cfg := map[string]interface{}{"enabled": true, "tag": "bmf035-22"}

	data, err := p.TemplateData(article, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cards := data["books"].([]amazonbooks.BookCard)
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}

	c0 := cards[0]
	if c0.ASIN != "4781920004" {
		t.Errorf("ASIN = %q, want %q", c0.ASIN, "4781920004")
	}
	if c0.Title != "組織を変える5つの対話" {
		t.Errorf("Title = %q", c0.Title)
	}
	wantImage := "https://images-na.ssl-images-amazon.com/images/P/4781920004.01._SL250_.jpg"
	if c0.ImageURL != wantImage {
		t.Errorf("ImageURL = %q, want %q", c0.ImageURL, wantImage)
	}
	wantLink := "https://www.amazon.co.jp/dp/4781920004?tag=bmf035-22"
	if c0.LinkURL != wantLink {
		t.Errorf("LinkURL = %q, want %q", c0.LinkURL, wantLink)
	}

	// Second card: no title → falls back to ASIN
	c1 := cards[1]
	if c1.ASIN != "4873119464" {
		t.Errorf("ASIN = %q", c1.ASIN)
	}
	if c1.Title != "4873119464" {
		t.Errorf("Title fallback = %q, want ASIN as fallback", c1.Title)
	}
}

func TestAmazonBooks_TemplateData_InvalidBooks(t *testing.T) {
	p := amazonbooks.New()
	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{
				Extra: map[string]interface{}{
					"books": "not-a-list",
				},
			},
		},
	}
	cfg := map[string]interface{}{"enabled": true, "tag": "test-22"}

	_, err := p.TemplateData(article, cfg)
	if err == nil {
		t.Error("expected error for invalid books value, got nil")
	}
}

func TestAmazonBooks_TemplateData_SkipsEmptyASIN(t *testing.T) {
	p := amazonbooks.New()
	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{
				Extra: map[string]interface{}{
					"books": []interface{}{
						map[string]interface{}{"asin": ""},
						map[string]interface{}{"asin": "4873119464"},
					},
				},
			},
		},
	}
	cfg := map[string]interface{}{"enabled": true, "tag": "test-22"}

	data, err := p.TemplateData(article, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cards := data["books"].([]amazonbooks.BookCard)
	if len(cards) != 1 {
		t.Errorf("expected 1 card (empty ASIN skipped), got %d", len(cards))
	}
}
