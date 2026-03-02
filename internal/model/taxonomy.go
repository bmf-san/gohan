package model

// Taxonomy represents a single tag or category entry.
type Taxonomy struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// TaxonomyRegistry holds the master lists loaded from taxonomy YAML files.
type TaxonomyRegistry struct {
	Tags       []Taxonomy `yaml:"tags"`
	Categories []Taxonomy `yaml:"categories"`
}
