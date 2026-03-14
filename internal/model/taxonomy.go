package model

// Taxonomy represents a single tag or category entry.
type Taxonomy struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	URL         string `yaml:"-"` // set at render time; locale-aware canonical URL
}

// TaxonomyRegistry holds the master lists loaded from taxonomy YAML files.
type TaxonomyRegistry struct {
	Tags       []Taxonomy `yaml:"tags"`
	Categories []Taxonomy `yaml:"categories"`
}
