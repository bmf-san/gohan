package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmf-san/gohan/internal/model"
)

// GenerateSitemap writes sitemap.xml to outDir, listing all article URLs.
// When articles have Translations populated (i18n), xhtml:link hreflang
// alternates are included for SEO.
// Articles are sorted newest-first. baseURL must not have a trailing slash.
func GenerateSitemap(outDir, baseURL string, articles []*model.ProcessedArticle) error {
	sorted := make([]*model.ProcessedArticle, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FrontMatter.Date.After(sorted[j].FrontMatter.Date)
	})

	// Determine whether any article has hreflang alternates.
	needsHreflang := false
	for _, a := range sorted {
		if len(a.Translations) > 0 {
			needsHreflang = true
			break
		}
	}

	var buf strings.Builder
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	if needsHreflang {
		buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">` + "\n")
	} else {
		buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	}

	for _, a := range sorted {
		// Prefer pre-computed URL; fall back to slug-based path for single-lang sites.
		articleURL := a.URL
		if articleURL == "" {
			s := a.FrontMatter.Slug
			if s == "" {
				s = slugify(a.FrontMatter.Title)
			}
			articleURL = "/posts/" + s + "/"
		}
		loc := baseURL + articleURL

		buf.WriteString("  <url>\n")
		buf.WriteString("    <loc>" + loc + "</loc>\n")
		if !a.FrontMatter.Date.IsZero() {
			buf.WriteString("    <lastmod>" + a.FrontMatter.Date.UTC().Format("2006-01-02") + "</lastmod>\n")
		}
		if len(a.Translations) > 0 {
			// Self-referencing hreflang (recommended by Google).
			locale := a.Locale
			if locale == "" {
				locale = "x-default"
			}
			buf.WriteString(fmt.Sprintf("    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", locale, loc))
			for _, tr := range a.Translations {
				buf.WriteString(fmt.Sprintf("    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", tr.Locale, baseURL+tr.URL))
			}
		}
		buf.WriteString("  </url>\n")
	}
	buf.WriteString("</urlset>\n")

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "sitemap.xml"), []byte(buf.String()), 0o644)
}
