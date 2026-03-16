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
//	  - asin: "4781920004"           # Amazon ASIN — generates image + Amazon link
//	    title: "組織を変える5つの対話"   # optional; used for alt text
//	  - url: "https://booth.pm/..."  # non-Amazon: direct sales URL (no cover image)
//	    title: "同人誌タイトル"
//
// # Template usage (bookshelf.html)
//
//	{{range index .VirtualPageData "categories"}}
//	  <h2>{{if .Name}}{{.Name}}{{else}}Uncategorized{{end}}</h2>
//	  {{range .Books}}
//	    <a href="{{.LinkURL}}" target="_blank" rel="noopener">
//	      <img src="{{.ImageURL}}" alt="{{.Title}}">
//	    </a>
//	    {{if .ArticleURL}}<a href="{{.ArticleURL}}">{{.ArticleTitle}}</a>{{end}}
//	  {{end}}
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
	Categories   []string
}

// CategoryGroup groups BookEntries under a single category name.
// When Name is empty the books have no category assigned.
type CategoryGroup struct {
	Name  string
	Books []BookEntry
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
// Each page's Data map contains:
//
//	"books"      []BookEntry       — all entries sorted by date descending
//	"categories" []CategoryGroup   — entries grouped by article category, sorted alphabetically
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
			directURL := strVal(m, "url", "")
			if asin == "" && directURL == "" {
				continue
			}
			titleDef := asin
			if titleDef == "" {
				titleDef = directURL
			}
			title := strVal(m, "title", titleDef)
			var imageURL, linkURL string
			if asin != "" {
				imageURL = fmt.Sprintf(imageURLTemplate, asin)
				linkURL = fmt.Sprintf(linkURLTemplate, asin, url.QueryEscape(tag))
			} else {
				imageURL = ""
				linkURL = directURL
			}
			entry := BookEntry{
				ASIN:         asin,
				Title:        title,
				ImageURL:     imageURL,
				LinkURL:      linkURL,
				ArticleSlug:  article.FrontMatter.Slug,
				ArticleTitle: article.FrontMatter.Title,
				ArticleURL:   article.URL,
				Date:         article.FrontMatter.Date,
				Categories:   article.FrontMatter.Categories,
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
		// Sort all entries by date descending (newest first).
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
				"books":      le.entries,
				"categories": buildCategoryGroups(le.entries),
			},
		})
	}

	return pages, nil
}

// buildCategoryGroups groups entries by category and returns them sorted
// alphabetically by category name. Entries with no categories are placed
// in a group with an empty Name at the end.
func buildCategoryGroups(entries []BookEntry) []CategoryGroup {
	catMap := map[string][]BookEntry{}
	for _, e := range entries {
		if len(e.Categories) == 0 {
			catMap[""] = append(catMap[""], e)
		} else {
			for _, cat := range e.Categories {
				catMap[cat] = append(catMap[cat], e)
			}
		}
	}

	// Collect and sort names; empty (uncategorized) goes last.
	names := make([]string, 0, len(catMap))
	for name := range catMap {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if names[i] == "" {
			return false
		}
		if names[j] == "" {
			return true
		}
		return names[i] < names[j]
	})

	groups := make([]CategoryGroup, 0, len(names))
	for _, name := range names {
		books := catMap[name]
		// Books within each group are already date-sorted via le.entries order.
		groups = append(groups, CategoryGroup{Name: name, Books: books})
	}
	return groups
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
