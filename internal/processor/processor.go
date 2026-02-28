// Package processor builds dependency graphs and taxonomies from parsed articles.
package processor

import (
	"github.com/bmf-san/gohan/internal/model"
)

// Processor enriches raw Article data and builds the site-wide dependency graph
// and taxonomy registry that are required for incremental builds.
type Processor interface {
	// Process converts a slice of Articles into ProcessedArticles, rendering
	// Markdown to HTML, extracting summaries, resolving output paths, etc.
	Process(articles []*model.Article, cfg model.Config) ([]*model.ProcessedArticle, error)

	// BuildDependencyGraph constructs the full DependencyGraph from the
	// complete set of ProcessedArticles.
	BuildDependencyGraph(articles []*model.ProcessedArticle) (*model.DependencyGraph, error)

	// BuildTaxonomyRegistry collects all tags and categories referenced across
	// all articles and validates them against the configured taxonomy YAML files.
	BuildTaxonomyRegistry(articles []*model.ProcessedArticle, cfg model.Config) (*model.TaxonomyRegistry, error)
}
