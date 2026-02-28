package processor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/bmf-san/gohan/internal/model"
)

// LoadTaxonomyRegistry reads tags.yaml and categories.yaml from taxonomyDir
// and returns the combined TaxonomyRegistry.
// Missing files are treated as empty (no error).
func LoadTaxonomyRegistry(taxonomyDir string) (*model.TaxonomyRegistry, error) {
	reg := &model.TaxonomyRegistry{}

	tags, err := loadTaxonomyFile(filepath.Join(taxonomyDir, "tags.yaml"))
	if err != nil {
		return nil, fmt.Errorf("taxonomy: load tags: %w", err)
	}
	reg.Tags = tags

	cats, err := loadTaxonomyFile(filepath.Join(taxonomyDir, "categories.yaml"))
	if err != nil {
		return nil, fmt.Errorf("taxonomy: load categories: %w", err)
	}
	reg.Categories = cats

	return reg, nil
}

// loadTaxonomyFile reads a YAML file containing a list of Taxonomy entries.
// Returns nil slice if the file does not exist.
func loadTaxonomyFile(path string) ([]model.Taxonomy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var entries []model.Taxonomy
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return entries, nil
}

// ValidateArticleTaxonomies checks that every tag and category referenced in
// an article exists in the registry.  It returns one error per violation.
func ValidateArticleTaxonomies(articles []*model.ProcessedArticle, registry *model.TaxonomyRegistry) []error {
	tagSet := make(map[string]bool, len(registry.Tags))
	for _, t := range registry.Tags {
		tagSet[t.Name] = true
	}
	catSet := make(map[string]bool, len(registry.Categories))
	for _, c := range registry.Categories {
		catSet[c.Name] = true
	}

	var errs []error
	for _, a := range articles {
		for _, t := range a.FrontMatter.Tags {
			if !tagSet[t] {
				errs = append(errs, fmt.Errorf("article %q: unknown tag %q", a.FilePath, t))
			}
		}
		for _, c := range a.FrontMatter.Categories {
			if !catSet[c] {
				errs = append(errs, fmt.Errorf("article %q: unknown category %q", a.FilePath, c))
			}
		}
	}
	return errs
}

// BuildTagIndex returns a map from tag name to the articles that use that tag.
func BuildTagIndex(articles []*model.ProcessedArticle) map[string][]*model.ProcessedArticle {
	idx := make(map[string][]*model.ProcessedArticle)
	for _, a := range articles {
		for _, t := range a.FrontMatter.Tags {
			idx[t] = append(idx[t], a)
		}
	}
	return idx
}

// BuildCategoryIndex returns a map from category name to the articles that use it.
func BuildCategoryIndex(articles []*model.ProcessedArticle) map[string][]*model.ProcessedArticle {
	idx := make(map[string][]*model.ProcessedArticle)
	for _, a := range articles {
		for _, c := range a.FrontMatter.Categories {
			idx[c] = append(idx[c], a)
		}
	}
	return idx
}
