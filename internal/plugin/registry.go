package plugin

import (
	"fmt"

	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/plugin/amazonbooks"
	"github.com/bmf-san/gohan/internal/plugin/bookshelf"
	"github.com/bmf-san/gohan/internal/plugin/ogp"
)

// Registry holds the set of built-in plugins.
type Registry struct {
	plugins      []Plugin
	sitePlugins  []SitePlugin
	assetPlugins []AssetPlugin
}

// DefaultRegistry returns a Registry pre-loaded with all built-in plugins.
func DefaultRegistry() *Registry {
	return &Registry{
		plugins: []Plugin{
			amazonbooks.New(),
		},
		sitePlugins: []SitePlugin{
			bookshelf.New(),
		},
		assetPlugins: []AssetPlugin{
			ogp.New(),
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

// EnrichVirtual runs all enabled SitePlugins over the full site, collecting
// VirtualPages and appending them to site.VirtualPages.
// Call this after Enrich and before generating HTML.
func (r *Registry) EnrichVirtual(site *model.Site) error {
	pluginsCfg := site.Config.Plugins

	for _, sp := range r.sitePlugins {
		cfg := pluginCfg(pluginsCfg, sp.Name())
		if !sp.Enabled(cfg) {
			continue
		}
		pages, err := sp.VirtualPages(site, cfg)
		if err != nil {
			return fmt.Errorf("site plugin %s: %w", sp.Name(), err)
		}
		site.VirtualPages = append(site.VirtualPages, pages...)

		// Optionally collect global, cross-page data from the plugin.
		if sdp, ok := sp.(SiteDataProvider); ok {
			data, err := sdp.SiteData(site, cfg)
			if err != nil {
				return fmt.Errorf("site plugin %s site data: %w", sp.Name(), err)
			}
			if data != nil {
				if site.SiteData == nil {
					site.SiteData = make(map[string]interface{})
				}
				site.SiteData[sp.Name()] = data
			}
		}
	}
	return nil
}

// GenerateAssets runs all enabled AssetPlugins, writing build-time artifacts
// (e.g. OGP images) into outDir. changeSet enables incremental skipping.
// Call this after the site has been processed (typically after HTML generation).
func (r *Registry) GenerateAssets(site *model.Site, outDir string, changeSet *model.ChangeSet) error {
	pluginsCfg := site.Config.Plugins

	for _, ap := range r.assetPlugins {
		cfg := pluginCfg(pluginsCfg, ap.Name())
		if !ap.Enabled(cfg) {
			continue
		}
		if err := ap.GenerateAssets(site, outDir, changeSet, cfg); err != nil {
			return fmt.Errorf("asset plugin %s: %w", ap.Name(), err)
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
