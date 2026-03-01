package generator

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bmf-san/gohan/internal/mermaid"
	"github.com/bmf-san/gohan/internal/model"
	gohantemplate "github.com/bmf-san/gohan/internal/template"
)

// HTMLGenerator satisfies OutputGenerator by writing HTML pages and assets.
type HTMLGenerator struct {
	outDir string
	engine gohantemplate.TemplateEngine
	cfg    model.Config
}

// NewHTMLGenerator returns an HTMLGenerator that writes to outDir.
func NewHTMLGenerator(outDir string, engine gohantemplate.TemplateEngine, cfg model.Config) *HTMLGenerator {
	return &HTMLGenerator{outDir: outDir, engine: engine, cfg: cfg}
}

// writeJob describes a single page to render.
type writeJob struct {
	path string
	tmpl string
	data *model.Site
}

// Generate writes all HTML pages and copies static assets.
// When changeSet is nil every page is written; otherwise only affected pages.
func (g *HTMLGenerator) Generate(site *model.Site, changeSet *model.ChangeSet) error {
	parallelism := g.cfg.Build.Parallelism
	if parallelism <= 0 {
		parallelism = 1
	}

	jobs := g.buildJobs(site)
	sem := make(chan struct{}, parallelism)
	errc := make(chan error, len(jobs))
	var wg sync.WaitGroup

	for _, job := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(j writeJob) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := g.writePage(j.path, j.tmpl, j.data); err != nil {
				errc <- err
			}
		}(job)
	}
	wg.Wait()
	close(errc)

	for err := range errc {
		if err != nil {
			return err
		}
	}

	if g.cfg.Build.AssetsDir != "" {
		if err := CopyAssets(g.cfg.Build.AssetsDir, filepath.Join(g.outDir, "assets")); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("copy assets: %w", err)
			}
		}
	}
	return nil
}

func (g *HTMLGenerator) buildJobs(site *model.Site) []writeJob {
	var jobs []writeJob

	// Index page
	jobs = append(jobs, writeJob{
		path: filepath.Join(g.outDir, "index.html"),
		tmpl: "index.html",
		data: site,
	})

	// Article pages: public/posts/<slug>/index.html
	for _, a := range site.Articles {
		a := a
		slug := a.FrontMatter.Slug
		if slug == "" {
			slug = slugify(a.FrontMatter.Title)
		}
		jobs = append(jobs, writeJob{
			path: filepath.Join(g.outDir, "posts", slug, "index.html"),
			tmpl: "article.html",
			data: siteFor(site, []*model.ProcessedArticle{a}),
		})
	}

	// Tag pages: public/tags/<name>/index.html
	for _, tag := range site.Tags {
		t := tag
		jobs = append(jobs, writeJob{
			path: filepath.Join(g.outDir, "tags", t.Name, "index.html"),
			tmpl: "tag.html",
			data: filteredSite(site, func(a *model.ProcessedArticle) bool {
				for _, tt := range a.FrontMatter.Tags {
					if tt == t.Name {
						return true
					}
				}
				return false
			}),
		})
	}

	// Category pages: public/categories/<name>/index.html
	for _, cat := range site.Categories {
		c := cat
		jobs = append(jobs, writeJob{
			path: filepath.Join(g.outDir, "categories", c.Name, "index.html"),
			tmpl: "category.html",
			data: filteredSite(site, func(a *model.ProcessedArticle) bool {
				for _, cc := range a.FrontMatter.Categories {
					if cc == c.Name {
						return true
					}
				}
				return false
			}),
		})
	}

	// Archive pages: public/archives/<year>/<month>/index.html
	type ym struct {
		year  int
		month time.Month
	}
	archives := map[ym][]*model.ProcessedArticle{}
	for _, a := range site.Articles {
		key := ym{a.FrontMatter.Date.Year(), a.FrontMatter.Date.Month()}
		archives[key] = append(archives[key], a)
	}
	for key, articles := range archives {
		as := articles
		k := key
		jobs = append(jobs, writeJob{
			path: filepath.Join(g.outDir, "archives",
				fmt.Sprintf("%04d", k.year),
				fmt.Sprintf("%02d", int(k.month)),
				"index.html"),
			tmpl: "archive.html",
			data: siteFor(site, as),
		})
	}

	return jobs
}

func (g *HTMLGenerator) writePage(path, tmplName string, data *model.Site) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	var buf bytes.Buffer
	if err := g.engine.Render(&buf, tmplName, data); err != nil {
		return fmt.Errorf("render %s: %w", tmplName, err)
	}
	pageBytes := buf.Bytes()
	if bytes.Contains(pageBytes, []byte(mermaid.MermaidMarker)) {
		pageBytes = mermaid.InjectScript(pageBytes)
	}
	if err := os.WriteFile(path, pageBytes, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// GenerateSitemap writes a sitemap.xml listing all article URLs.
func (g *HTMLGenerator) GenerateSitemap(site *model.Site) error {
	baseURL := site.Config.Site.BaseURL
	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	for _, a := range site.Articles {
		slug := a.FrontMatter.Slug
		if slug == "" {
			slug = slugify(a.FrontMatter.Title)
		}
		buf.WriteString("  <url><loc>" + baseURL + "/posts/" + slug + "/</loc></url>\n")
	}
	buf.WriteString("</urlset>\n")
	if err := os.MkdirAll(g.outDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(g.outDir, "sitemap.xml"), buf.Bytes(), 0o644)
}

// GenerateFeed writes an atom.xml (Atom feed) listing all articles.
func (g *HTMLGenerator) GenerateFeed(site *model.Site) error {
	baseURL := site.Config.Site.BaseURL
	title := site.Config.Site.Title
	now := time.Now().UTC().Format(time.RFC3339)

	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString("<feed xmlns=\"http://www.w3.org/2005/Atom\">\n")
	buf.WriteString("  <title>" + title + "</title>\n")
	buf.WriteString("  <link href=\"" + baseURL + "\"/>\n")
	buf.WriteString("  <updated>" + now + "</updated>\n")
	for _, a := range site.Articles {
		slug := a.FrontMatter.Slug
		if slug == "" {
			slug = slugify(a.FrontMatter.Title)
		}
		updated := a.FrontMatter.Date.UTC().Format(time.RFC3339)
		buf.WriteString("  <entry>\n")
		buf.WriteString("    <title>" + a.FrontMatter.Title + "</title>\n")
		buf.WriteString("    <link href=\"" + baseURL + "/posts/" + slug + "/\"/>\n")
		buf.WriteString("    <updated>" + updated + "</updated>\n")
		buf.WriteString("    <summary>" + a.Summary + "</summary>\n")
		buf.WriteString("  </entry>\n")
	}
	buf.WriteString("</feed>\n")

	if err := os.MkdirAll(g.outDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(g.outDir, "atom.xml"), buf.Bytes(), 0o644)
}

// CopyAssets recursively copies all files from srcDir into dstDir.
func CopyAssets(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		dst := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		return copyFile(path, dst)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// slugify converts s to a lowercase hyphen-separated URL slug.
func slugify(s string) string {
	var out []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			out = append(out, c+32)
		case c == ' ' || c == '_':
			out = append(out, '-')
		case (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-':
			out = append(out, c)
		}
	}
	return string(out)
}

// siteFor creates a site copy with a custom article list.
func siteFor(base *model.Site, articles []*model.ProcessedArticle) *model.Site {
	return &model.Site{
		Config:     base.Config,
		Articles:   articles,
		Tags:       base.Tags,
		Categories: base.Categories,
	}
}

// filteredSite creates a site copy with articles matching pred.
func filteredSite(base *model.Site, pred func(*model.ProcessedArticle) bool) *model.Site {
	var filtered []*model.ProcessedArticle
	for _, a := range base.Articles {
		if pred(a) {
			filtered = append(filtered, a)
		}
	}
	return siteFor(base, filtered)
}
