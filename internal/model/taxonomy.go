package model

// Taxonomy represents a single tag or category entry.
type Taxonomy struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	// TranslationKey links taxonomies that represent the same concept across
	// locales (e.g. EN "Application" and JA "アプリケーション" both set
	// translation_key: application). Empty means "no cross-locale binding".
	TranslationKey string `yaml:"translation_key"`
	URL            string `yaml:"-"` // set at render time; locale-aware canonical URL
	// Translations maps locale → URL for every other locale's taxonomy that
	// shares the same TranslationKey. Populated at render time on
	// CurrentTaxonomy; nil elsewhere. Template access:
	//   {{index .CurrentTaxonomy.Translations "en"}}
	Translations map[string]string `yaml:"-"`
}

// TaxonomyRegistry holds the master lists loaded from taxonomy YAML files.
type TaxonomyRegistry struct {
	Tags       []Taxonomy `yaml:"tags"`
	Categories []Taxonomy `yaml:"categories"`
}
