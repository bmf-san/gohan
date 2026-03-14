package processor

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmf-san/gohan/internal/highlight"
	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/parser"
)

// SiteProcessor implements the Processor interface.
type SiteProcessor struct{}

// NewSiteProcessor returns a new SiteProcessor.
func NewSiteProcessor() *SiteProcessor {
	return &SiteProcessor{}
}

// Process converts raw Articles into ProcessedArticles by rendering Markdown
// to HTML, extracting summaries, and computing output paths.
func (p *SiteProcessor) Process(articles []*model.Article, cfg model.Config) ([]*model.ProcessedArticle, error) {
	hlCfg := highlight.Config{
		Theme:       cfg.SyntaxHighlight.Theme,
		LineNumbers: cfg.SyntaxHighlight.LineNumbers,
	}
	convOpts := []parser.ConverterOption{parser.WithGFM(), parser.WithMermaid()}
	if hlCfg.Theme != "" {
		convOpts = append(convOpts, parser.WithHighlighting(hlCfg))
	}
	conv := parser.NewConverter(convOpts...)
	result := make([]*model.ProcessedArticle, 0, len(articles))
	for _, a := range articles {
		html, err := conv.Convert([]byte(a.RawContent))
		if err != nil {
			return nil, fmt.Errorf("processor: render %s: %w", a.FilePath, err)
		}
		processed := &model.ProcessedArticle{
			Article:     *a,
			HTMLContent: html,
			Summary:     extractSummary(a.RawContent, 200),
			OutputPath:  computeOutputPath(a, cfg),
			ContentPath: computeContentPath(a, cfg),
			Locale:      detectLocale(a, cfg),
			URL:         computeArticleURL(a, cfg),
		}
		result = append(result, processed)
	}
	return result, nil
}

// BuildDependencyGraph constructs a DependencyGraph from all processed articles,
// linking each article to its tag, category, and archive (year) nodes.
func (p *SiteProcessor) BuildDependencyGraph(articles []*model.ProcessedArticle) (*model.DependencyGraph, error) {
	g := &model.DependencyGraph{
		Nodes: make(map[string]*model.Node),
		Edges: make(map[string][]string),
	}
	for _, a := range articles {
		articlePath := a.FilePath
		addNode(g, &model.Node{
			Path:         articlePath,
			Type:         model.NodeTypeArticle,
			LastModified: a.LastModified,
		})
		for _, tag := range a.FrontMatter.Tags {
			tagPath := "tag:" + tag
			addNode(g, &model.Node{Path: tagPath, Type: model.NodeTypeTag, LastModified: time.Time{}})
			addEdge(g, articlePath, tagPath)
		}
		for _, cat := range a.FrontMatter.Categories {
			catPath := "category:" + cat
			addNode(g, &model.Node{Path: catPath, Type: model.NodeTypeCategory, LastModified: time.Time{}})
			addEdge(g, articlePath, catPath)
		}
		if !a.FrontMatter.Date.IsZero() {
			year := fmt.Sprintf("archive:%d", a.FrontMatter.Date.Year())
			addNode(g, &model.Node{Path: year, Type: model.NodeTypeArchive, LastModified: time.Time{}})
			addEdge(g, articlePath, year)
		}
	}
	return g, nil
}

// BuildTaxonomyRegistry collects all unique tags and categories referenced
// across the article set and returns a TaxonomyRegistry.
func (p *SiteProcessor) BuildTaxonomyRegistry(articles []*model.ProcessedArticle, cfg model.Config) (*model.TaxonomyRegistry, error) {
	tagSeen := make(map[string]bool)
	catSeen := make(map[string]bool)
	reg := &model.TaxonomyRegistry{}
	for _, a := range articles {
		for _, t := range a.FrontMatter.Tags {
			if !tagSeen[t] {
				tagSeen[t] = true
				reg.Tags = append(reg.Tags, model.Taxonomy{Name: t})
			}
		}
		for _, c := range a.FrontMatter.Categories {
			if !catSeen[c] {
				catSeen[c] = true
				reg.Categories = append(reg.Categories, model.Taxonomy{Name: c})
			}
		}
	}
	return reg, nil
}

// computeContentPath returns the content-dir-relative path to the source file
// (e.g. "posts/hello-world.md"), using forward slashes for URL compatibility.
func computeContentPath(a *model.Article, cfg model.Config) string {
	rel, err := filepath.Rel(cfg.Build.ContentDir, a.FilePath)
	if err != nil {
		return filepath.Base(a.FilePath)
	}
	// Normalise to forward slashes so the value is safe to embed in a URL.
	return filepath.ToSlash(rel)
}

// detectLocale returns the locale code for the article by matching the first
// path segment after the content directory against cfg.I18n.Locales.
// Returns an empty string when i18n is not configured or no match is found.
func detectLocale(a *model.Article, cfg model.Config) string {
	if len(cfg.I18n.Locales) == 0 {
		return ""
	}
	rel, err := filepath.Rel(cfg.Build.ContentDir, a.FilePath)
	if err != nil {
		return ""
	}
	parts := strings.SplitN(filepath.ToSlash(rel), "/", 2)
	if len(parts) == 0 {
		return ""
	}
	for _, loc := range cfg.I18n.Locales {
		if parts[0] == loc {
			return loc
		}
	}
	return ""
}

// computeArticleURL returns the canonical URL path for an article
// (e.g. "/posts/hello/" or "/ja/posts/hello/").
// Returns an empty string when i18n is not configured.
func computeArticleURL(a *model.Article, cfg model.Config) string {
	if len(cfg.I18n.Locales) == 0 {
		return ""
	}
	outPath := computeOutputPath(a, cfg)
	rel, err := filepath.Rel(cfg.Build.OutputDir, outPath)
	if err != nil {
		return ""
	}
	// "posts/hello/index.html" → dir="posts/hello" → "/posts/hello/"
	// "ja/posts/hello/index.html" → dir="ja/posts/hello" → "/ja/posts/hello/"
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "." {
		return "/"
	}
	return "/" + dir + "/"
}

// BuildTranslationMap populates the Translations field of each ProcessedArticle
// that has a TranslationKey set, linking it to sibling articles in other locales.
// Call this once after all articles have been processed.
func (p *SiteProcessor) BuildTranslationMap(articles []*model.ProcessedArticle) {
	byKey := make(map[string][]*model.ProcessedArticle)
	for _, a := range articles {
		if a.FrontMatter.TranslationKey != "" {
			byKey[a.FrontMatter.TranslationKey] = append(byKey[a.FrontMatter.TranslationKey], a)
		}
	}
	for _, a := range articles {
		key := a.FrontMatter.TranslationKey
		if key == "" {
			continue
		}
		for _, sibling := range byKey[key] {
			if sibling == a {
				continue
			}
			// Skip siblings with no locale or URL (non-i18n sites where
			// translation_key is used without an i18n configuration).
			if sibling.Locale == "" || sibling.URL == "" {
				continue
			}
			a.Translations = append(a.Translations, model.LocaleRef{
				Locale: sibling.Locale,
				URL:    sibling.URL,
			})
		}
	}
}

// computeOutputPath determines the output HTML path for an article.
// Respects FrontMatter.Slug when set; otherwise uses the file base name.
// When i18n is active, strips the locale segment from the content path and
// re-adds it as a URL prefix for non-default locales.
func computeOutputPath(a *model.Article, cfg model.Config) string {
	rel, err := filepath.Rel(cfg.Build.ContentDir, a.FilePath)
	if err != nil {
		rel = filepath.Base(a.FilePath)
	}
	dir := filepath.Dir(rel)
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	if a.FrontMatter.Slug != "" {
		// BUG-A: sanitise slug against path traversal — take only the last path
		// component so that "../../etc/passwd" reduces to "passwd".
		s := filepath.Base(filepath.FromSlash(a.FrontMatter.Slug))
		if s == "." || s == ".." {
			// degenerate input — keep the filename-derived base
		} else {
			base = s
		}
	}
	// i18n: strip the locale segment from dir, re-add only for non-default locales.
	if locale := detectLocale(a, cfg); locale != "" {
		parts := strings.SplitN(filepath.ToSlash(dir), "/", 2)
		if len(parts) == 1 {
			dir = "."
		} else {
			dir = filepath.FromSlash(parts[1])
		}
		if locale != cfg.I18n.DefaultLocale {
			dir = filepath.Join(locale, dir)
		}
	}
	return filepath.Join(cfg.Build.OutputDir, dir, base, "index.html")
}

// extractSummary returns the first paragraph of content, truncated to maxChars runes.
func extractSummary(content string, maxChars int) string {
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if idx := strings.Index(content, "\n\n"); idx > 0 {
		paragraph := strings.TrimSpace(content[:idx])
		if len([]rune(paragraph)) <= maxChars {
			return paragraph
		}
	}
	if len(runes) <= maxChars {
		return content
	}
	return string(runes[:maxChars]) + "..."
}

// Ensure SiteProcessor implements the Processor interface at compile time.
var _ Processor = (*SiteProcessor)(nil)
