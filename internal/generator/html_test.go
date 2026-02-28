package generator

import (
	htmltemplate "html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

type mockEngine struct{ calls []string }

func (m *mockEngine) Load(_ string, _ htmltemplate.FuncMap) error { return nil }
func (m *mockEngine) Render(w io.Writer, name string, _ *model.Site) error {
	m.calls = append(m.calls, name)
	_, err := io.WriteString(w, "<html>"+name+"</html>")
	return err
}

func makeSite() *model.Site {
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	return &model.Site{
		Config: model.Config{
			Site:  model.SiteConfig{Title: "Test Site", BaseURL: "https://example.com"},
			Build: model.BuildConfig{Parallelism: 2},
		},
		Articles: []*model.ProcessedArticle{{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Title: "Hello World", Slug: "hello-world",
				Tags: []string{"go"}, Categories: []string{"tech"}, Date: now,
			}},
			Summary: "A summary.",
		}},
		Tags:       []model.Taxonomy{{Name: "go"}},
		Categories: []model.Taxonomy{{Name: "tech"}},
	}
}

func TestGenerate_WritesExpectedFiles(t *testing.T) {
	outDir := t.TempDir()
	g := NewHTMLGenerator(outDir, &mockEngine{}, model.Config{Build: model.BuildConfig{Parallelism: 2}})
	if err := g.Generate(makeSite(), nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, rel := range []string{"index.html", "posts/hello-world/index.html",
		"tags/go/index.html", "categories/tech/index.html", "archives/2024/03/index.html"} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestGenerate_SlugifiesTitle(t *testing.T) {
	outDir := t.TempDir()
	g := NewHTMLGenerator(outDir, &mockEngine{}, model.Config{Build: model.BuildConfig{Parallelism: 1}})
	site := makeSite()
	site.Articles[0].FrontMatter.Slug = ""
	site.Articles[0].FrontMatter.Title = "My Test Post"
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "posts", "my-test-post", "index.html")); err != nil {
		t.Errorf("expected slugified path: %v", err)
	}
}

func TestGenerate_CopiesAssets(t *testing.T) {
	srcDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "style.css"), []byte("body{}"), 0o644)
	outDir := t.TempDir()
	cfg := model.Config{Build: model.BuildConfig{Parallelism: 1, AssetsDir: srcDir}}
	if err := NewHTMLGenerator(outDir, &mockEngine{}, cfg).Generate(makeSite(), nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "assets", "style.css")); err != nil {
		t.Errorf("expected copied asset: %v", err)
	}
}

func TestGenerateSitemap(t *testing.T) {
	outDir := t.TempDir()
	if err := NewHTMLGenerator(outDir, &mockEngine{}, model.Config{}).GenerateSitemap(makeSite()); err != nil {
		t.Fatalf("GenerateSitemap: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(outDir, "sitemap.xml"))
	if !strings.Contains(string(data), "hello-world") || !strings.Contains(string(data), "urlset") {
		t.Errorf("sitemap wrong:\n%s", data)
	}
}

func TestGenerateFeed(t *testing.T) {
	outDir := t.TempDir()
	if err := NewHTMLGenerator(outDir, &mockEngine{}, model.Config{}).GenerateFeed(makeSite()); err != nil {
		t.Fatalf("GenerateFeed: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(outDir, "atom.xml"))
	if !strings.Contains(string(data), "Hello World") || !strings.Contains(string(data), "Test Site") {
		t.Errorf("feed wrong:\n%s", data)
	}
}

func TestCopyAssets_PreservesStructure(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "css")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "main.css"), []byte(".a{}"), 0o644)
	dst := t.TempDir()
	if err := CopyAssets(src, dst); err != nil {
		t.Fatalf("CopyAssets: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(dst, "css", "main.css")); string(got) != ".a{}" {
		t.Errorf("unexpected content: %s", got)
	}
}

func TestSlugify(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"Hello World", "hello-world"}, {"My Post", "my-post"},
		{"already-fine", "already-fine"}, {"CamelCase", "camelcase"},
	} {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFilteredSite(t *testing.T) {
	site := makeSite()
	got := filteredSite(site, func(a *model.ProcessedArticle) bool { return true })
	if len(got.Articles) != 1 {
		t.Errorf("expected 1, got %d", len(got.Articles))
	}
	empty := filteredSite(site, func(a *model.ProcessedArticle) bool { return false })
	if len(empty.Articles) != 0 {
		t.Errorf("expected 0, got %d", len(empty.Articles))
	}
}
