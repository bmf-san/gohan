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

// computeOutputPath determines the output HTML path for an article.
// Respects FrontMatter.Slug when set; otherwise uses the file base name.
func computeOutputPath(a *model.Article, cfg model.Config) string {
	rel, err := filepath.Rel(cfg.Build.ContentDir, a.FilePath)
	if err != nil {
		rel = filepath.Base(a.FilePath)
	}
	dir := filepath.Dir(rel)
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	if a.FrontMatter.Slug != "" {
		base = a.FrontMatter.Slug
	}
	return filepath.Join(cfg.Build.OutputDir, dir, base, "index.html")
}

// extractSummary returns the first paragraph of content, truncated to maxChars.
func extractSummary(content string, maxChars int) string {
	content = strings.TrimSpace(content)
	if idx := strings.Index(content, "\n\n"); idx > 0 && idx <= maxChars {
		return strings.TrimSpace(content[:idx])
	}
	if len(content) <= maxChars {
		return content
	}
	return content[:maxChars] + "..."
}

// Ensure SiteProcessor implements the Processor interface at compile time.
var _ Processor = (*SiteProcessor)(nil)
