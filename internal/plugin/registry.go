package plugin

import (
	"fmt"

	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/plugin/amazonbooks"
)

// Registry holds the set of built-in plugins.
type Registry struct {
	plugins []Plugin
}

// DefaultRegistry returns a Registry pre-loaded with all built-in plugins.
func DefaultRegistry() *Registry {
	return &Registry{
		plugins: []Plugin{
			amazonbooks.New(),
		},
	}
}

// Enrich runs all enabled plugins over every article in site,
// populating article.PluginData with each plugin's output.
// Call this after processing articles and before generating HTML.
func (r *Registry) Enrich(site *model.Site) error {
	pluginsCfg := site.Config.Plugins

	for _, p := range r.plugins {
		cfg := pluginCfg(pluginsCfg, p.Name())
		if !p.Enabled(cfg) {
			continue
		}
		for _, article := range site.Articles {
			data, err := p.TemplateData(article, cfg)
			if err != nil {
				return fmt.Errorf("plugin %s: article %q: %w", p.Name(), article.FrontMatter.Title, err)
			}
			if article.PluginData == nil {
				article.PluginData = make(map[string]interface{})
			}
			article.PluginData[p.Name()] = data
		}
	}
	return nil
}

// pluginCfg extracts the config sub-map for the named plugin.
// Returns an empty map when not set.
func pluginCfg(all map[string]interface{}, name string) map[string]interface{} {
	if all == nil {
		return map[string]interface{}{}
	}
	v, ok := all[name]
	if !ok {
		return map[string]interface{}{}
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	return m
}
