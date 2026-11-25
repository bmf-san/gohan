package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/plugin"
)

func TestRegistry_Enrich_DisabledPlugin(t *testing.T) {
	site := &model.Site{
		Config: model.Config{
			Plugins: map[string]interface{}{
				"amazon_books": map[string]interface{}{"enabled": false},
			},
		},
		Articles: []*model.ProcessedArticle{
			{Article: model.Article{FrontMatter: model.FrontMatter{Title: "A"}}},
		},
	}

	if err := plugin.DefaultRegistry().Enrich(site); err != nil {
		t.Fatalf("Enrich error: %v", err)
	}

	// PluginData should be nil / empty — disabled plugin must not populate it
	if site.Articles[0].PluginData != nil {
		if _, ok := site.Articles[0].PluginData["amazon_books"]; ok {
			t.Error("disabled plugin should not populate PluginData")
		}
	}
}

func TestRegistry_Enrich_EnabledPlugin(t *testing.T) {
	site := &model.Site{
		Config: model.Config{
			Plugins: map[string]interface{}{
				"amazon_books": map[string]interface{}{
					"enabled": true,
					"tag":     "test-22",
				},
			},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{
					FrontMatter: model.FrontMatter{
						Title: "A",
						Extra: map[string]interface{}{
							"books": []interface{}{
								map[string]interface{}{"asin": "4873119464", "title": "入門"},
							},
						},
					},
				},
			},
		},
	}

	if err := plugin.DefaultRegistry().Enrich(site); err != nil {
		t.Fatalf("Enrich error: %v", err)
	}

	pd := site.Articles[0].PluginData
	if pd == nil {
		t.Fatal("PluginData is nil")
	}
	if _, ok := pd["amazon_books"]; !ok {
		t.Error("expected amazon_books key in PluginData")
	}
}

func TestRegistry_Enrich_NoPluginsConfig(t *testing.T) {
	// nil Plugins map should not panic
	site := &model.Site{
		Config:   model.Config{},
		Articles: []*model.ProcessedArticle{{Article: model.Article{FrontMatter: model.FrontMatter{Title: "X"}}}},
	}
	if err := plugin.DefaultRegistry().Enrich(site); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistry_GenerateAssets_EnabledOGP(t *testing.T) {
	outDir := t.TempDir()
	site := &model.Site{
		Config: model.Config{
			Plugins: map[string]interface{}{
				"ogp": map[string]interface{}{
					"enabled": true,
					"width":   120,
					"height":  63,
				},
			},
		},
		Articles: []*model.ProcessedArticle{
			{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Hello", Slug: "hello"}}},
		},
	}

	if err := plugin.DefaultRegistry().GenerateAssets(site, outDir, nil); err != nil {
		t.Fatalf("GenerateAssets error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ogp", "hello.png")); err != nil {
		t.Errorf("expected ogp/hello.png to be generated: %v", err)
	}
}

func TestRegistry_GenerateAssets_DisabledOGP(t *testing.T) {
	outDir := t.TempDir()
	site := &model.Site{
		Config: model.Config{
			Plugins: map[string]interface{}{
				"ogp": map[string]interface{}{"enabled": false},
			},
		},
		Articles: []*model.ProcessedArticle{
			{Article: model.Article{FrontMatter: model.FrontMatter{Title: "Hello", Slug: "hello"}}},
		},
	}

	if err := plugin.DefaultRegistry().GenerateAssets(site, outDir, nil); err != nil {
		t.Fatalf("GenerateAssets error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ogp")); !os.IsNotExist(err) {
		t.Errorf("disabled ogp plugin should not create the ogp dir (err=%v)", err)
	}
}

func TestRegistry_GenerateAssets_NoPluginsConfig(t *testing.T) {
	// nil Plugins map should not panic and should produce nothing.
	outDir := t.TempDir()
	site := &model.Site{
		Config:   model.Config{},
		Articles: []*model.ProcessedArticle{{Article: model.Article{FrontMatter: model.FrontMatter{Title: "X", Slug: "x"}}}},
	}
	if err := plugin.DefaultRegistry().GenerateAssets(site, outDir, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func glossarySite(enabled bool) *model.Site {
	return &model.Site{
		Config: model.Config{
			Site: model.SiteConfig{Language: "en"},
			Plugins: map[string]interface{}{
				"glossary": map[string]interface{}{"enabled": enabled},
			},
		},
		Articles: []*model.ProcessedArticle{
			{
				Article: model.Article{FrontMatter: model.FrontMatter{
					Title: "Authorization",
					Extra: map[string]interface{}{"glossary": []interface{}{
						map[string]interface{}{"term": "PEP", "definition": "Policy enforcement point."},
					}},
				}},
				URL: "/posts/authz/",
			},
		},
	}
}

func TestRegistry_EnrichVirtual_Glossary(t *testing.T) {
	site := glossarySite(true)
	if err := plugin.DefaultRegistry().EnrichVirtual(site); err != nil {
		t.Fatalf("EnrichVirtual error: %v", err)
	}

	var found bool
	for _, vp := range site.VirtualPages {
		if vp.Template == "glossary.html" && vp.URL == "/glossary/" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a glossary virtual page at /glossary/, got %#v", site.VirtualPages)
	}
	if _, ok := site.SiteData["glossary"]; !ok {
		t.Error("expected SiteData[\"glossary\"] to be set")
	}
}

func TestRegistry_EnrichVirtual_GlossaryDisabled(t *testing.T) {
	site := glossarySite(false)
	if err := plugin.DefaultRegistry().EnrichVirtual(site); err != nil {
		t.Fatalf("EnrichVirtual error: %v", err)
	}
	for _, vp := range site.VirtualPages {
		if vp.Template == "glossary.html" {
			t.Errorf("disabled glossary plugin must not generate pages, got %q", vp.URL)
		}
	}
	if _, ok := site.SiteData["glossary"]; ok {
		t.Error("disabled glossary plugin must not set SiteData")
	}
}
