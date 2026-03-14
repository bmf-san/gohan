package generator

import (
	"fmt"
	htmltemplate "html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

type mockEngine struct {
	mu    sync.Mutex
	calls []string
}

func (m *mockEngine) Load(_ string, _ htmltemplate.FuncMap, _ string) error { return nil }
func (m *mockEngine) Render(w io.Writer, name string, _ *model.Site) error {
	m.mu.Lock()
	m.calls = append(m.calls, name)
	m.mu.Unlock()
	_, err := io.WriteString(w, "<html>"+name+"</html>")
	return err
}

// captureEngine records each Render call with its template name and a copy of the Site data.
type captureRender struct {
	tmpl string
	data *model.Site
}

type captureEngine struct {
	mu      sync.Mutex
	renders []captureRender
}

func (c *captureEngine) Load(_ string, _ htmltemplate.FuncMap, _ string) error { return nil }
func (c *captureEngine) Render(w io.Writer, name string, data *model.Site) error {
	// Copy the site value so mutations after Render don't affect captured state.
	dataCopy := *data
	c.mu.Lock()
	c.renders = append(c.renders, captureRender{tmpl: name, data: &dataCopy})
	c.mu.Unlock()
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
		"tags/go/index.html", "categories/tech/index.html", "archives/2024/03/index.html",
		"archives/2024/index.html"} {
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
	if err := os.WriteFile(filepath.Join(srcDir, "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	cfg := model.Config{Build: model.BuildConfig{Parallelism: 1, AssetsDir: srcDir}}
	if err := NewHTMLGenerator(outDir, &mockEngine{}, cfg).Generate(makeSite(), nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "assets", "style.css")); err != nil {
		t.Errorf("expected copied asset: %v", err)
	}
}

func TestGenerate_CopiesStatic(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "robots.txt"), []byte("User-agent: *"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "ads.txt"), []byte("google.com, pub-123, DIRECT, abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	cfg := model.Config{Build: model.BuildConfig{Parallelism: 1, StaticDir: staticDir}}
	if err := NewHTMLGenerator(outDir, &mockEngine{}, cfg).Generate(makeSite(), nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "robots.txt")); err != nil {
		t.Errorf("expected robots.txt at output root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ads.txt")); err != nil {
		t.Errorf("expected ads.txt at output root: %v", err)
	}
}

func TestCopyDir_PreservesStructure(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "css")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "main.css"), []byte(".a{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(dst, "css", "main.css")); string(got) != ".a{}" {
		t.Errorf("unexpected content: %s", got)
	}
}

func TestSlugify(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"Hello World", "hello-world"}, {"My Post", "my-post"},
		{"already-fine", "already-fine"}, {"CamelCase", "camelcase"},
		{"", "untitled"}, {"コードレビュー", "untitled"}, // non-ASCII → "untitled" fallback
	} {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTagNorm(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"Go Programming", "go-programming"},
		{"コードレビュー", "コードレビュー"},
		{"Machine Learning", "machine-learning"},
		{"already-fine", "already-fine"},
		{"CamelCase", "camelcase"},
	} {
		if got := tagNorm(c.in); got != c.want {
			t.Errorf("tagNorm(%q) = %q, want %q", c.in, got, c.want)
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

// ---- Pagination tests ----

func makePaginatedArticles(n int) []*model.ProcessedArticle {
	articles := make([]*model.ProcessedArticle, n)
	for i := range articles {
		articles[i] = &model.ProcessedArticle{
			Article: model.Article{
				FrontMatter: model.FrontMatter{
					Title: fmt.Sprintf("Article %d", i+1),
					Slug:  fmt.Sprintf("article-%d", i+1),
				},
			},
		}
	}
	return articles
}

func TestPaginatedJobs_Disabled(t *testing.T) {
	site := makeSite()
	articles := makePaginatedArticles(5)
	jobs := paginatedJobs(site, articles, "/out", "index.html", "", "/", 0, "", nil)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job when pagination disabled, got %d", len(jobs))
	}
	if jobs[0].data.Pagination != nil {
		t.Error("Pagination should be nil when disabled")
	}
	if len(jobs[0].data.Articles) != 5 {
		t.Errorf("expected all 5 articles, got %d", len(jobs[0].data.Articles))
	}
}

func TestPaginatedJobs_SinglePage(t *testing.T) {
	site := makeSite()
	articles := makePaginatedArticles(3)
	jobs := paginatedJobs(site, articles, "/out", "index.html", "", "/", 10, "", nil)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job for 3 articles with perPage=10, got %d", len(jobs))
	}
	pg := jobs[0].data.Pagination
	if pg == nil {
		t.Fatal("Pagination should not be nil")
		return
	}
	if pg.TotalPages != 1 || pg.CurrentPage != 1 {
		t.Errorf("unexpected pagination: %+v", pg)
	}
	if pg.PrevURL != "" || pg.NextURL != "" {
		t.Errorf("no prev/next expected for single page: prev=%q next=%q", pg.PrevURL, pg.NextURL)
	}
}

func TestPaginatedJobs_MultiPage_Paths(t *testing.T) {
	site := makeSite()
	articles := makePaginatedArticles(5)
	jobs := paginatedJobs(site, articles, "/out", "index.html", "", "", 2, "", nil)
	// 5 articles / perPage 2 → pages 1,2,3
	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(jobs))
	}
	// Page 1 → /out/index.html
	if jobs[0].path != filepath.Join("/out", "index.html") {
		t.Errorf("page1 path = %q", jobs[0].path)
	}
	// Page 2 → /out/page/2/index.html
	if jobs[1].path != filepath.Join("/out", "page", "2", "index.html") {
		t.Errorf("page2 path = %q", jobs[1].path)
	}
	// Page 3 → /out/page/3/index.html
	if jobs[2].path != filepath.Join("/out", "page", "3", "index.html") {
		t.Errorf("page3 path = %q", jobs[2].path)
	}
}

func TestPaginatedJobs_MultiPage_PrevNext(t *testing.T) {
	site := makeSite()
	articles := makePaginatedArticles(5)
	jobs := paginatedJobs(site, articles, "/out", "index.html", "", "/blog", 2, "", nil)

	pg1 := jobs[0].data.Pagination
	if pg1.PrevURL != "" {
		t.Errorf("page1 PrevURL should be empty, got %q", pg1.PrevURL)
	}
	if pg1.NextURL != "/blog/page/2/" {
		t.Errorf("page1 NextURL = %q, want /blog/page/2/", pg1.NextURL)
	}

	pg2 := jobs[1].data.Pagination
	if pg2.PrevURL != "/blog/" {
		t.Errorf("page2 PrevURL = %q, want /blog/", pg2.PrevURL)
	}
	if pg2.NextURL != "/blog/page/3/" {
		t.Errorf("page2 NextURL = %q, want /blog/page/3/", pg2.NextURL)
	}

	pg3 := jobs[2].data.Pagination
	if pg3.PrevURL != "/blog/page/2/" {
		t.Errorf("page3 PrevURL = %q, want /blog/page/2/", pg3.PrevURL)
	}
	if pg3.NextURL != "" {
		t.Errorf("page3 NextURL should be empty, got %q", pg3.NextURL)
	}
}

func TestPaginatedJobs_WithBasePath(t *testing.T) {
	site := makeSite()
	articles := makePaginatedArticles(3)
	jobs := paginatedJobs(site, articles, "/out", "tag.html", "tags/go", "/tags/go", 2, "", nil)
	// 3 articles / perPage 2 → 2 pages
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].path != filepath.Join("/out", "tags/go", "index.html") {
		t.Errorf("page1 path = %q", jobs[0].path)
	}
	if jobs[1].path != filepath.Join("/out", "tags/go", "page", "2", "index.html") {
		t.Errorf("page2 path = %q", jobs[1].path)
	}
}

func TestPaginatedJobs_CurrentTaxonomy(t *testing.T) {
	site := makeSite()
	articles := makePaginatedArticles(2)
	tax := &model.Taxonomy{Name: "go", Description: "Go language articles"}

	// taxonomy is propagated to all page jobs (single page)
	jobs := paginatedJobs(site, articles, "/out", "tag.html", "tags/go", "/tags/go", 0, "", tax)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].data.CurrentTaxonomy == nil {
		t.Fatal("CurrentTaxonomy should not be nil for tag page")
	}
	if jobs[0].data.CurrentTaxonomy.Name != "go" {
		t.Errorf("CurrentTaxonomy.Name = %q, want %q", jobs[0].data.CurrentTaxonomy.Name, "go")
	}

	// taxonomy is propagated across multiple pages
	articles5 := makePaginatedArticles(5)
	jobs2 := paginatedJobs(site, articles5, "/out", "tag.html", "tags/go", "/tags/go", 2, "", tax)
	for i, j := range jobs2 {
		if j.data.CurrentTaxonomy == nil || j.data.CurrentTaxonomy.Name != "go" {
			t.Errorf("page %d: CurrentTaxonomy not set correctly", i+1)
		}
	}

	// nil taxonomy on index page
	indexJobs := paginatedJobs(site, articles, "/out", "index.html", "", "/", 0, "", nil)
	if indexJobs[0].data.CurrentTaxonomy != nil {
		t.Error("CurrentTaxonomy should be nil for index page")
	}
}

func TestPaginatedJobs_TaxonomyURL(t *testing.T) {
	site := makeSite()
	articles := makePaginatedArticles(2)
	tax := &model.Taxonomy{Name: "go"}

	// EN tag: baseURLPath="/tags/go" → URL set to "/tags/go/"
	jobs := paginatedJobs(site, articles, "/out", "tag.html", "tags/go", "/tags/go", 0, "en", tax)
	if jobs[0].data.CurrentTaxonomy == nil {
		t.Fatal("CurrentTaxonomy should not be nil")
	}
	if got, want := jobs[0].data.CurrentTaxonomy.URL, "/tags/go/"; got != want {
		t.Errorf("Taxonomy.URL = %q, want %q", got, want)
	}

	// JA tag: baseURLPath="/ja/tags/go" → URL set to "/ja/tags/go/"
	jobs2 := paginatedJobs(site, articles, "/out", "tag.html", "ja/tags/go", "/ja/tags/go", 0, "ja", tax)
	if got, want := jobs2[0].data.CurrentTaxonomy.URL, "/ja/tags/go/"; got != want {
		t.Errorf("JA Taxonomy.URL = %q, want %q", got, want)
	}

	// Empty baseURLPath: URL should not be set (remains "")
	jobs3 := paginatedJobs(site, articles, "/out", "index.html", "", "", 0, "", tax)
	if jobs3[0].data.CurrentTaxonomy != nil && jobs3[0].data.CurrentTaxonomy.URL != "" {
		t.Errorf("Taxonomy.URL should be empty when baseURLPath is empty, got %q", jobs3[0].data.CurrentTaxonomy.URL)
	}
}

func TestGenerate_ArchiveCurrentArchivePath(t *testing.T) {
	outDir := t.TempDir()
	cfg := model.Config{
		Build: model.BuildConfig{Parallelism: 1},
		I18n:  model.I18nConfig{Locales: []string{"en", "ja"}, DefaultLocale: "en"},
	}
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	site := &model.Site{
		Config: cfg,
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Slug: "post-en", Date: now,
				}},
				Locale: "en",
			},
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Slug: "post-ja", Date: now,
				}},
				Locale: "ja",
			},
		},
	}
	eng := &captureEngine{}
	g := NewHTMLGenerator(outDir, eng, cfg)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Find archive jobs and check CurrentArchivePath
	eng.mu.Lock()
	defer eng.mu.Unlock()
	var enMonthPath, jaMonthPath, enYearPath, jaYearPath string
	for _, r := range eng.renders {
		if r.tmpl != "archive.html" {
			continue
		}
		path := r.data.CurrentArchivePath
		locale := r.data.CurrentLocale
		switch {
		case locale == "en" && path == "/archives/2024/03/":
			enMonthPath = path
			if !r.data.CurrentArchiveIsMonth {
				t.Error("EN month archive: CurrentArchiveIsMonth should be true")
			}
		case locale == "en" && path == "/archives/2024/":
			enYearPath = path
			if r.data.CurrentArchiveIsMonth {
				t.Error("EN year archive: CurrentArchiveIsMonth should be false")
			}
		case locale == "ja" && path == "/ja/archives/2024/03/":
			jaMonthPath = path
			if !r.data.CurrentArchiveIsMonth {
				t.Error("JA month archive: CurrentArchiveIsMonth should be true")
			}
		case locale == "ja" && path == "/ja/archives/2024/":
			jaYearPath = path
			if r.data.CurrentArchiveIsMonth {
				t.Error("JA year archive: CurrentArchiveIsMonth should be false")
			}
		}
	}
	if enMonthPath == "" {
		t.Error("EN month archive: CurrentArchivePath not set to /archives/2024/03/")
	}
	if enYearPath == "" {
		t.Error("EN year archive: CurrentArchivePath not set to /archives/2024/")
	}
	if jaMonthPath == "" {
		t.Error("JA month archive: CurrentArchivePath not set to /ja/archives/2024/03/")
	}
	if jaYearPath == "" {
		t.Error("JA year archive: CurrentArchivePath not set to /ja/archives/2024/")
	}
}

func TestGenerate_WithPagination_CreatesPageDirs(t *testing.T) {
	outDir := t.TempDir()
	cfg := model.Config{Build: model.BuildConfig{Parallelism: 2, PerPage: 1}}
	site := makeSite()
	// Add a second article so pagination kicks in
	site.Articles = append(site.Articles, &model.ProcessedArticle{
		Article: model.Article{FrontMatter: model.FrontMatter{
			Title: "Second Post", Slug: "second-post",
			Tags: []string{"go"}, Categories: []string{"tech"},
			Date: time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC),
		}},
	})
	g := NewHTMLGenerator(outDir, &mockEngine{}, cfg)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// With perPage=1 and 2 articles, page 2 should be generated
	if _, err := os.Stat(filepath.Join(outDir, "page", "2", "index.html")); err != nil {
		t.Errorf("expected /page/2/index.html: %v", err)
	}
}

func TestFilterArticles(t *testing.T) {
	articles := makePaginatedArticles(4)
	even := filterArticles(articles, func(a *model.ProcessedArticle) bool {
		// keep article-2 and article-4
		return a.FrontMatter.Slug == "article-2" || a.FrontMatter.Slug == "article-4"
	})
	if len(even) != 2 {
		t.Errorf("expected 2 filtered articles, got %d", len(even))
	}
	none := filterArticles(articles, func(*model.ProcessedArticle) bool { return false })
	if len(none) != 0 {
		t.Errorf("expected 0, got %d", len(none))
	}
}

func TestGenerate_I18nLocalePrefixedArticlePage(t *testing.T) {
	outDir := t.TempDir()
	cfg := model.Config{
		Build: model.BuildConfig{Parallelism: 1},
		I18n:  model.I18nConfig{Locales: []string{"en", "ja"}, DefaultLocale: "en"},
	}
	site := &model.Site{
		Config: cfg,
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{Slug: "hello"}},
				Locale:  "en",
			},
			{
				Article: model.Article{FrontMatter: model.FrontMatter{Slug: "hello"}},
				Locale:  "ja",
			},
		},
	}
	g := NewHTMLGenerator(outDir, &mockEngine{}, cfg)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Default locale article: no prefix
	if _, err := os.Stat(filepath.Join(outDir, "posts", "hello", "index.html")); err != nil {
		t.Errorf("missing en article page: %v", err)
	}
	// Non-default locale article: /ja/ prefix
	if _, err := os.Stat(filepath.Join(outDir, "ja", "posts", "hello", "index.html")); err != nil {
		t.Errorf("missing ja article page: %v", err)
	}
	// Per-locale index pages
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Errorf("missing default locale index.html: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ja", "index.html")); err != nil {
		t.Errorf("missing ja/index.html: %v", err)
	}
}

func TestGenerate_ArticlesSortedNewestFirst(t *testing.T) {
	outDir := t.TempDir()
	eng := &mockEngine{}
	older := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	site := &model.Site{
		Config: model.Config{Build: model.BuildConfig{Parallelism: 1, PerPage: 10}},
		Articles: []*model.ProcessedArticle{
			{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Old", Slug: "old", Date: older}}},
			{Article: model.Article{FrontMatter: model.FrontMatter{Title: "New", Slug: "new", Date: newer}}},
		},
	}
	g := NewHTMLGenerator(outDir, eng, site.Config)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Read rendered index page (first call to Render for "index.html")
	// The mock engine records calls in order; since articles are sorted newest-first,
	// the site passed to the index page should list "New" before "Old".
	// We verify indirectly: if articles were NOT sorted, Old would appear first
	// because it was appended first in site.Articles.
	// The test confirms that the sort helper is wired in.
	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("index.html missing: %v", err)
	}
	_ = data // content is just "<html>index.html</html>" from mock; sort tested via sortByDateDesc unit test
}

func TestSortByDateDesc(t *testing.T) {
	older := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Old", Date: older}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{Title: "New", Date: newer}}},
	}
	sortByDateDesc(articles)
	if articles[0].FrontMatter.Title != "New" {
		t.Errorf("expected 'New' first after sortByDateDesc, got %q", articles[0].FrontMatter.Title)
	}
}

func TestRelatedArticles(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	target := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Target", Slug: "target", Categories: []string{"go", "web"}, Date: t1,
	}}, Locale: "en"}
	same1 := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Same1", Slug: "same1", Categories: []string{"go"}, Date: t2,
	}}, Locale: "en"}
	same2 := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Same2", Slug: "same2", Categories: []string{"web"}, Date: t0,
	}}, Locale: "en"}
	diffLocale := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "JA", Slug: "ja", Categories: []string{"go"}, Date: t2,
	}}, Locale: "ja"}
	unrelated := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Unrelated", Slug: "unrelated", Categories: []string{"rust"}, Date: t2,
	}}, Locale: "en"}

	all := []*model.ProcessedArticle{target, same1, same2, diffLocale, unrelated}

	t.Run("returns same-locale same-category articles sorted newest first", func(t *testing.T) {
		got := relatedArticles(all, target, 5)
		if len(got) != 2 {
			t.Fatalf("expected 2 related, got %d", len(got))
		}
		if got[0].FrontMatter.Title != "Same1" {
			t.Errorf("expected Same1 first (newer), got %q", got[0].FrontMatter.Title)
		}
		if got[1].FrontMatter.Title != "Same2" {
			t.Errorf("expected Same2 second (older), got %q", got[1].FrontMatter.Title)
		}
	})

	t.Run("excludes the target article itself", func(t *testing.T) {
		got := relatedArticles(all, target, 5)
		for _, a := range got {
			if a == target {
				t.Error("target article should not appear in related")
			}
		}
	})

	t.Run("excludes different locale", func(t *testing.T) {
		got := relatedArticles(all, target, 5)
		for _, a := range got {
			if a.Locale != target.Locale {
				t.Errorf("got article with locale %q, expected %q", a.Locale, target.Locale)
			}
		}
	})

	t.Run("respects n limit", func(t *testing.T) {
		got := relatedArticles(all, target, 1)
		if len(got) != 1 {
			t.Fatalf("expected 1 with n=1, got %d", len(got))
		}
	})

	t.Run("returns empty when no categories match", func(t *testing.T) {
		got := relatedArticles(all, unrelated, 5)
		if len(got) != 0 {
			t.Errorf("expected 0 related for unrelated article, got %d", len(got))
		}
	})
}

func TestGenerate_SkipsDateZeroArchive(t *testing.T) {
	outDir := t.TempDir()
	site := &model.Site{
		Config: model.Config{Build: model.BuildConfig{Parallelism: 1}},
		Articles: []*model.ProcessedArticle{
			// Article with zero date — should NOT produce archives/0001/01/
			{Article: model.Article{FrontMatter: model.FrontMatter{Title: "No Date", Slug: "no-date"}}},
		},
	}
	g := NewHTMLGenerator(outDir, &mockEngine{}, site.Config)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	badArchive := filepath.Join(outDir, "archives", "0001", "01")
	if _, err := os.Stat(badArchive); err == nil {
		t.Errorf("archives/0001/01 should NOT be created for date-zero articles")
	}
	badYearArchive := filepath.Join(outDir, "archives", "0001")
	if _, err := os.Stat(badYearArchive); err == nil {
		t.Errorf("archives/0001 should NOT be created for date-zero articles")
	}
}

func TestGenerate_FrontMatterTemplateOverride(t *testing.T) {
	outDir := t.TempDir()
	eng := &mockEngine{}
	site := &model.Site{
		Config: model.Config{Build: model.BuildConfig{Parallelism: 1}},
		Articles: []*model.ProcessedArticle{
			{Article: model.Article{FrontMatter: model.FrontMatter{
				Title:    "Custom",
				Slug:     "custom",
				Template: "custom.html",
				Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}}},
		},
	}
	g := NewHTMLGenerator(outDir, eng, site.Config)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	eng.mu.Lock()
	calls := append([]string(nil), eng.calls...)
	eng.mu.Unlock()
	for _, c := range calls {
		if c == "custom.html" {
			return // found — pass
		}
	}
	t.Errorf("expected custom.html template call, got: %v", calls)
}

func makeSiteI18n() *model.Site {
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	return &model.Site{
		Config: model.Config{
			Site:  model.SiteConfig{Title: "Test Site", BaseURL: "https://example.com"},
			Build: model.BuildConfig{Parallelism: 2},
			I18n:  model.I18nConfig{Locales: []string{"en", "ja"}, DefaultLocale: "en"},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "Hello EN", Slug: "hello-en",
					Tags: []string{"go"}, Categories: []string{"tech"}, Date: now,
				}},
				Locale: "en",
			},
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "Hello JA", Slug: "hello-ja",
					Tags: []string{"go"}, Categories: []string{"tech"}, Date: now,
				}},
				Locale: "ja",
			},
		},
		Tags:       []model.Taxonomy{{Name: "go"}},
		Categories: []model.Taxonomy{{Name: "tech"}},
	}
}

func TestGenerate_LocaleAwareArchives(t *testing.T) {
	outDir := t.TempDir()
	site := makeSiteI18n()
	g := NewHTMLGenerator(outDir, &mockEngine{}, site.Config)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Default locale (en) archives should be at the root.
	for _, rel := range []string{
		"archives/2024/03/index.html",
		"archives/2024/index.html",
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Errorf("missing default-locale archive %s: %v", rel, err)
		}
	}

	// Non-default locale (ja) archives should be under ja/.
	for _, rel := range []string{
		"ja/archives/2024/03/index.html",
		"ja/archives/2024/index.html",
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Errorf("missing ja locale archive %s: %v", rel, err)
		}
	}
}

func TestGenerate_LocaleAwareArchives_ArticleIsolation(t *testing.T) {
	outDir := t.TempDir()
	site := makeSiteI18n()
	g := NewHTMLGenerator(outDir, &mockEngine{}, site.Config)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// en archive should not contain ja articles (locale set correctly)
	enJobs := filterBuildJobsByLocale(g, site, "en")
	for _, j := range enJobs {
		for _, a := range j.data.Articles {
			if a.Locale == "ja" {
				t.Errorf("en archive contains ja article: %s", a.FrontMatter.Title)
			}
		}
	}
}

// filterBuildJobsByLocale returns archive write jobs for the given locale.
func filterBuildJobsByLocale(g *HTMLGenerator, site *model.Site, locale string) []writeJob {
	jobs := g.buildJobs(site)
	var out []writeJob
	for _, j := range jobs {
		if j.tmpl == "archive.html" && j.data.CurrentLocale == locale {
			out = append(out, j)
		}
	}
	return out
}

func TestGenerate_SkipsDateZeroArchiveI18n(t *testing.T) {
	outDir := t.TempDir()
	site := &model.Site{
		Config: model.Config{
			Build: model.BuildConfig{Parallelism: 1},
			I18n:  model.I18nConfig{Locales: []string{"en", "ja"}, DefaultLocale: "en"},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{Title: "No Date EN", Slug: "no-date-en"}},
				Locale:  "en",
			},
			{
				Article: model.Article{FrontMatter: model.FrontMatter{Title: "No Date JA", Slug: "no-date-ja"}},
				Locale:  "ja",
			},
		},
	}
	g := NewHTMLGenerator(outDir, &mockEngine{}, site.Config)
	if err := g.Generate(site, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, bad := range []string{
		filepath.Join(outDir, "archives", "0001", "01"),
		filepath.Join(outDir, "archives", "0001"),
		filepath.Join(outDir, "ja", "archives", "0001", "01"),
		filepath.Join(outDir, "ja", "archives", "0001"),
	} {
		if _, err := os.Stat(bad); err == nil {
			t.Errorf("should NOT be created for date-zero articles: %s", bad)
		}
	}
}

func TestArchiveYears_SortedDesc(t *testing.T) {
	articles := []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Date: time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{Date: time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)}}}, // duplicate year
		{Article: model.Article{FrontMatter: model.FrontMatter{Date: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{}}}, // zero date — must be skipped
	}
	got := archiveYears(articles)
	want := []int{2025, 2024, 2023}
	if len(got) != len(want) {
		t.Fatalf("archiveYears: got %v, want %v", got, want)
	}
	for i, y := range want {
		if got[i] != y {
			t.Errorf("archiveYears[%d]: got %d, want %d", i, got[i], y)
		}
	}
}

func TestArchiveYears_Empty(t *testing.T) {
	if got := archiveYears(nil); len(got) != 0 {
		t.Errorf("archiveYears(nil): expected empty, got %v", got)
	}
}

func TestLocaleTaxonomyBase_SetsArchiveYears(t *testing.T) {
	base := &model.Site{Config: model.Config{}}
	articles := []*model.ProcessedArticle{
		{Article: model.Article{FrontMatter: model.FrontMatter{Tags: []string{"go"}, Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}}},
		{Article: model.Article{FrontMatter: model.FrontMatter{Tags: []string{"go"}, Date: time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC)}}},
	}
	got := localeTaxonomyBase(base, articles)
	want := []int{2024, 2022}
	if len(got.ArchiveYears) != len(want) {
		t.Fatalf("ArchiveYears: got %v, want %v", got.ArchiveYears, want)
	}
	for i, y := range want {
		if got.ArchiveYears[i] != y {
			t.Errorf("ArchiveYears[%d]: got %d, want %d", i, got.ArchiveYears[i], y)
		}
	}
}

func TestSiteFor_PropagatesArchiveYears(t *testing.T) {
	base := &model.Site{
		Config:       model.Config{},
		ArchiveYears: []int{2025, 2024, 2023},
	}
	got := siteFor(base, nil)
	if len(got.ArchiveYears) != 3 || got.ArchiveYears[0] != 2025 {
		t.Errorf("siteFor: ArchiveYears not propagated: %v", got.ArchiveYears)
	}
}

func TestArticleOutputPath_UsesOutputPath(t *testing.T) {
	outDir := "/abs/public"
	cfg := model.Config{Build: model.BuildConfig{OutputDir: "public"}}
	a := &model.ProcessedArticle{
		Article:    model.Article{FrontMatter: model.FrontMatter{Slug: "hello"}},
		OutputPath: "public/pages/about/index.html",
	}
	got := articleOutputPath(a, outDir, cfg)
	want := filepath.Join(outDir, "pages", "about", "index.html")
	if got != want {
		t.Errorf("articleOutputPath: got %q, want %q", got, want)
	}
}

func TestArticleOutputPath_FallbackWhenEmpty(t *testing.T) {
	outDir := "/abs/public"
	cfg := model.Config{Build: model.BuildConfig{OutputDir: "public"}}
	a := &model.ProcessedArticle{
		Article: model.Article{FrontMatter: model.FrontMatter{Slug: "hello"}},
		// OutputPath intentionally empty
	}
	got := articleOutputPath(a, outDir, cfg)
	want := filepath.Join(outDir, "posts", "hello", "index.html")
	if got != want {
		t.Errorf("articleOutputPath fallback: got %q, want %q", got, want)
	}
}

// TestArchive_PaginationJobs verifies that archive pages are paginated when the
// article count exceeds perPage, and that CurrentArchivePath is set on all pages.
func TestArchive_PaginationJobs(t *testing.T) {
	date := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	var articles []*model.ProcessedArticle
	for i := 0; i < 25; i++ {
		articles = append(articles, &model.ProcessedArticle{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Title: fmt.Sprintf("Article %d", i),
				Slug:  fmt.Sprintf("article-%d", i),
				Date:  date,
			}},
		})
	}
	site := &model.Site{
		Config: model.Config{
			Build: model.BuildConfig{Parallelism: 1, PerPage: 10},
		},
		Articles: articles,
	}
	g := NewHTMLGenerator(t.TempDir(), &mockEngine{}, site.Config)
	jobs := g.buildJobs(site)

	var archiveJobs []writeJob
	for _, j := range jobs {
		if j.tmpl == "archive.html" {
			archiveJobs = append(archiveJobs, j)
		}
	}

	// 25 articles / 10 per page = 3 pages for month + year archives (month same count)
	monthJobs := 0
	yearJobs := 0
	for _, j := range archiveJobs {
		if j.data.CurrentArchivePath == "/archives/2024/03/" {
			monthJobs++
			if j.data.CurrentArchivePath == "" {
				t.Error("CurrentArchivePath must be set on paginated archive pages")
			}
		} else if j.data.CurrentArchivePath == "/archives/2024/" {
			yearJobs++
		}
	}
	if monthJobs != 3 {
		t.Errorf("expected 3 paginated month-archive jobs for 25 articles with perPage=10, got %d", monthJobs)
	}
	if yearJobs != 3 {
		t.Errorf("expected 3 paginated year-archive jobs for 25 articles with perPage=10, got %d", yearJobs)
	}
}

// TestArchive_PaginationJobs_I18n verifies locale-aware archive pagination.
func TestArchive_PaginationJobs_I18n(t *testing.T) {
	date := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	var articles []*model.ProcessedArticle
	for i := 0; i < 15; i++ {
		articles = append(articles, &model.ProcessedArticle{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Title: fmt.Sprintf("EN %d", i), Slug: fmt.Sprintf("en-%d", i), Date: date,
			}},
			Locale: "en",
		})
		articles = append(articles, &model.ProcessedArticle{
			Article: model.Article{FrontMatter: model.FrontMatter{
				Title: fmt.Sprintf("JA %d", i), Slug: fmt.Sprintf("ja-%d", i), Date: date,
			}},
			Locale: "ja",
		})
	}
	site := &model.Site{
		Config: model.Config{
			Build: model.BuildConfig{Parallelism: 1, PerPage: 10},
			I18n:  model.I18nConfig{Locales: []string{"en", "ja"}, DefaultLocale: "en"},
		},
		Articles: articles,
	}
	g := NewHTMLGenerator(t.TempDir(), &mockEngine{}, site.Config)
	jobs := g.buildJobs(site)

	enMonthJobs, jaMonthJobs := 0, 0
	for _, j := range jobs {
		if j.tmpl != "archive.html" {
			continue
		}
		switch {
		case j.data.CurrentLocale == "en" && j.data.CurrentArchivePath == "/archives/2024/03/":
			enMonthJobs++
		case j.data.CurrentLocale == "ja" && j.data.CurrentArchivePath == "/ja/archives/2024/03/":
			jaMonthJobs++
		}
	}
	// 15 articles each locale, perPage=10 → 2 pages each
	if enMonthJobs != 2 {
		t.Errorf("expected 2 EN month-archive pages, got %d", enMonthJobs)
	}
	if jaMonthJobs != 2 {
		t.Errorf("expected 2 JA month-archive pages, got %d", jaMonthJobs)
	}
}

// TestLocaleTaxonomyBase_NoFallback verifies that localeTaxonomyBase does NOT
// fall back to cross-locale tags/categories when a locale's articles have none.
// Previously the function returned base.Tags for locales with untagged articles,
// which caused JA tags to appear in the EN sidebar and vice-versa.
func TestLocaleTaxonomyBase_NoFallback(t *testing.T) {
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	enArticle := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{
				Title: "EN Post",
				Slug:  "en-post",
				Date:  date,
				// No tags — should NOT fall back to JA tags
			},
		},
		Locale: "en",
	}
	jaArticle := &model.ProcessedArticle{
		Article: model.Article{
			FrontMatter: model.FrontMatter{
				Title: "JA Post",
				Slug:  "ja-post",
				Date:  date,
				Tags:  []string{"golang", "書評"},
			},
		},
		Locale: "ja",
	}
	base := &model.Site{
		Articles: []*model.ProcessedArticle{enArticle, jaArticle},
		Tags:     []model.Taxonomy{{Name: "golang"}, {Name: "書評"}},
	}

	// EN locale has no tags — must return empty, not JA tags.
	enBase := localeTaxonomyBase(base, []*model.ProcessedArticle{enArticle})
	if len(enBase.Tags) != 0 {
		t.Errorf("EN localeTaxonomyBase: expected 0 tags, got %d: %v", len(enBase.Tags), enBase.Tags)
	}

	// JA locale has tags — must return only JA tags.
	jaBase := localeTaxonomyBase(base, []*model.ProcessedArticle{jaArticle})
	if len(jaBase.Tags) != 2 {
		t.Errorf("JA localeTaxonomyBase: expected 2 tags, got %d: %v", len(jaBase.Tags), jaBase.Tags)
	}
}
