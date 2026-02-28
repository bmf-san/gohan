package generator

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"sort"

	"github.com/bmf-san/gohan/internal/model"
)

// urlSet is the root element of sitemap.xml.
type urlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// sitemapURL is a single <url> entry.
type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// GenerateSitemap writes sitemap.xml to outDir, listing all article URLs.
// Articles are sorted newest-first. baseURL must not have a trailing slash.
func GenerateSitemap(outDir, baseURL string, articles []*model.ProcessedArticle) error {
	sorted := make([]*model.ProcessedArticle, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FrontMatter.Date.After(sorted[j].FrontMatter.Date)
	})

	us := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, a := range sorted {
		s := a.FrontMatter.Slug
		if s == "" {
			s = slugify(a.FrontMatter.Title)
		}
		us.URLs = append(us.URLs, sitemapURL{
			Loc:     baseURL + "/posts/" + s + "/",
			LastMod: a.FrontMatter.Date.UTC().Format("2006-01-02"),
		})
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	out, err := os.Create(filepath.Join(outDir, "sitemap.xml"))
	if err != nil {
		return err
	}
	defer out.Close()

	out.WriteString(xml.Header)
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")
	if err := enc.Encode(us); err != nil {
		return err
	}
	return enc.Flush()
}
