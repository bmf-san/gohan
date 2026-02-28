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
func (e *Engine) Load(templateDir string, funcs template.FuncMap) error {
	allFuncs := builtinFuncs()
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
func builtinFuncs() template.FuncMap {
	conv := parser.NewConverter(parser.WithGFM())
	return template.FuncMap{
		// formatDate formats t using layout (e.g. "2006-01-02").
		"formatDate": func(layout string, t time.Time) string {
			return t.Format(layout)
		},
		// tagURL returns the canonical URL for a tag.
		"tagURL": func(tag string) string {
			return "/tags/" + toSlug(tag) + "/"
		},
		// categoryURL returns the canonical URL for a category.
		"categoryURL": func(cat string) string {
			return "/categories/" + toSlug(cat) + "/"
		},
		// markdownify converts a Markdown string to safe HTML.
		"markdownify": func(s string) (template.HTML, error) {
			return conv.Convert([]byte(s))
		},
	}
}

// toSlug converts a display name to a URL-friendly slug by lowercasing and
// replacing spaces with hyphens.
func toSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// Ensure Engine implements TemplateEngine at compile time.
var _ TemplateEngine = (*Engine)(nil)
