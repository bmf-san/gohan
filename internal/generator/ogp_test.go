package generator

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

func ogpSite(articles ...*model.ProcessedArticle) *model.Site {
	s := makeSite()
	if len(articles) > 0 {
		s.Articles = articles
	}
	return s
}

func TestOGPGenerator_Disabled(t *testing.T) {
	outDir := t.TempDir()
	cfg := model.OGPConfig{Enabled: false}
	gen := NewOGPGenerator(outDir, "", cfg)
	site := ogpSite()
	if err := gen.Generate(site, nil); err != nil {
		t.Fatalf("unexpected error when disabled: %v", err)
	}
	// No files should be created
	entries, _ := os.ReadDir(outDir)
	if len(entries) != 0 {
		t.Errorf("expected no output files when disabled, got %d", len(entries))
	}
}

func TestOGPGenerator_NoFontFile_ProducesImage(t *testing.T) {
	// Even without a font, a background image should be generated (text skipped gracefully)
	outDir := t.TempDir()
	cfg := model.OGPConfig{
		Enabled: true,
		Width:   120,
		Height:  63,
	}
	gen := NewOGPGenerator(outDir, "", cfg)
	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{
				Title: "Test OGP",
				Slug:  "test-ogp",
				Date:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	site := ogpSite(article)
	if err := gen.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	path := filepath.Join(outDir, "ogp", "test-ogp.png")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected OGP image at %s: %v", path, err)
	}
	// Verify it's a valid PNG
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 120 || b.Dy() != 63 {
		t.Errorf("image size = %dx%d, want 120x63", b.Dx(), b.Dy())
	}
}

func TestOGPGenerator_SkipsUnchanged(t *testing.T) {
	outDir := t.TempDir()
	cfg := model.OGPConfig{
		Enabled: true,
		Width:   120,
		Height:  63,
	}
	gen := NewOGPGenerator(outDir, "", cfg)

	article := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{Title: "A", Slug: "a"},
			FilePath:    "/content/posts/a.md",
		},
	}
	unchangedArticle := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{Title: "B", Slug: "b"},
			FilePath:    "/content/posts/b.md",
		},
	}
	site := ogpSite(article, unchangedArticle)

	// First, generate both so "b.png" already exists on disk
	if err := gen.Generate(site, nil); err != nil {
		t.Fatalf("initial Generate: %v", err)
	}

	// Remove "a.png" to simulate it needing regeneration; keep "b.png"
	if err := os.Remove(filepath.Join(outDir, "ogp", "a.png")); err != nil {
		t.Fatal(err)
	}

	// changeSet only contains "a" — "b" should be skipped (file exists + not in changeSet)
	changeSet := &model.ChangeSet{ModifiedFiles: []string{"/content/posts/a.md"}}
	if err := gen.Generate(site, changeSet); err != nil {
		t.Fatalf("Generate with changeSet: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "ogp", "a.png")); err != nil {
		t.Errorf("expected a.png to be regenerated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ogp", "b.png")); err != nil {
		t.Errorf("b.png should still exist (not deleted): %v", err)
	}
}

func TestOGPGenerator_NilChangeSet_GeneratesAll(t *testing.T) {
	outDir := t.TempDir()
	cfg := model.OGPConfig{
		Enabled: true,
		Width:   120,
		Height:  63,
	}
	gen := NewOGPGenerator(outDir, "", cfg)

	articles := []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "First", Slug: "first"}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Second", Slug: "second"}}},
	}
	site := ogpSite(articles...)
	if err := gen.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, slug := range []string{"first", "second"} {
		if _, err := os.Stat(filepath.Join(outDir, "ogp", slug+".png")); err != nil {
			t.Errorf("expected ogp/%s.png: %v", slug, err)
		}
	}
}

func TestOGPGenerator_DefaultDimensions(t *testing.T) {
	outDir := t.TempDir()
	// Width/Height = 0 should use defaults (1200x630)
	cfg := model.OGPConfig{
		Enabled: true,
	}
	gen := NewOGPGenerator(outDir, "", cfg)
	article := &model.ProcessedArticle{
		Article: model.Article{FrontMatter: model.FrontMatter{Title: "Def", Slug: "def"}},
	}
	if err := gen.Generate(ogpSite(article), nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	f, _ := os.Open(filepath.Join(outDir, "ogp", "def.png"))
	defer func() { _ = f.Close() }()
	img, _ := png.Decode(f)
	b := img.Bounds()
	if b.Dx() != ogpDefaultWidth || b.Dy() != ogpDefaultHeight {
		t.Errorf("expected %dx%d, got %dx%d", ogpDefaultWidth, ogpDefaultHeight, b.Dx(), b.Dy())
	}
}
