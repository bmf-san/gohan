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

// loadTaxonomyFileWithFallback returns entries from primary if the file exists,
// otherwise falls back to fallback. A missing primary is not an error.
func loadTaxonomyFileWithFallback(primary, fallback string) ([]model.Taxonomy, error) {
	if _, err := os.Stat(primary); err == nil {
		return loadTaxonomyFile(primary)
	}
	return loadTaxonomyFile(fallback)
}

// LoadLocaleAwareTaxonomyRegistries loads a taxonomy registry per locale.
// For each locale it looks for {contentDir}/{locale}/tags.yaml (and categories.yaml),
// falling back to the global {contentDir}/tags.yaml when the locale file is absent.
// The returned map also has an "" (empty string) key holding the global registry.
func LoadLocaleAwareTaxonomyRegistries(contentDir string, locales []string) (map[string]*model.TaxonomyRegistry, error) {
	registries := make(map[string]*model.TaxonomyRegistry, len(locales)+1)

	global, err := LoadTaxonomyRegistry(contentDir)
	if err != nil {
		return nil, err
	}
	registries[""] = global

	for _, locale := range locales {
		tags, err := loadTaxonomyFileWithFallback(
			filepath.Join(contentDir, locale, "tags.yaml"),
			filepath.Join(contentDir, "tags.yaml"),
		)
		if err != nil {
			return nil, fmt.Errorf("taxonomy: locale %q tags: %w", locale, err)
		}
		cats, err := loadTaxonomyFileWithFallback(
			filepath.Join(contentDir, locale, "categories.yaml"),
			filepath.Join(contentDir, "categories.yaml"),
		)
		if err != nil {
			return nil, fmt.Errorf("taxonomy: locale %q categories: %w", locale, err)
		}
		registries[locale] = &model.TaxonomyRegistry{Tags: tags, Categories: cats}
	}
	return registries, nil
}

// MergeTaxonomyRegistries returns a single registry that is the deduplicated
// union of all registries in the map. Useful for populating site.Tags/Categories.
func MergeTaxonomyRegistries(registries map[string]*model.TaxonomyRegistry) *model.TaxonomyRegistry {
	tagSeen := make(map[string]bool)
	catSeen := make(map[string]bool)
	merged := &model.TaxonomyRegistry{}
	for _, reg := range registries {
		for _, t := range reg.Tags {
			if !tagSeen[t.Name] {
				tagSeen[t.Name] = true
				merged.Tags = append(merged.Tags, t)
			}
		}
		for _, c := range reg.Categories {
			if !catSeen[c.Name] {
				catSeen[c.Name] = true
				merged.Categories = append(merged.Categories, c)
			}
		}
	}
	return merged
}

// ValidateArticleTaxonomiesLocale validates each article against its locale's
// registry from the registries map. Falls back to the "" key when no
// locale-specific registry is found. Only validates when the registry has entries.
func ValidateArticleTaxonomiesLocale(articles []*model.ProcessedArticle, registries map[string]*model.TaxonomyRegistry) []error {
	type setsPair struct {
		tags map[string]bool
		cats map[string]bool
	}
	sets := make(map[string]setsPair, len(registries))
	for locale, reg := range registries {
		ts := make(map[string]bool, len(reg.Tags))
		for _, t := range reg.Tags {
			ts[t.Name] = true
		}
		cs := make(map[string]bool, len(reg.Categories))
		for _, c := range reg.Categories {
			cs[c.Name] = true
		}
		sets[locale] = setsPair{ts, cs}
	}

	var errs []error
	for _, a := range articles {
		sp, ok := sets[a.Locale]
		if !ok {
			sp = sets[""]
		}
		if len(sp.tags) > 0 {
			for _, t := range a.FrontMatter.Tags {
				if !sp.tags[t] {
					errs = append(errs, fmt.Errorf("article %q: unknown tag %q", a.FilePath, t))
				}
			}
		}
		if len(sp.cats) > 0 {
			for _, c := range a.FrontMatter.Categories {
				if !sp.cats[c] {
					errs = append(errs, fmt.Errorf("article %q: unknown category %q", a.FilePath, c))
				}
			}
		}
	}
	return errs
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

// ValidateOutputPaths checks that no two articles resolve to the same output
// path, which would cause one page to silently overwrite the other during
// HTML generation.  It returns one error per duplicate pair.
func ValidateOutputPaths(articles []*model.ProcessedArticle) []error {
	seen := make(map[string]string, len(articles)) // OutputPath -> FilePath
	var errs []error
	for _, a := range articles {
		if prev, ok := seen[a.OutputPath]; ok {
			errs = append(errs, fmt.Errorf("duplicate output path %q: %q and %q", a.OutputPath, prev, a.FilePath))
		} else {
			seen[a.OutputPath] = a.FilePath
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
