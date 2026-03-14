package template

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/parser"
)

// Engine is the concrete implementation of TemplateEngine.
// It loads html/template files from a theme directory and renders pages with
// site/article data.
type Engine struct {
	tmpl *template.Template
}

// NewEngine returns a new, empty Engine.  Call Load before calling Render.
func NewEngine() *Engine {
	return &Engine{}
}

// Load parses all .html files found (recursively) under templateDir.
// Built-in helper functions (formatDate, tagURL, categoryURL, markdownify) are
// registered automatically; callers may supply additional functions via funcs.
// defaultLocale is the site's primary locale (e.g. "en"); pass "" for non-i18n
// sites. tagURL and categoryURL use it to decide when to omit the locale prefix.
func (e *Engine) Load(templateDir string, funcs template.FuncMap, defaultLocale string) error {
	allFuncs := builtinFuncs(defaultLocale)
	for k, v := range funcs {
		allFuncs[k] = v
	}

	var files []string
	err := filepath.WalkDir(templateDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".html" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("template: walk %s: %w", templateDir, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("template: no .html files found in %s", templateDir)
	}

	tmpl, err := template.New("").Funcs(allFuncs).ParseFiles(files...)
	if err != nil {
		return fmt.Errorf("template: parse: %w", err)
	}
	e.tmpl = tmpl
	return nil
}

// Render executes the named template, writing the rendered output to w.
// templateName should match the base filename (e.g. "article.html") or the
// name of a {{define}} block inside a loaded file.
func (e *Engine) Render(w io.Writer, templateName string, data *model.Site) error {
	if e.tmpl == nil {
		return fmt.Errorf("template: not loaded; call Load first")
	}
	if err := e.tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		return fmt.Errorf("template: render %q: %w", templateName, err)
	}
	return nil
}

// builtinFuncs returns the default template function map.
// defaultLocale is the site's primary locale; tagURL and categoryURL use it to
// omit the locale prefix for the default locale (and for non-i18n sites when
// "" is passed).
func builtinFuncs(defaultLocale string) template.FuncMap {
	conv := parser.NewConverter(parser.WithGFM())
	return template.FuncMap{
		// formatDate formats t using layout (e.g. "2006-01-02").
		"formatDate": func(layout string, t time.Time) string {
			return t.Format(layout)
		},
		// tagURL returns the locale-aware canonical URL for a tag.
		// locale="" or locale==defaultLocale → /tags/{slug}/
		// otherwise → /{locale}/tags/{slug}/
		"tagURL": func(locale, tag string) string {
			if locale == "" || locale == defaultLocale {
				return "/tags/" + toSlug(tag) + "/"
			}
			return "/" + locale + "/tags/" + toSlug(tag) + "/"
		},
		// categoryURL returns the locale-aware canonical URL for a category.
		"categoryURL": func(locale, cat string) string {
			if locale == "" || locale == defaultLocale {
				return "/categories/" + toSlug(cat) + "/"
			}
			return "/" + locale + "/categories/" + toSlug(cat) + "/"
		},
		// markdownify converts a Markdown string to safe HTML.
		"markdownify": func(s string) (template.HTML, error) {
			return conv.Convert([]byte(s))
		},
		// paginationPages returns a slice of page numbers (and -1 for ellipsis)
		// suitable for rendering a pagination control. For example, for
		// current=5 total=10 it returns [1 -1 4 5 6 -1 10].
		"paginationPages": func(current, total int) []int {
			if total <= 1 || current < 1 {
				return nil
			}
			pages := []int{}
			addPage := func(p int) {
				if len(pages) > 0 && pages[len(pages)-1] == -1 && p == -1 {
					return
				}
				pages = append(pages, p)
			}
			for p := 1; p <= total; p++ {
				if p == 1 || p == total || (p >= current-2 && p <= current+2) {
					addPage(p)
				} else if len(pages) > 0 && pages[len(pages)-1] != -1 {
					addPage(-1) // ellipsis
				}
			}
			return pages
		},
		// pageURL returns the URL for page number p within a base URL path.
		// Page 1 returns baseURL+"/" (or "/" when baseURL is empty).
		"pageURL": func(baseURL string, p int) string {
			if p <= 1 {
				if baseURL == "" {
					return "/"
				}
				return baseURL + "/"
			}
			return fmt.Sprintf("%s/page/%d/", baseURL, p)
		},
	}
}

// toSlug converts a display name to a URL-friendly slug by lowercasing ASCII
// letters and replacing spaces with hyphens. Non-ASCII characters (e.g.
// Japanese, accented Latin) are kept intact, matching the behaviour of
// tagNorm in the generator so template links are consistent with page paths.
// Returns "untitled" when the input produces an empty result.
func toSlug(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		case r == ' ':
			b.WriteRune('-')
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "untitled"
	}
	return b.String()
}

// Ensure Engine implements TemplateEngine at compile time.
var _ TemplateEngine = (*Engine)(nil)
