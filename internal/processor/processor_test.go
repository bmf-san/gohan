package processor

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

func testArticle(path, title, raw string, tags, cats []string, date time.Time) *model.Article {
	return &model.Article{
		FrontMatter: model.FrontMatter{
			Title:      title,
			Tags:       tags,
			Categories: cats,
			Date:       date,
		},
		RawContent:   raw,
		FilePath:     path,
		LastModified: time.Now(),
	}
}

func TestSiteProcessor_Process_Basic(t *testing.T) {
	p := NewSiteProcessor()
	articles := []*model.Article{
		testArticle("content/posts/hello.md", "Hello", "# Hello\n\nWorld.", []string{"go"}, nil, time.Time{}),
	}
	cfg := model.Config{Build: model.BuildConfig{ContentDir: "content", OutputDir: "public"}}
	processed, err := p.Process(articles, cfg)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if len(processed) != 1 {
		t.Fatalf("expected 1 article, got %d", len(processed))
	}
	a := processed[0]
	if !strings.Contains(string(a.HTMLContent), "<h1") {
		t.Errorf("HTMLContent missing h1: %q", a.HTMLContent)
	}
	if a.Summary == "" {
		t.Error("Summary should not be empty")
	}
	if a.OutputPath == "" {
		t.Error("OutputPath should not be empty")
	}
}

func TestSiteProcessor_Process_OutputPath_WithSlug(t *testing.T) {
	p := NewSiteProcessor()
	a := &model.Article{
		FrontMatter:  model.FrontMatter{Slug: "my-slug"},
		RawContent:   "content",
		FilePath:     "content/posts/post.md",
		LastModified: time.Now(),
	}
	cfg := model.Config{Build: model.BuildConfig{ContentDir: "content", OutputDir: "public"}}
	processed, err := p.Process([]*model.Article{a}, cfg)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	expected := filepath.Join("public", "posts", "my-slug", "index.html")
	if processed[0].OutputPath != expected {
		t.Errorf("OutputPath: got %q, want %q", processed[0].OutputPath, expected)
	}
}

func TestSiteProcessor_Process_OutputPath_NoSlug(t *testing.T) {
	p := NewSiteProcessor()
	a := &model.Article{
		FilePath:     "content/posts/my-post.md",
		RawContent:   "content",
		LastModified: time.Now(),
	}
	cfg := model.Config{Build: model.BuildConfig{ContentDir: "content", OutputDir: "public"}}
	processed, err := p.Process([]*model.Article{a}, cfg)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	expected := filepath.Join("public", "posts", "my-post", "index.html")
	if processed[0].OutputPath != expected {
		t.Errorf("OutputPath: got %q, want %q", processed[0].OutputPath, expected)
	}
}

func TestSiteProcessor_BuildDependencyGraph(t *testing.T) {
	p := NewSiteProcessor()
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "A", "", []string{"go", "ssg"}, []string{"news"}, date)},
		{Article: *testArticle("b.md", "B", "", []string{"go"}, nil, time.Time{})},
	}
	g, err := p.BuildDependencyGraph(articles)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}
	if _, ok := g.Nodes["a.md"]; !ok {
		t.Error("expected node for a.md")
	}
	if _, ok := g.Nodes["tag:go"]; !ok {
		t.Error("expected tag:go node")
	}
	if _, ok := g.Nodes["tag:ssg"]; !ok {
		t.Error("expected tag:ssg node")
	}
	if _, ok := g.Nodes["category:news"]; !ok {
		t.Error("expected category:news node")
	}
	if _, ok := g.Nodes["archive:2024"]; !ok {
		t.Error("expected archive:2024 node")
	}
	if len(g.Edges["a.md"]) != 4 {
		t.Errorf("a.md: expected 4 edges, got %d: %v", len(g.Edges["a.md"]), g.Edges["a.md"])
	}
}

func TestSiteProcessor_BuildTaxonomyRegistry(t *testing.T) {
	p := NewSiteProcessor()
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "A", "", []string{"go", "ssg"}, []string{"web"}, time.Time{})},
		{Article: *testArticle("b.md", "B", "", []string{"go"}, []string{"news"}, time.Time{})},
	}
	reg, err := p.BuildTaxonomyRegistry(articles, model.Config{})
	if err != nil {
		t.Fatalf("BuildTaxonomyRegistry: %v", err)
	}
	if len(reg.Tags) != 2 {
		t.Errorf("Tags: got %d, want 2", len(reg.Tags))
	}
	if len(reg.Categories) != 2 {
		t.Errorf("Categories: got %d, want 2", len(reg.Categories))
	}
}

func TestCalculateImpact(t *testing.T) {
	g := &model.DependencyGraph{
		Nodes: map[string]*model.Node{
			"article.md": {Path: "article.md", Type: model.NodeTypeArticle, Dependents: []string{}},
			"tag:go":     {Path: "tag:go", Type: model.NodeTypeTag, Dependents: []string{"article.md"}},
		},
		Edges: map[string][]string{
			"article.md": {"tag:go"},
		},
	}
	impact := CalculateImpact(g, "tag:go")
	if len(impact) != 2 {
		t.Errorf("expected 2 impacted nodes, got %d: %v", len(impact), impact)
	}
}

func TestCalculateDiff(t *testing.T) {
	old := &model.DependencyGraph{
		Nodes: map[string]*model.Node{
			"a.md": {Path: "a.md"},
			"b.md": {Path: "b.md"},
		},
		Edges: map[string][]string{},
	}
	newG := &model.DependencyGraph{
		Nodes: map[string]*model.Node{
			"a.md": {Path: "a.md"},
			"c.md": {Path: "c.md"},
		},
		Edges: map[string][]string{},
	}
	cs, err := CalculateDiff(old, newG)
	if err != nil {
		t.Fatalf("CalculateDiff: %v", err)
	}
	if len(cs.AddedFiles) != 1 || cs.AddedFiles[0] != "c.md" {
		t.Errorf("AddedFiles: %v", cs.AddedFiles)
	}
	if len(cs.DeletedFiles) != 1 || cs.DeletedFiles[0] != "b.md" {
		t.Errorf("DeletedFiles: %v", cs.DeletedFiles)
	}
	if len(cs.ModifiedFiles) != 1 || cs.ModifiedFiles[0] != "a.md" {
		t.Errorf("ModifiedFiles: %v", cs.ModifiedFiles)
	}
}

func TestCalculateDiff_NilGraph(t *testing.T) {
	_, err := CalculateDiff(nil, &model.DependencyGraph{Nodes: map[string]*model.Node{}, Edges: map[string][]string{}})
	if err == nil {
		t.Error("expected error for nil oldGraph")
	}
}

func TestExtractSummary_FirstParagraph(t *testing.T) {
	content := "First paragraph here.\n\nSecond paragraph."
	got := extractSummary(content, 200)
	if got != "First paragraph here." {
		t.Errorf("extractSummary: got %q", got)
	}
}

func TestExtractSummary_Truncate(t *testing.T) {
	content := strings.Repeat("x", 300)
	got := extractSummary(content, 200)
	if len(got) > 204 {
		t.Errorf("extractSummary too long: %d chars", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("expected truncation marker ...")
	}
}

// ---- i18n tests ----

func i18nCfg() model.Config {
	return model.Config{
		Build: model.BuildConfig{ContentDir: "content", OutputDir: "public"},
		I18n:  model.I18nConfig{Locales: []string{"en", "ja"}, DefaultLocale: "en"},
	}
}

func TestDetectLocale(t *testing.T) {
	cfg := i18nCfg()
	tests := []struct {
		filePath string
		want     string
	}{
		{"content/en/posts/hello.md", "en"},
		{"content/ja/posts/hello.md", "ja"},
		{"content/posts/hello.md", ""},  // no matching locale segment
		{"other/en/posts/hello.md", ""}, // outside content dir
	}
	for _, tc := range tests {
		a := &model.Article{FilePath: tc.filePath}
		got := detectLocale(a, cfg)
		if got != tc.want {
			t.Errorf("detectLocale(%q) = %q, want %q", tc.filePath, got, tc.want)
		}
	}
}

func TestDetectLocale_NoI18n(t *testing.T) {
	cfg := model.Config{Build: model.BuildConfig{ContentDir: "content", OutputDir: "public"}}
	a := &model.Article{FilePath: "content/en/posts/hello.md"}
	if got := detectLocale(a, cfg); got != "" {
		t.Errorf("detectLocale without i18n config: got %q, want empty", got)
	}
}

func TestComputeOutputPath_I18n(t *testing.T) {
	cfg := i18nCfg()
	tests := []struct {
		filePath string
		want     string
	}{
		// Default locale (en) → no URL prefix
		{"content/en/posts/hello.md", filepath.Join("public", "posts", "hello", "index.html")},
		// Non-default locale (ja) → /ja/ prefix
		{"content/ja/posts/hello.md", filepath.Join("public", "ja", "posts", "hello", "index.html")},
	}
	for _, tc := range tests {
		a := &model.Article{FilePath: tc.filePath}
		got := computeOutputPath(a, cfg)
		if got != tc.want {
			t.Errorf("computeOutputPath(%q) = %q, want %q", tc.filePath, got, tc.want)
		}
	}
}

func TestComputeArticleURL(t *testing.T) {
	cfg := i18nCfg()
	tests := []struct {
		filePath string
		want     string
	}{
		{"content/en/posts/hello.md", "/posts/hello/"},
		{"content/ja/posts/hello.md", "/ja/posts/hello/"},
	}
	for _, tc := range tests {
		a := &model.Article{FilePath: tc.filePath}
		got := computeArticleURL(a, cfg)
		if got != tc.want {
			t.Errorf("computeArticleURL(%q) = %q, want %q", tc.filePath, got, tc.want)
		}
	}
}

func TestComputeArticleURL_NoI18n(t *testing.T) {
	cfg := model.Config{Build: model.BuildConfig{ContentDir: "content", OutputDir: "public"}}
	a := &model.Article{FilePath: "content/posts/hello.md"}
	if got := computeArticleURL(a, cfg); got != "" {
		t.Errorf("computeArticleURL without i18n: got %q, want empty", got)
	}
}

func TestBuildTranslationMap(t *testing.T) {
	p := NewSiteProcessor()
	articles := []*model.ProcessedArticle{
		{
			Article: model.Article{FrontMatter: model.FrontMatter{TranslationKey: "hello"}},
			Locale:  "en",
			URL:     "/posts/hello/",
		},
		{
			Article: model.Article{FrontMatter: model.FrontMatter{TranslationKey: "hello"}},
			Locale:  "ja",
			URL:     "/ja/posts/hello/",
		},
		{
			Article: model.Article{FrontMatter: model.FrontMatter{TranslationKey: ""}},
			Locale:  "en",
			URL:     "/posts/other/",
		},
	}
	p.BuildTranslationMap(articles)

	enArt := articles[0]
	if len(enArt.Translations) != 1 {
		t.Fatalf("en article: expected 1 translation, got %d", len(enArt.Translations))
	}
	if enArt.Translations[0].Locale != "ja" || enArt.Translations[0].URL != "/ja/posts/hello/" {
		t.Errorf("en Translations[0] = %+v", enArt.Translations[0])
	}

	jaArt := articles[1]
	if len(jaArt.Translations) != 1 {
		t.Fatalf("ja article: expected 1 translation, got %d", len(jaArt.Translations))
	}
	if jaArt.Translations[0].Locale != "en" || jaArt.Translations[0].URL != "/posts/hello/" {
		t.Errorf("ja Translations[0] = %+v", jaArt.Translations[0])
	}

	// Article without TranslationKey should remain unchanged.
	if len(articles[2].Translations) != 0 {
		t.Errorf("unkeyed article should have no translations")
	}
}

func TestProcess_I18nLocaleAndURL(t *testing.T) {
	p := NewSiteProcessor()
	cfg := i18nCfg()
	articles := []*model.Article{
		{FilePath: "content/en/posts/hello.md", RawContent: "Hello", LastModified: time.Now()},
		{FilePath: "content/ja/posts/hello.md", RawContent: "こんにちは", LastModified: time.Now()},
	}
	processed, err := p.Process(articles, cfg)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	for _, a := range processed {
		if a.Locale == "" {
			t.Errorf("Locale not set for %s", a.FilePath)
		}
		if a.URL == "" {
			t.Errorf("URL not set for %s", a.FilePath)
		}
	}
	// Verify actual URLs.
	byLocale := map[string]string{}
	for _, a := range processed {
		byLocale[a.Locale] = a.URL
	}
	if byLocale["en"] != "/posts/hello/" {
		t.Errorf("en URL = %q, want /posts/hello/", byLocale["en"])
	}
	if byLocale["ja"] != "/ja/posts/hello/" {
		t.Errorf("ja URL = %q, want /ja/posts/hello/", byLocale["ja"])
	}
}

func TestComputeContentPath(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		contentDir string
		want       string
	}{
		{
			name:       "nested post",
			filePath:   "content/posts/hello.md",
			contentDir: "content",
			want:       "posts/hello.md",
		},
		{
			name:       "root-level file",
			filePath:   "content/index.md",
			contentDir: "content",
			want:       "index.md",
		},
		{
			name:       "deeply nested",
			filePath:   "content/a/b/c.md",
			contentDir: "content",
			want:       "a/b/c.md",
		},
		{
			name:       "fallback when Rel fails (no common prefix)",
			filePath:   "other/posts/x.md",
			contentDir: "content",
			want: filepath.ToSlash(func() string {
				rel, err := filepath.Rel("content", "other/posts/x.md")
				if err != nil {
					return filepath.Base("other/posts/x.md")
				}
				return filepath.ToSlash(rel)
			}()),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &model.Article{FilePath: tc.filePath}
			cfg := model.Config{Build: model.BuildConfig{ContentDir: tc.contentDir}}
			got := computeContentPath(a, cfg)
			if got != tc.want {
				t.Errorf("computeContentPath: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProcess_ContentPathSet(t *testing.T) {
	p := NewSiteProcessor()
	a := &model.Article{
		FilePath:     "content/posts/my-post.md",
		RawContent:   "Hello",
		LastModified: time.Now(),
	}
	cfg := model.Config{Build: model.BuildConfig{ContentDir: "content", OutputDir: "public"}}
	processed, err := p.Process([]*model.Article{a}, cfg)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	want := "posts/my-post.md"
	if processed[0].ContentPath != want {
		t.Errorf("ContentPath: got %q, want %q", processed[0].ContentPath, want)
	}
}
