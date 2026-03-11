// Package amazonbooks is a gohan built-in plugin that generates Amazon book cards
// for article templates.
//
// # Configuration (config.yaml)
//
//	plugins:
//	  amazon_books:
//	    enabled: true
//	    tag: "your-associate-tag-22"   # Amazon Associates tracking tag
//
// # Front-matter (article .md)
//
//	books:
//	  - asin: "4781920004"
//	    title: "組織を変える5つの対話"   # optional; used for alt text
//	  - asin: "4873119464"
//
// # Template usage
//
//	{{with index .PluginData "amazon_books"}}
//	  {{range .Books}}
//	    <a href="{{.LinkURL}}" target="_blank" rel="noopener">
//	      <img src="{{.ImageURL}}" alt="{{.Title}}">
//	    </a>
//	  {{end}}
//	{{end}}
package amazonbooks

import (
	"fmt"
	"net/url"

	"github.com/bmf-san/gohan/internal/model"
)

const (
	// Name is the plugin identifier, used as the key in config.yaml and PluginData.
	Name = "amazon_books"

	imageURLTemplate = "https://images-na.ssl-images-amazon.com/images/P/%s.01._SL250_.jpg"
	linkURLTemplate  = "https://www.amazon.co.jp/dp/%s?tag=%s"
	defaultTag       = ""
)

// BookCard is the data exposed to templates for a single book.
type BookCard struct {
	ASIN     string
	Title    string
	ImageURL string
	LinkURL  string
}

// AmazonBooks implements plugin.Plugin.
type AmazonBooks struct{}

// New returns a new AmazonBooks plugin instance.
func New() *AmazonBooks { return &AmazonBooks{} }

// compile-time interface check
var _ interface {
	Name() string
	Enabled(map[string]interface{}) bool
	TemplateData(*model.ProcessedArticle, map[string]interface{}) (map[string]interface{}, error)
} = (*AmazonBooks)(nil)

// Name returns the plugin identifier.
func (a *AmazonBooks) Name() string { return Name }

// Enabled returns true when the plugin config has enabled: true.
func (a *AmazonBooks) Enabled(cfg map[string]interface{}) bool {
	v, ok := cfg["enabled"]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// TemplateData builds BookCard slice from the article's front-matter `books` key.
// Returns an empty map (not an error) when the article has no books.
func (a *AmazonBooks) TemplateData(article *model.ProcessedArticle, cfg map[string]interface{}) (map[string]interface{}, error) {
	tag := strVal(cfg, "tag", defaultTag)

	raw, ok := article.FrontMatter.Extra["books"]
	if !ok || raw == nil {
		return map[string]interface{}{"books": []BookCard{}}, nil
	}

	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("amazon_books: 'books' must be a YAML sequence")
	}

	cards := make([]BookCard, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		asin := strVal(m, "asin", "")
		if asin == "" {
			continue
		}
		title := strVal(m, "title", asin)
		cards = append(cards, BookCard{
			ASIN:     asin,
			Title:    title,
			ImageURL: fmt.Sprintf(imageURLTemplate, asin),
			LinkURL:  fmt.Sprintf(linkURLTemplate, asin, url.QueryEscape(tag)),
		})
	}

	return map[string]interface{}{"books": cards}, nil
}

// strVal reads a string value from m[key], returning def if absent or wrong type.
func strVal(m map[string]interface{}, key, def string) string {
	v, ok := m[key]
	if !ok {
		return def
	}
	s, ok := v.(string)
	if !ok {
		return def
	}
	return s
}
