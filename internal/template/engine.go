// Package template loads Go html/template files and renders articles into HTML.
package template

import (
	"html/template"
	"io"

	"github.com/bmf-san/gohan/internal/model"
)

// TemplateEngine loads theme templates from disk and renders pages using
// the standard library html/template package.
type TemplateEngine interface {
	// Load parses all template files rooted at templateDir (e.g. theme/templates).
	Load(templateDir string, funcs template.FuncMap) error

	// Render executes the named template with the given data, writing the result
	// to w.  templateName corresponds to a file base name such as "article.html".
	Render(w io.Writer, templateName string, data *model.Site) error
}
