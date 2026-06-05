// Package bookshelf is a gohan built-in SitePlugin that generates a bookshelf
// page by aggregating book entries from every article's front-matter.
//
// # Configuration (config.yaml)
//
//	plugins:
//	  bookshelf:
//	    enabled: true
//	    tag: "your-associate-tag-22"   # Amazon Associates tracking tag
//	    recent_limit: 5                # optional; books exposed via SiteData (default 5)
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
//
// # Recently-read widget (any template, e.g. homepage sidebar)
//
// In addition to the bookshelf VirtualPage, the plugin exposes the newest
// books per locale on every page via Site.SiteData (see SiteDataProvider):
//
//	{{$bs := index .SiteData "bookshelf"}}
//	{{if $bs}}{{$recent := index (index $bs "recent") .CurrentLocale}}
//	  {{range $recent}}
//	    <a href="{{.LinkURL}}" rel="sponsored noopener noreferrer">
//	      <img src="{{.ImageURL}}" alt="{{.Title}}">
//	    </a>
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

	// defaultRecentLimit is the number of newest books exposed per locale via
	// SiteData when `recent_limit` is not set in config.
	defaultRecentLimit = 5
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
	SiteData(*model.Site, map[string]interface{}) (interface{}, error)
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

	byLocale, err := collectEntriesByLocale(site, tag)
	if err != nil {
		return nil, err
	}
	if len(byLocale) == 0 {
		return nil, nil
	}

	defaultLocale := site.Config.I18n.DefaultLocale
	if defaultLocale == "" {
		defaultLocale = site.Config.Site.Language
	}

	var pages []*model.VirtualPage
	for locale, entries := range byLocale {
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
				"books":      entries,
				"categories": buildCategoryGroups(entries),
			},
		})
	}

	return pages, nil
}

// SiteData implements plugin.SiteDataProvider. It exposes the most recently
// read books per locale on every rendered page, so themes can show a "recently
// read" widget outside the bookshelf page (e.g. the homepage sidebar).
//
// The returned value has the shape:
//
//	{ "recent": { "<locale>": []BookEntry } }
//
// Each locale's slice is sorted newest-first and capped at `recent_limit`
// books (config key under plugins.bookshelf; default defaultRecentLimit).
// Returns nil when no article declares books.
func (b *Bookshelf) SiteData(site *model.Site, cfg map[string]interface{}) (interface{}, error) {
	tag := strVal(cfg, "tag", defaultTag)
	limit := intVal(cfg, "recent_limit", defaultRecentLimit)

	byLocale, err := collectEntriesByLocale(site, tag)
	if err != nil {
		return nil, err
	}
	if len(byLocale) == 0 {
		return nil, nil
	}

	recent := make(map[string]interface{}, len(byLocale))
	for locale, entries := range byLocale {
		if limit > 0 && len(entries) > limit {
			entries = entries[:limit]
		}
		recent[locale] = entries
	}

	return map[string]interface{}{"recent": recent}, nil
}

// collectEntriesByLocale aggregates book entries from every article's
// front-matter, grouped by locale and sorted by date descending (newest
// first). Returns an empty map when no article declares books.
func collectEntriesByLocale(site *model.Site, tag string) (map[string][]BookEntry, error) {
	byLocale := map[string][]BookEntry{}

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
			byLocale[locale] = append(byLocale[locale], BookEntry{
				ASIN:         asin,
				Title:        title,
				ImageURL:     imageURL,
				LinkURL:      linkURL,
				ArticleSlug:  article.FrontMatter.Slug,
				ArticleTitle: article.FrontMatter.Title,
				ArticleURL:   article.URL,
				Date:         article.FrontMatter.Date,
				Categories:   article.FrontMatter.Categories,
			})
		}
	}

	// Sort each locale's entries by date descending (newest first).
	for locale := range byLocale {
		entries := byLocale[locale]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Date.After(entries[j].Date)
		})
		byLocale[locale] = entries
	}

	return byLocale, nil
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

// intVal extracts an int value from a map, returning def when missing or of an
// unexpected type. YAML may decode integers as int, int64, or float64.
func intVal(m map[string]interface{}, key string, def int) int {
	v, ok := m[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return def
	}
}
