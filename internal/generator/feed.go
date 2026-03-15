package generator

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

// ---- RSS 2.0 structs -------------------------------------------------------

type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// ---- Atom 1.0 structs ------------------------------------------------------

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Links   []atomLink  `xml:"link"`
	Author  atomAuthor  `xml:"author"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Href string `xml:"href,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomEntry struct {
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Link    atomLink `xml:"link"`
	Updated string   `xml:"updated"`
	Summary string   `xml:"summary"`
}

// GenerateFeeds writes feed.xml (RSS 2.0) and atom.xml (Atom 1.0) to outDir.
// Articles are sorted newest-first. baseURL must not have a trailing slash.
// When cfg has I18n.Locales configured, per-locale feeds are also written:
//
//	{locale}/feed.xml and {locale}/atom.xml for each non-default locale.
//
// The root feed.xml / atom.xml contain only articles from the default locale
// (or all articles when i18n is not configured).
func GenerateFeeds(outDir, baseURL, siteTitle string, articles []*model.ProcessedArticle, cfg model.Config) error {
	sorted := make([]*model.ProcessedArticle, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FrontMatter.Date.After(sorted[j].FrontMatter.Date)
	})

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// When i18n is active, filter root feeds to the default locale only and
	// write per-locale feeds under their locale subdirectory.
	if len(cfg.I18n.Locales) > 0 {
		rootArticles := filterFeedArticles(sorted, cfg.I18n.DefaultLocale)
		if err := writeRSS(outDir, baseURL, siteTitle, rootArticles); err != nil {
			return err
		}
		if err := writeAtom(outDir, baseURL, siteTitle, rootArticles, cfg); err != nil {
			return err
		}
		for _, loc := range cfg.I18n.Locales {
			if loc == cfg.I18n.DefaultLocale {
				continue // already written at root
			}
			locDir := filepath.Join(outDir, loc)
			if err := os.MkdirAll(locDir, 0o755); err != nil {
				return err
			}
			locArticles := filterFeedArticles(sorted, loc)
			// channelURL is the locale index (used for <channel><link>).
			// Article item links use the site root baseURL because a.URL already
			// includes the locale prefix (e.g. /ja/posts/hello/).
			var channelURL string
			if baseURL != "" {
				channelURL = baseURL + "/" + loc + "/"
			} else {
				channelURL = "/" + loc + "/"
			}
			if err := writeRSSWithChannelURL(locDir, baseURL, channelURL, siteTitle, locArticles); err != nil {
				return err
			}
			if err := writeAtomWithChannelURL(locDir, baseURL, channelURL, siteTitle, locArticles, cfg); err != nil {
				return err
			}
		}
		return nil
	}

	if err := writeRSS(outDir, baseURL, siteTitle, sorted); err != nil {
		return err
	}
	return writeAtom(outDir, baseURL, siteTitle, sorted, cfg)
}

func writeRSS(outDir, baseURL, title string, articles []*model.ProcessedArticle) error {
	// BUG-C: channel URL must have a trailing slash (consistent with writeAtom).
	return writeRSSWithChannelURL(outDir, baseURL, baseURL+"/", title, articles)
}

func writeRSSWithChannelURL(outDir, itemBaseURL, channelURL, title string, articles []*model.ProcessedArticle) error {
	now := time.Now().UTC().Format(time.RFC1123Z)
	ch := rssChannel{
		Title:       title,
		Link:        channelURL,
		Description: title,
		PubDate:     now,
	}
	for _, a := range articles {
		if a.FrontMatter.Date.IsZero() {
			continue // skip articles with no publication date
		}
		link := articleLink(itemBaseURL, a)
		ch.Items = append(ch.Items, rssItem{
			Title:       a.FrontMatter.Title,
			Link:        link,
			Description: a.Summary,
			PubDate:     a.FrontMatter.Date.UTC().Format(time.RFC1123Z),
			GUID:        link,
		})
	}
	root := rssRoot{Version: "2.0", Channel: ch}
	return writeXML(filepath.Join(outDir, "feed.xml"), root)
}

func writeAtom(outDir, baseURL, title string, articles []*model.ProcessedArticle, cfg model.Config) error {
	return writeAtomWithChannelURL(outDir, baseURL, baseURL+"/", title, articles, cfg)
}

func writeAtomWithChannelURL(outDir, itemBaseURL, channelURL, title string, articles []*model.ProcessedArticle, cfg model.Config) error {
	updated := time.Now().UTC().Format(time.RFC3339)
	for _, a := range articles {
		if !a.FrontMatter.Date.IsZero() {
			updated = a.FrontMatter.Date.UTC().Format(time.RFC3339)
			break
		}
	}
	author := cfg.Theme.Params["author"]
	baseAtomURL := channelURL + "atom.xml"
	feed := atomFeed{
		Xmlns: "http://www.w3.org/2005/Atom",
		ID:    baseAtomURL,
		Title: title,
		Links: []atomLink{
			{Rel: "alternate", Type: "text/html", Href: channelURL},
			{Rel: "self", Type: "application/atom+xml", Href: baseAtomURL},
		},
		Author:  atomAuthor{Name: author},
		Updated: updated,
	}
	for _, a := range articles {
		if a.FrontMatter.Date.IsZero() {
			continue // skip articles with no publication date
		}
		link := articleLink(itemBaseURL, a)
		feed.Entries = append(feed.Entries, atomEntry{
			ID:      link,
			Title:   a.FrontMatter.Title,
			Link:    atomLink{Rel: "alternate", Type: "text/html", Href: link},
			Updated: a.FrontMatter.Date.UTC().Format(time.RFC3339),
			Summary: a.Summary,
		})
	}
	return writeXML(filepath.Join(outDir, "atom.xml"), feed)
}

// articleLink returns the full URL for an article.
// When a.URL is set (i18n mode), it is appended to baseURL.
// Otherwise the URL is constructed from the article slug.
func articleLink(baseURL string, a *model.ProcessedArticle) string {
	if a.URL != "" {
		return baseURL + a.URL
	}
	s := a.FrontMatter.Slug
	if s == "" {
		s = slugify(a.FrontMatter.Title)
	}
	return baseURL + "/posts/" + s + "/"
}

func writeXML(path string, v interface{}) error {
	var buf bytes.Buffer
	if _, err := buf.WriteString(xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	return writeFileAtomic(path, buf.Bytes(), 0o644)
}

// filterFeedArticles returns only articles whose Locale matches locale.
func filterFeedArticles(articles []*model.ProcessedArticle, locale string) []*model.ProcessedArticle {
	var out []*model.ProcessedArticle
	for _, a := range articles {
		if a.Locale == locale {
			out = append(out, a)
		}
	}
	return out
}
