package generator

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bmf-san/gohan/internal/model"
)

var (
	metaTagRe        = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	robotsNameRe     = regexp.MustCompile(`(?i)\bname\s*=\s*["']?(?:robots|googlebot)(?:["'\s/>]|$)`)
	noindexContentRe = regexp.MustCompile(`(?i)\bcontent\s*=\s*["'][^"']*\b(?:noindex|none)\b`)
)

// maxHeadBytes bounds how much of a rendered page is scanned for robots meta
// tags; they belong in <head>, so the rest of the document is never read.
const maxHeadBytes = 64 << 10

// noindexChecker reports whether the rendered page for a URL path declares
// <meta name="robots" content="noindex"> (or "none"). Results are cached
// because article translations are looked up more than once.
type noindexChecker struct {
	outDir string
	cache  map[string]bool
}

func newNoindexChecker(outDir string) *noindexChecker {
	return &noindexChecker{outDir: outDir, cache: map[string]bool{}}
}

func (c *noindexChecker) isNoindex(urlPath string) bool {
	if v, ok := c.cache[urlPath]; ok {
		return v
	}
	v := pageIsNoindex(c.outDir, urlPath)
	c.cache[urlPath] = v
	return v
}

// pageIsNoindex resolves urlPath to the HTML file under outDir and scans its
// <head> for a robots noindex directive. Missing or unreadable files are
// treated as indexable so the sitemap never loses pages it cannot inspect.
func pageIsNoindex(outDir, urlPath string) bool {
	p := urlPath
	if i := strings.IndexAny(p, "?#"); i >= 0 {
		p = p[:i]
	}
	file := filepath.Join(outDir, filepath.FromSlash(strings.TrimPrefix(p, "/")))
	if p == "" || strings.HasSuffix(p, "/") {
		file = filepath.Join(file, "index.html")
	}
	f, err := os.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()

	head, err := readHead(f)
	if err != nil {
		return false
	}
	for _, tag := range metaTagRe.FindAll(head, -1) {
		if robotsNameRe.Match(tag) && noindexContentRe.Match(tag) {
			return true
		}
	}
	return false
}

// readHead returns the document up to </head>, capped at maxHeadBytes.
func readHead(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	chunk := make([]byte, 8<<10)
	for buf.Len() < maxHeadBytes {
		n, err := r.Read(chunk)
		buf.Write(chunk[:n])
		if i := bytes.Index(bytes.ToLower(buf.Bytes()), []byte("</head>")); i >= 0 {
			return buf.Bytes()[:i], nil
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// GenerateSitemap writes sitemap.xml to outDir, listing all article URLs,
// virtual pages produced by SitePlugins (e.g. /bookshelf/), and any extra
// URL paths supplied via extraURLs (e.g. taxonomy and archive pages).
// When articles have Translations populated (i18n), xhtml:link hreflang
// alternates are included for SEO.
// Articles are sorted newest-first. baseURL must not have a trailing slash.
// When cfg has I18n.Locales configured, the locale index pages (/ and /ja/
// etc.) are prepended to the sitemap as important entry points.
// Pages whose rendered HTML in outDir carries <meta name="robots"
// content="noindex"> are left out, so a template that noindexes thin pages
// (for example, tags with a single article) keeps them out of the sitemap too.
// GenerateSitemap must therefore run after the HTML has been rendered.
func GenerateSitemap(outDir, baseURL string, articles []*model.ProcessedArticle, virtualPages []*model.VirtualPage, extraURLs []string, cfg model.Config) error {
	noindex := newNoindexChecker(outDir)
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
			indexPath := "/" + loc + "/"
			if loc == cfg.I18n.DefaultLocale {
				indexPath = "/"
			}
			if noindex.isNoindex(indexPath) {
				continue
			}
			buf.WriteString("  <url>\n")
			buf.WriteString("    <loc>" + html.EscapeString(baseURL+indexPath) + "</loc>\n")
			buf.WriteString("  </url>\n")
		}
	}

	// Emit virtual page URLs (e.g. /bookshelf/, /ja/bookshelf/).
	for _, vp := range virtualPages {
		if vp.URL == "" || noindex.isNoindex(vp.URL) {
			continue
		}
		buf.WriteString("  <url>\n")
		buf.WriteString("    <loc>" + html.EscapeString(baseURL+vp.URL) + "</loc>\n")
		buf.WriteString("  </url>\n")
	}

	// Emit extra URLs (taxonomy pages: tags, categories, archives).
	for _, u := range extraURLs {
		if u == "" || noindex.isNoindex(u) {
			continue
		}
		buf.WriteString("  <url>\n")
		buf.WriteString("    <loc>" + html.EscapeString(baseURL+u) + "</loc>\n")
		buf.WriteString("  </url>\n")
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
		if noindex.isNoindex(articleURL) {
			continue
		}
		loc := baseURL + articleURL

		buf.WriteString("  <url>\n")
		buf.WriteString("    <loc>" + html.EscapeString(loc) + "</loc>\n")
		if !a.FrontMatter.LastMod.IsZero() {
			buf.WriteString("    <lastmod>" + a.FrontMatter.LastMod.UTC().Format("2006-01-02") + "</lastmod>\n")
		} else if !a.FrontMatter.Date.IsZero() {
			buf.WriteString("    <lastmod>" + a.FrontMatter.Date.UTC().Format("2006-01-02") + "</lastmod>\n")
		}
		// A noindex translation must not be advertised as an alternate either.
		var translations []model.LocaleRef
		for _, tr := range a.Translations {
			if !noindex.isNoindex(tr.URL) {
				translations = append(translations, tr)
			}
		}
		if len(translations) > 0 {
			// Self-referencing hreflang (recommended by Google).
			locale := a.Locale
			if locale == "" {
				locale = "x-default"
			}
			fmt.Fprintf(&buf, "    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", locale, html.EscapeString(loc))
			for _, tr := range translations {
				fmt.Fprintf(&buf, "    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", tr.Locale, html.EscapeString(baseURL+tr.URL))
			}
			// x-default points to the default-locale variant so search engines
			// have a clear fallback when no locale matches the visitor's language.
			if cfg.I18n.DefaultLocale != "" {
				xdefault := loc // self is default unless we find a translation that is
				if a.Locale != cfg.I18n.DefaultLocale {
					for _, tr := range translations {
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
