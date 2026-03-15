package generator

import (
	"fmt"
	"html"
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
// When cfg has I18n.Locales configured, the locale index pages (/ and /ja/
// etc.) are prepended to the sitemap as important entry points.
func GenerateSitemap(outDir, baseURL string, articles []*model.ProcessedArticle, cfg model.Config) error {
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

	// Prepend locale index pages (/, /ja/, ...) when i18n is configured.
	if len(cfg.I18n.Locales) > 0 {
		for _, loc := range cfg.I18n.Locales {
			var indexURL string
			if loc == cfg.I18n.DefaultLocale {
				indexURL = baseURL + "/"
			} else {
				indexURL = baseURL + "/" + loc + "/"
			}
			buf.WriteString("  <url>\n")
			buf.WriteString("    <loc>" + html.EscapeString(indexURL) + "</loc>\n")
			buf.WriteString("  </url>\n")
		}
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
		buf.WriteString("    <loc>" + html.EscapeString(loc) + "</loc>\n")
		if !a.FrontMatter.Date.IsZero() {
			buf.WriteString("    <lastmod>" + a.FrontMatter.Date.UTC().Format("2006-01-02") + "</lastmod>\n")
		}
		if len(a.Translations) > 0 {
			// Self-referencing hreflang (recommended by Google).
			// href values must be XML-escaped (consistent with <loc>).
			locale := a.Locale
			if locale == "" {
				locale = "x-default"
			}
			fmt.Fprintf(&buf, "    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", locale, html.EscapeString(loc))
			for _, tr := range a.Translations {
				fmt.Fprintf(&buf, "    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", tr.Locale, html.EscapeString(baseURL+tr.URL))
			}
			// x-default points to the default-locale variant so search engines
			// have a clear fallback when no locale matches the visitor's language.
			if cfg.I18n.DefaultLocale != "" {
				xdefault := loc // self is default unless we find a translation that is
				if a.Locale != cfg.I18n.DefaultLocale {
					for _, tr := range a.Translations {
						if tr.Locale == cfg.I18n.DefaultLocale {
							xdefault = baseURL + tr.URL
							break
						}
					}
				}
				fmt.Fprintf(&buf, "    <xhtml:link rel=\"alternate\" hreflang=\"x-default\" href=\"%s\"/>\n", html.EscapeString(xdefault))
			}
		}
		buf.WriteString("  </url>\n")
	}
	buf.WriteString("</urlset>\n")

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(outDir, "sitemap.xml"), []byte(buf.String()), 0o644)
}
