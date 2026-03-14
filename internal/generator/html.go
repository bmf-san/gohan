package generator

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

	var errs []error
	for err := range errc {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if g.cfg.Build.AssetsDir != "" {
		if err := CopyDir(g.cfg.Build.AssetsDir, filepath.Join(g.outDir, "assets")); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("copy assets: %w", err)
			}
		}
	}

	if g.cfg.Build.StaticDir != "" {
		if err := CopyDir(g.cfg.Build.StaticDir, g.outDir); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("copy static: %w", err)
			}
		}
	}

	if g.cfg.OGP.Enabled {
		ogpGen := NewOGPGenerator(g.outDir, g.cfg.OGP)
		if err := ogpGen.Generate(site, changeSet); err != nil {
			return fmt.Errorf("ogp generation: %w", err)
		}
	}

	return nil
}

func (g *HTMLGenerator) buildJobs(site *model.Site) []writeJob {
	var jobs []writeJob
	perPage := g.cfg.Build.PerPage

	// Index pages (paginated) — one set per locale when i18n is active.
	if len(g.cfg.I18n.Locales) > 0 {
		for _, loc := range g.cfg.I18n.Locales {
			locale := loc
			locArticles := filterArticles(site.Articles, func(a *model.ProcessedArticle) bool {
				return a.Locale == locale
			})
			sortByDateDesc(locArticles)
			var basePath, baseURLPath string
			if locale == g.cfg.I18n.DefaultLocale {
				basePath = ""
				baseURLPath = ""
			} else {
				basePath = locale
				baseURLPath = "/" + locale
			}
			jobs = append(jobs, paginatedJobs(site, locArticles, g.outDir, "index.html", basePath, baseURLPath, perPage, locale, nil)...)
		}
	} else {
		allArticles := make([]*model.ProcessedArticle, len(site.Articles))
		copy(allArticles, site.Articles)
		sortByDateDesc(allArticles)
		jobs = append(jobs, paginatedJobs(site, allArticles, g.outDir, "index.html", "", "", perPage, "", nil)...)
	}

	// Article pages: use pre-computed output path and respect FrontMatter.Template.
	for _, a := range site.Articles {
		a := a
		articlePath := articleOutputPath(a, g.outDir, g.cfg)
		tmplName := "article.html"
		if a.FrontMatter.Template != "" {
			tmplName = a.FrontMatter.Template
		}
		d := siteFor(site, []*model.ProcessedArticle{a})
		d.CurrentLocale = a.Locale
		d.RelatedArticles = relatedArticles(site.Articles, a, 5)
		jobs = append(jobs, writeJob{
			path: articlePath,
			tmpl: tmplName,
			data: d,
		})
	}

	// Tag pages (paginated) — locale-aware when i18n is active
	if len(g.cfg.I18n.Locales) > 0 {
		for _, loc := range g.cfg.I18n.Locales {
			locale := loc
			for _, tag := range site.Tags {
				t := tag
				filtered := filterArticles(site.Articles, func(a *model.ProcessedArticle) bool {
					if a.Locale != locale {
						return false
					}
					for _, tt := range a.FrontMatter.Tags {
						if tt == t.Name {
							return true
						}
					}
					return false
				})
				if len(filtered) == 0 {
					continue
				}
				sortByDateDesc(filtered)
				var basePath, baseURLPath string
				if locale == g.cfg.I18n.DefaultLocale {
					basePath = filepath.Join("tags", tagNorm(t.Name))
					baseURLPath = "/tags/" + tagNorm(t.Name)
				} else {
					basePath = filepath.Join(locale, "tags", tagNorm(t.Name))
					baseURLPath = "/" + locale + "/tags/" + tagNorm(t.Name)
				}
				jobs = append(jobs, paginatedJobs(site, filtered, g.outDir, "tag.html", basePath, baseURLPath, perPage, locale, &t)...)
			}
		}
	} else {
		for _, tag := range site.Tags {
			t := tag
			filtered := filterArticles(site.Articles, func(a *model.ProcessedArticle) bool {
				for _, tt := range a.FrontMatter.Tags {
					if tt == t.Name {
						return true
					}
				}
				return false
			})
			if len(filtered) == 0 {
				continue
			}
			sortByDateDesc(filtered)
			basePath := filepath.Join("tags", tagNorm(t.Name))
			baseURLPath := "/tags/" + tagNorm(t.Name)
			jobs = append(jobs, paginatedJobs(site, filtered, g.outDir, "tag.html", basePath, baseURLPath, perPage, "", &t)...)
		}
	}

	// Category pages (paginated) — locale-aware when i18n is active
	if len(g.cfg.I18n.Locales) > 0 {
		for _, loc := range g.cfg.I18n.Locales {
			locale := loc
			for _, cat := range site.Categories {
				c := cat
				filtered := filterArticles(site.Articles, func(a *model.ProcessedArticle) bool {
					if a.Locale != locale {
						return false
					}
					for _, cc := range a.FrontMatter.Categories {
						if cc == c.Name {
							return true
						}
					}
					return false
				})
				if len(filtered) == 0 {
					continue
				}
				sortByDateDesc(filtered)
				var basePath, baseURLPath string
				if locale == g.cfg.I18n.DefaultLocale {
					basePath = filepath.Join("categories", tagNorm(c.Name))
					baseURLPath = "/categories/" + tagNorm(c.Name)
				} else {
					basePath = filepath.Join(locale, "categories", tagNorm(c.Name))
					baseURLPath = "/" + locale + "/categories/" + tagNorm(c.Name)
				}
				jobs = append(jobs, paginatedJobs(site, filtered, g.outDir, "category.html", basePath, baseURLPath, perPage, locale, &c)...)
			}
		}
	} else {
		for _, cat := range site.Categories {
			c := cat
			filtered := filterArticles(site.Articles, func(a *model.ProcessedArticle) bool {
				for _, cc := range a.FrontMatter.Categories {
					if cc == c.Name {
						return true
					}
				}
				return false
			})
			if len(filtered) == 0 {
				continue
			}
			sortByDateDesc(filtered)
			basePath := filepath.Join("categories", tagNorm(c.Name))
			baseURLPath := "/categories/" + tagNorm(c.Name)
			jobs = append(jobs, paginatedJobs(site, filtered, g.outDir, "category.html", basePath, baseURLPath, perPage, "", &c)...)
		}
	}

	// Archive pages — locale-aware when i18n is active.
	// Articles with a zero date are skipped to avoid generating archives/0001/01/.
	type ym struct {
		year  int
		month time.Month
	}

	if len(g.cfg.I18n.Locales) > 0 {
		for _, loc := range g.cfg.I18n.Locales {
			locale := loc
			locArticles := filterArticles(site.Articles, func(a *model.ProcessedArticle) bool {
				return a.Locale == locale && !a.FrontMatter.Date.IsZero()
			})

			archives := map[ym][]*model.ProcessedArticle{}
			yearArchives := map[int][]*model.ProcessedArticle{}
			for _, a := range locArticles {
				key := ym{a.FrontMatter.Date.Year(), a.FrontMatter.Date.Month()}
				archives[key] = append(archives[key], a)
				yearArchives[a.FrontMatter.Date.Year()] = append(yearArchives[a.FrontMatter.Date.Year()], a)
			}

			var archivePrefix string
			if locale != g.cfg.I18n.DefaultLocale {
				archivePrefix = locale
			}

			for key, articles := range archives {
				as := make([]*model.ProcessedArticle, len(articles))
				copy(as, articles)
				sortByDateDesc(as)
				k := key
				d := siteFor(site, as)
				d.CurrentLocale = locale
				jobs = append(jobs, writeJob{
					path: filepath.Join(g.outDir, archivePrefix, "archives",
						fmt.Sprintf("%04d", k.year),
						fmt.Sprintf("%02d", int(k.month)),
						"index.html"),
					tmpl: "archive.html",
					data: d,
				})
			}

			for year, articles := range yearArchives {
				as := make([]*model.ProcessedArticle, len(articles))
				copy(as, articles)
				sortByDateDesc(as)
				y := year
				d := siteFor(site, as)
				d.CurrentLocale = locale
				jobs = append(jobs, writeJob{
					path: filepath.Join(g.outDir, archivePrefix, "archives",
						fmt.Sprintf("%04d", y),
						"index.html"),
					tmpl: "archive.html",
					data: d,
				})
			}
		}
	} else {
		// Non-i18n: global archive pages (original behavior).
		archives := map[ym][]*model.ProcessedArticle{}
		for _, a := range site.Articles {
			if a.FrontMatter.Date.IsZero() {
				continue
			}
			key := ym{a.FrontMatter.Date.Year(), a.FrontMatter.Date.Month()}
			archives[key] = append(archives[key], a)
		}
		for key, articles := range archives {
			as := make([]*model.ProcessedArticle, len(articles))
			copy(as, articles)
			sortByDateDesc(as)
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

		yearArchives := map[int][]*model.ProcessedArticle{}
		for _, a := range site.Articles {
			if a.FrontMatter.Date.IsZero() {
				continue
			}
			yearArchives[a.FrontMatter.Date.Year()] = append(yearArchives[a.FrontMatter.Date.Year()], a)
		}
		for year, articles := range yearArchives {
			as := make([]*model.ProcessedArticle, len(articles))
			copy(as, articles)
			sortByDateDesc(as)
			y := year
			jobs = append(jobs, writeJob{
				path: filepath.Join(g.outDir, "archives",
					fmt.Sprintf("%04d", y),
					"index.html"),
				tmpl: "archive.html",
				data: siteFor(site, as),
			})
		}
	}

	return jobs
}

// paginatedJobs returns writeJobs for a paginated listing page.
// basePath is the filesystem path prefix relative to outDir (e.g. "tags/go").
// baseURLPath is the URL prefix for computing PrevURL/NextURL.
// taxonomy is the tag or category being listed; nil for plain index pages.
// When perPage <= 0, a single job with all articles and no Pagination is returned.
func paginatedJobs(
	site *model.Site,
	articles []*model.ProcessedArticle,
	outDir, tmpl, basePath, baseURLPath string,
	perPage int,
	currentLocale string,
	taxonomy *model.Taxonomy,
) []writeJob {
	// Build a locale-specific base so that listing pages see only the
	// tags/categories present in the locale's articles (all of them,
	// not just the current page slice).
	base := localeTaxonomyBase(site, articles)
	if perPage <= 0 {
		var path string
		if basePath == "" {
			path = filepath.Join(outDir, "index.html")
		} else {
			path = filepath.Join(outDir, basePath, "index.html")
		}
		d := siteFor(base, articles)
		d.CurrentLocale = currentLocale
		d.CurrentTaxonomy = taxonomy
		return []writeJob{{path: path, tmpl: tmpl, data: d}}
	}

	total := len(articles)
	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	var jobs []writeJob
	for page := 1; page <= totalPages; page++ {
		start := (page - 1) * perPage
		end := start + perPage
		if end > total {
			end = total
		}
		slice := articles[start:end]

		pg := &model.Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     perPage,
			TotalItems:  total,
			BaseURL:     baseURLPath,
		}
		if page > 1 {
			if page == 2 {
				if baseURLPath != "" {
					pg.PrevURL = baseURLPath + "/"
				} else {
					pg.PrevURL = "/"
				}
			} else {
				pg.PrevURL = fmt.Sprintf("%s/page/%d/", baseURLPath, page-1)
			}
		}
		if page < totalPages {
			pg.NextURL = fmt.Sprintf("%s/page/%d/", baseURLPath, page+1)
		}

		var path string
		if page == 1 {
			if basePath == "" {
				path = filepath.Join(outDir, "index.html")
			} else {
				path = filepath.Join(outDir, basePath, "index.html")
			}
		} else {
			if basePath == "" {
				path = filepath.Join(outDir, "page", fmt.Sprintf("%d", page), "index.html")
			} else {
				path = filepath.Join(outDir, basePath, "page", fmt.Sprintf("%d", page), "index.html")
			}
		}

		d := siteWithPagination(base, slice, pg)
		d.CurrentLocale = currentLocale
		d.CurrentTaxonomy = taxonomy
		jobs = append(jobs, writeJob{
			path: path,
			tmpl: tmpl,
			data: d,
		})
	}
	return jobs
}

// filterArticles returns articles matching pred.
func filterArticles(articles []*model.ProcessedArticle, pred func(*model.ProcessedArticle) bool) []*model.ProcessedArticle {
	var out []*model.ProcessedArticle
	for _, a := range articles {
		if pred(a) {
			out = append(out, a)
		}
	}
	return out
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
	if err := writeFileAtomic(path, pageBytes, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// CopyDir recursively copies all files from srcDir into dstDir.
func CopyDir(srcDir, dstDir string) error {
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
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return writeFileAtomic(dst, data, info.Mode().Perm())
}

// slugify converts s to a lowercase hyphen-separated URL slug.
// Only ASCII letters, digits, and hyphens are kept; spaces and underscores
// become hyphens. Returns "untitled" when the input produces an empty result
// (e.g. all non-ASCII characters with no slug in front-matter).
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
	if len(out) == 0 {
		return "untitled"
	}
	return string(out)
}

// tagNorm normalises a tag or category name for use in URL paths and
// filesystem directories. ASCII letters are lowercased and spaces become
// hyphens; non-ASCII characters (e.g. Japanese) are left intact.
// This keeps tagNorm consistent with the tagURL / categoryURL template helpers.
func tagNorm(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		case r == ' ':
			b.WriteByte('-')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// siteFor creates a site copy with a custom article list.
func siteFor(base *model.Site, articles []*model.ProcessedArticle) *model.Site {
	return &model.Site{
		Config:       base.Config,
		Articles:     articles,
		Tags:         base.Tags,
		Categories:   base.Categories,
		ArchiveYears: base.ArchiveYears,
	}
}

// localeTaxonomyBase returns a copy of base whose Tags and Categories are
// derived from the unique values found in the given articles' frontmatter.
// This ensures that locale-specific listing pages (e.g. /ja/ index) only
// expose taxonomy entries that exist in that locale's articles.
func localeTaxonomyBase(base *model.Site, articles []*model.ProcessedArticle) *model.Site {
	tagSeen := make(map[string]bool)
	catSeen := make(map[string]bool)
	var tags []model.Taxonomy
	var cats []model.Taxonomy
	for _, a := range articles {
		for _, t := range a.FrontMatter.Tags {
			if !tagSeen[t] {
				tagSeen[t] = true
				tags = append(tags, model.Taxonomy{Name: t})
			}
		}
		for _, c := range a.FrontMatter.Categories {
			if !catSeen[c] {
				catSeen[c] = true
				cats = append(cats, model.Taxonomy{Name: c})
			}
		}
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].Name < tags[j].Name })
	sort.Slice(cats, func(i, j int) bool { return cats[i].Name < cats[j].Name })
	if len(tags) == 0 {
		tags = base.Tags
	}
	if len(cats) == 0 {
		cats = base.Categories
	}
	return &model.Site{
		Config:       base.Config,
		Articles:     base.Articles,
		Tags:         tags,
		Categories:   cats,
		ArchiveYears: archiveYears(articles),
	}
}

// archiveYears returns the unique years present in articles, sorted newest-first.
func archiveYears(articles []*model.ProcessedArticle) []int {
	seen := make(map[int]bool)
	for _, a := range articles {
		if !a.FrontMatter.Date.IsZero() {
			seen[a.FrontMatter.Date.Year()] = true
		}
	}
	years := make([]int, 0, len(seen))
	for y := range seen {
		years = append(years, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	return years
}

// siteWithPagination creates a site copy with a custom article list and pagination metadata.
func siteWithPagination(base *model.Site, articles []*model.ProcessedArticle, pg *model.Pagination) *model.Site {
	s := siteFor(base, articles)
	s.Pagination = pg
	return s
}

// sortByDateDesc sorts a slice of processed articles newest-first in place.
func sortByDateDesc(articles []*model.ProcessedArticle) {
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].FrontMatter.Date.After(articles[j].FrontMatter.Date)
	})
}

// relatedArticles returns up to n articles that share at least one category
// with a (same locale), excluding a itself, sorted newest-first.
func relatedArticles(all []*model.ProcessedArticle, a *model.ProcessedArticle, n int) []*model.ProcessedArticle {
	catSet := make(map[string]bool, len(a.FrontMatter.Categories))
	for _, c := range a.FrontMatter.Categories {
		catSet[c] = true
	}
	var related []*model.ProcessedArticle
	for _, candidate := range all {
		if candidate == a || candidate.Locale != a.Locale {
			continue
		}
		for _, c := range candidate.FrontMatter.Categories {
			if catSet[c] {
				related = append(related, candidate)
				break
			}
		}
	}
	sortByDateDesc(related)
	if len(related) > n {
		related = related[:n]
	}
	return related
}

// articleOutputPath returns the absolute filesystem path for an article page.
// When a.OutputPath is a valid relative path under cfg.Build.OutputDir, it is
// translated to an absolute path under outDir.  Otherwise (e.g. in tests that
// create ProcessedArticles without OutputPath), it falls back to the
// slug-based locale-aware path used in previous versions.
func articleOutputPath(a *model.ProcessedArticle, outDir string, cfg model.Config) string {
	if a.OutputPath != "" && cfg.Build.OutputDir != "" {
		rel, err := filepath.Rel(cfg.Build.OutputDir, a.OutputPath)
		// Accept only valid descendants: no "." (same dir) and no ".." escapes.
		if err == nil && rel != "." && !strings.HasPrefix(filepath.ToSlash(rel), "..") {
			return filepath.Join(outDir, rel)
		}
	}
	// Fallback: construct from slug and locale.
	slug := a.FrontMatter.Slug
	if slug == "" {
		slug = slugify(a.FrontMatter.Title)
	}
	if a.Locale != "" && a.Locale != cfg.I18n.DefaultLocale {
		return filepath.Join(outDir, a.Locale, "posts", slug, "index.html")
	}
	return filepath.Join(outDir, "posts", slug, "index.html")
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
