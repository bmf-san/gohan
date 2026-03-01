package generator

import (
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
	Title   string      `xml:"title"`
	Link    atomLink    `xml:"link"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type atomEntry struct {
	Title   string   `xml:"title"`
	Link    atomLink `xml:"link"`
	Updated string   `xml:"updated"`
	Summary string   `xml:"summary"`
}

// GenerateFeeds writes feed.xml (RSS 2.0) and atom.xml (Atom 1.0) to outDir.
// Articles are sorted newest-first. baseURL must not have a trailing slash.
func GenerateFeeds(outDir, baseURL, siteTitle string, articles []*model.ProcessedArticle) error {
	sorted := make([]*model.ProcessedArticle, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FrontMatter.Date.After(sorted[j].FrontMatter.Date)
	})

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := writeRSS(outDir, baseURL, siteTitle, sorted); err != nil {
		return err
	}
	return writeAtom(outDir, baseURL, siteTitle, sorted)
}

func writeRSS(outDir, baseURL, title string, articles []*model.ProcessedArticle) error {
	now := time.Now().UTC().Format(time.RFC1123Z)
	ch := rssChannel{
		Title:       title,
		Link:        baseURL,
		Description: title,
		PubDate:     now,
	}
	for _, a := range articles {
		s := a.FrontMatter.Slug
		if s == "" {
			s = slugify(a.FrontMatter.Title)
		}
		link := baseURL + "/posts/" + s + "/"
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

func writeAtom(outDir, baseURL, title string, articles []*model.ProcessedArticle) error {
	updated := time.Now().UTC().Format(time.RFC3339)
	if len(articles) > 0 {
		updated = articles[0].FrontMatter.Date.UTC().Format(time.RFC3339)
	}
	feed := atomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		Title:   title,
		Link:    atomLink{Href: baseURL},
		Updated: updated,
	}
	for _, a := range articles {
		s := a.FrontMatter.Slug
		if s == "" {
			s = slugify(a.FrontMatter.Title)
		}
		feed.Entries = append(feed.Entries, atomEntry{
			Title:   a.FrontMatter.Title,
			Link:    atomLink{Href: baseURL + "/posts/" + s + "/"},
			Updated: a.FrontMatter.Date.UTC().Format(time.RFC3339),
			Summary: a.Summary,
		})
	}
	return writeXML(filepath.Join(outDir, "atom.xml"), feed)
}

func writeXML(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return enc.Flush()
}
