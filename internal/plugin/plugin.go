// Package plugin defines the gohan plugin interface and built-in plugin registry.
//
// Plugins are compiled into the gohan binary and enabled/disabled via config.yaml:
//
//	plugins:
//	  amazon_books:
//	    enabled: true
//	    tag: "your-associate-tag-22"
//
// Each plugin receives a ProcessedArticle and its own config section,
// and returns arbitrary data that the theme template can access via
// .PluginData.<plugin_name>.
//
// SitePlugins operate on the full site rather than individual articles.
// They generate VirtualPages — pages that have no corresponding Markdown
// source file (e.g. a bookshelf page aggregated from all articles).
package plugin

import "github.com/bmf-san/gohan/internal/model"

// Plugin is the interface that all gohan per-article plugins must implement.
type Plugin interface {
	// Name returns the unique identifier of the plugin.
	// This key is used in config.yaml under plugins.<name>
	// and as the key in ProcessedArticle.PluginData.
	Name() string

	// Enabled reports whether the plugin is active for the given config section.
	// cfg is the map under plugins.<name> in config.yaml.
	Enabled(cfg map[string]interface{}) bool

	// TemplateData returns data to inject into the template context for the
	// given article. The returned map is stored at .PluginData.<name> in
	// the template.
	TemplateData(article *model.ProcessedArticle, cfg map[string]interface{}) (map[string]interface{}, error)
}

// SitePlugin is the interface for plugins that operate on the full site and
// generate VirtualPages — pages with no corresponding Markdown source file.
//
// SitePlugins run after all articles have been processed and per-article
// plugins have been enriched. The VirtualPages they return are rendered by
// the HTML generator using the template named by VirtualPage.Template.
//
// Example uses: a bookshelf page aggregating book front-matter from all
// articles, a reading-list page, or any cross-article summary page.
type SitePlugin interface {
	// Name returns the unique identifier of the plugin.
	// This key is used in config.yaml under plugins.<name>.
	Name() string

	// Enabled reports whether the plugin is active for the given config section.
	Enabled(cfg map[string]interface{}) bool

	// VirtualPages inspects the full site and returns zero or more VirtualPages
	// to be rendered by the HTML generator.
	VirtualPages(site *model.Site, cfg map[string]interface{}) ([]*model.VirtualPage, error)
}
