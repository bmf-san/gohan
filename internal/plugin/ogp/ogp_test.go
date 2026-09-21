package ogp

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

func ogpSite(articles ...*model.ProcessedArticle) *model.Site {
	return &model.Site{
		Config:   model.Config{},
		Articles: articles,
	}
}

func TestOGP_Enabled(t *testing.T) {
	o := New()
	if o.Enabled(nil) {
		t.Error("Enabled(nil) = true, want false")
	}
	if o.Enabled(map[string]interface{}{"enabled": false}) {
		t.Error("Enabled(enabled:false) = true, want false")
	}
	if !o.Enabled(map[string]interface{}{"enabled": true}) {
		t.Error("Enabled(enabled:true) = false, want true")
	}
}

func TestOGP_ProducesImage(t *testing.T) {
	outDir := t.TempDir()
	o := New()
	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{
				Title: "Test OGP",
				Slug:  "test-ogp",
				Date:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	cfg := map[string]interface{}{"enabled": true, "width": 120, "height": 63}
	if err := o.GenerateAssets(ogpSite(article), outDir, nil, cfg); err != nil {
		t.Fatalf("GenerateAssets: %v", err)
	}
	path := filepath.Join(outDir, "ogp", "test-ogp.png")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("expected OGP image at %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 120 || b.Dy() != 63 {
		t.Errorf("image size = %dx%d, want 120x63", b.Dx(), b.Dy())
	}
}

func TestOGP_SkipsUnchanged(t *testing.T) {
	outDir := t.TempDir()
	o := New()
	cfg := map[string]interface{}{"enabled": true, "width": 120, "height": 63}

	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{Title: "A", Slug: "a"},
			FilePath:    "/content/posts/a.md",
		},
	}
	unchanged := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{Title: "B", Slug: "b"},
			FilePath:    "/content/posts/b.md",
		},
	}
	site := ogpSite(article, unchanged)

	if err := o.GenerateAssets(site, outDir, nil, cfg); err != nil {
		t.Fatalf("initial GenerateAssets: %v", err)
	}
	// Remove a.png so it must be regenerated; keep b.png.
	if err := os.Remove(filepath.Join(outDir, "ogp", "a.png")); err != nil {
		t.Fatal(err)
	}
	// changeSet only contains "a" — "b" should be skipped (exists + not changed).
	changeSet := &model.ChangeSet{ModifiedFiles: []string{"/content/posts/a.md"}}
	if err := o.GenerateAssets(site, outDir, changeSet, cfg); err != nil {
		t.Fatalf("GenerateAssets with changeSet: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ogp", "a.png")); err != nil {
		t.Errorf("expected a.png to be regenerated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ogp", "b.png")); err != nil {
		t.Errorf("b.png should still exist: %v", err)
	}
}

func TestOGP_NilChangeSet_GeneratesAll(t *testing.T) {
	outDir := t.TempDir()
	o := New()
	cfg := map[string]interface{}{"enabled": true, "width": 120, "height": 63}
	articles := []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "First", Slug: "first"}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Second", Slug: "second"}}},
	}
	if err := o.GenerateAssets(ogpSite(articles...), outDir, nil, cfg); err != nil {
		t.Fatalf("GenerateAssets: %v", err)
	}
	for _, slug := range []string{"first", "second"} {
		if _, err := os.Stat(filepath.Join(outDir, "ogp", slug+".png")); err != nil {
			t.Errorf("expected ogp/%s.png: %v", slug, err)
		}
	}
}

func TestOGP_DefaultDimensions(t *testing.T) {
	outDir := t.TempDir()
	o := New()
	cfg := map[string]interface{}{"enabled": true} // no width/height → defaults
	article := &model.ProcessedArticle{
		Article: model.Article{FrontMatter: model.FrontMatter{Title: "Def", Slug: "def"}},
	}
	if err := o.GenerateAssets(ogpSite(article), outDir, nil, cfg); err != nil {
		t.Fatalf("GenerateAssets: %v", err)
	}
	f, err := os.Open(filepath.Join(outDir, "ogp", "def.png"))
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer func() { _ = f.Close() }()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if b := img.Bounds(); b.Dx() != defaultWidth || b.Dy() != defaultHeight {
		t.Errorf("expected %dx%d, got %dx%d", defaultWidth, defaultHeight, b.Dx(), b.Dy())
	}
}
