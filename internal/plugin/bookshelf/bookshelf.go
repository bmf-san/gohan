// Package bookshelf is a gohan built-in SitePlugin that generates a bookshelf
// page by aggregating book entries from every article's front-matter.
//
// # Configuration (config.yaml)
//
//	plugins:
//	  bookshelf:
//	    enabled: true
//	    tag: "your-associate-tag-22"   # Amazon Associates tracking tag
//
// # Front-matter (article .md) — same key used by amazon_books plugin
//
//	books:
//	  - asin: "4781920004"
//	    title: "組織を変える5つの対話"   # optional; used for alt text
//
// # Template usage (bookshelf.html)
//
//	{{range index .VirtualPageData "books"}}
//	  <a href="{{.LinkURL}}" target="_blank" rel="noopener">
//	    <img src="{{.ImageURL}}" alt="{{.Title}}">
//	  </a>
//	  {{if .ArticleURL}}<a href="{{.ArticleURL}}">{{.ArticleTitle}}</a>{{end}}
//	{{end}}
package bookshelf

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

const (
	// Name is the plugin identifier used as the key in config.yaml and registry.
	Name = "bookshelf"

	imageURLTemplate = "https://images-na.ssl-images-amazon.com/images/P/%s.01._SL250_.jpg"
	linkURLTemplate  = "https://www.amazon.co.jp/dp/%s?tag=%s"
	defaultTag       = ""
)

// BookEntry is the data exposed to templates for a single book on the bookshelf.
type BookEntry struct {
	ASIN         string
	Title        string
	ImageURL     string
	LinkURL      string
	ArticleSlug  string
	ArticleTitle string
	ArticleURL   string
	Date         time.Time
}

// Bookshelf implements SitePlugin.
type Bookshelf struct{}

// New returns a new Bookshelf plugin instance.
func New() *Bookshelf { return &Bookshelf{} }

// compile-time interface check
var _ interface {
	Name() string
	Enabled(map[string]interface{}) bool
	VirtualPages(*model.Site, map[string]interface{}) ([]*model.VirtualPage, error)
} = (*Bookshelf)(nil)

// Name returns the plugin identifier.
func (b *Bookshelf) Name() string { return Name }

// Enabled returns true when the plugin config has enabled: true.
func (b *Bookshelf) Enabled(cfg map[string]interface{}) bool {
	v, ok := cfg["enabled"]
	if !ok {
		return false
	}
	bv, ok := v.(bool)
	return ok && bv
}

// VirtualPages collects all book entries from article front-matter and returns
// one VirtualPage per locale containing the aggregated bookshelf data.
func (b *Bookshelf) VirtualPages(site *model.Site, cfg map[string]interface{}) ([]*model.VirtualPage, error) {
	tag := strVal(cfg, "tag", defaultTag)

	// Group BookEntries by locale.
	type localeEntries struct {
		entries []BookEntry
	}
	byLocale := map[string]*localeEntries{}

	for _, article := range site.Articles {
		raw, ok := article.FrontMatter.Extra["books"]
		if !ok || raw == nil {
			continue
		}
		items, ok := raw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("bookshelf: 'books' in article %q must be a YAML sequence", article.FrontMatter.Title)
		}

		locale := article.Locale

		if _, exists := byLocale[locale]; !exists {
			byLocale[locale] = &localeEntries{}
		}

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
			entry := BookEntry{
				ASIN:         asin,
				Title:        title,
				ImageURL:     fmt.Sprintf(imageURLTemplate, asin),
				LinkURL:      fmt.Sprintf(linkURLTemplate, asin, url.QueryEscape(tag)),
				ArticleSlug:  article.FrontMatter.Slug,
				ArticleTitle: article.FrontMatter.Title,
				ArticleURL:   article.URL,
				Date:         article.FrontMatter.Date,
			}
			byLocale[locale].entries = append(byLocale[locale].entries, entry)
		}
	}

	if len(byLocale) == 0 {
		return nil, nil
	}

	defaultLocale := site.Config.I18n.DefaultLocale
	if defaultLocale == "" {
		defaultLocale = site.Config.Site.Language
	}

	var pages []*model.VirtualPage
	for locale, le := range byLocale {
		// Sort by date descending (newest first).
		sort.Slice(le.entries, func(i, j int) bool {
			return le.entries[i].Date.After(le.entries[j].Date)
		})

		var outputPath, pageURL string
		if locale == defaultLocale || locale == "" {
			outputPath = path.Join("bookshelf", "index.html")
			pageURL = "/bookshelf/"
		} else {
			outputPath = path.Join(locale, "bookshelf", "index.html")
			pageURL = "/" + locale + "/bookshelf/"
		}

		pages = append(pages, &model.VirtualPage{
			OutputPath: outputPath,
			URL:        pageURL,
			Template:   "bookshelf.html",
			Locale:     locale,
			Data: map[string]interface{}{
				"books": le.entries,
			},
		})
	}

	return pages, nil
}

// strVal extracts a string value from a map, returning def when missing or wrong type.
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
