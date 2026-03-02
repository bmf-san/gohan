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
package plugin

import "github.com/bmf-san/gohan/internal/model"

// Plugin is the interface that all gohan plugins must implement.
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
