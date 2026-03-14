package template

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

// writeTmpl writes a template file into dir and returns its path.
func writeTmpl(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTmpl: %v", err)
	}
	return path
}

// renderStr renders a named template to string.
func renderStr(t *testing.T, e *Engine, name string, data *model.Site) string {
	t.Helper()
	var buf bytes.Buffer
	if err := e.Render(&buf, name, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return buf.String()
}

// minSite returns a minimal *model.Site for testing.
func minSite(title string) *model.Site {
	return &model.Site{Config: model.Config{Site: model.SiteConfig{Title: title}}}
}

func TestEngine_Load_Valid(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "index.html", `{{define "index.html"}}hello{{end}}`)
	e := NewEngine()
	if err := e.Load(dir, nil, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngine_Load_NoHTML(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "readme.txt", "nothing here")
	e := NewEngine()
	if err := e.Load(dir, nil, ""); err == nil {
		t.Error("expected error when no .html files, got nil")
	}
}

func TestEngine_Load_DirNotFound(t *testing.T) {
	e := NewEngine()
	if err := e.Load("/nonexistent/themes/default", nil, ""); err == nil {
		t.Error("expected error for missing directory, got nil")
	}
}

func TestEngine_Render_NotLoaded(t *testing.T) {
	e := NewEngine()
	var buf bytes.Buffer
	if err := e.Render(&buf, "index.html", minSite("")); err == nil {
		t.Error("expected error when Load not called, got nil")
	}
}

func TestEngine_Render_MissingTemplate(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "index.html", `{{define "index.html"}}hi{{end}}`)
	e := NewEngine()
	if err := e.Load(dir, nil, ""); err != nil {
		t.Fatalf("Load: %v", err)
	}
	var buf bytes.Buffer
	if err := e.Render(&buf, "nonexistent.html", minSite("")); err == nil {
		t.Error("expected error for missing template, got nil")
	}
}

func TestEngine_Render_VariableExpansion(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "index.html", `{{define "index.html"}}{{.Config.Site.Title}}{{end}}`)
	e := NewEngine()
	if err := e.Load(dir, nil, ""); err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := renderStr(t, e, "index.html", minSite("My Blog"))
	if got != "My Blog" {
		t.Errorf("got %q, want %q", got, "My Blog")
	}
}

func TestEngine_Render_MultipleTemplates(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "index.html", `{{define "index.html"}}index{{end}}`)
	writeTmpl(t, dir, "article.html", `{{define "article.html"}}article{{end}}`)
	e := NewEngine()
	if err := e.Load(dir, nil, ""); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := renderStr(t, e, "index.html", minSite("")); got != "index" {
		t.Errorf("index: got %q", got)
	}
	if got := renderStr(t, e, "article.html", minSite("")); got != "article" {
		t.Errorf("article: got %q", got)
	}
}

func TestEngine_Builtin_FormatDate(t *testing.T) {
	fns := builtinFuncs("")
	fn, ok := fns["formatDate"].(func(string, time.Time) string)
	if !ok {
		t.Fatal("formatDate not found in builtinFuncs")
	}
	t1 := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	got := fn("2006-01-02", t1)
	if got != "2024-03-15" {
		t.Errorf("formatDate: got %q, want %q", got, "2024-03-15")
	}
}

func TestEngine_Builtin_TagURL(t *testing.T) {
	// no-locale (non-i18n): root path
	fns := builtinFuncs("")
	fn, ok := fns["tagURL"].(func(string, string) string)
	if !ok {
		t.Fatal("tagURL not found or wrong type")
	}
	if got := fn("", "Go Programming"); got != "/tags/go-programming/" {
		t.Errorf("tagURL no-locale: got %q, want %q", got, "/tags/go-programming/")
	}

	// default locale ("en"): same root path — no prefix
	fnsEN := builtinFuncs("en")
	fnEN := fnsEN["tagURL"].(func(string, string) string)
	if got := fnEN("en", "Go Programming"); got != "/tags/go-programming/" {
		t.Errorf("tagURL en (default): got %q, want %q", got, "/tags/go-programming/")
	}

	// non-default locale ("ja"): locale-prefixed path
	if got := fnEN("ja", "Go Programming"); got != "/ja/tags/go-programming/" {
		t.Errorf("tagURL ja: got %q, want %q", got, "/ja/tags/go-programming/")
	}
}

func TestEngine_Builtin_CategoryURL(t *testing.T) {
	// no-locale (non-i18n): root path
	fns := builtinFuncs("")
	fn, ok := fns["categoryURL"].(func(string, string) string)
	if !ok {
		t.Fatal("categoryURL not found or wrong type")
	}
	if got := fn("", "Web Development"); got != "/categories/web-development/" {
		t.Errorf("categoryURL no-locale: got %q, want %q", got, "/categories/web-development/")
	}

	// default locale ("en"): root path — no prefix
	fnsEN := builtinFuncs("en")
	fnEN := fnsEN["categoryURL"].(func(string, string) string)
	if got := fnEN("en", "Web Development"); got != "/categories/web-development/" {
		t.Errorf("categoryURL en (default): got %q, want %q", got, "/categories/web-development/")
	}

	// non-default locale ("ja"): locale-prefixed path
	if got := fnEN("ja", "Web Development"); got != "/ja/categories/web-development/" {
		t.Errorf("categoryURL ja: got %q, want %q", got, "/ja/categories/web-development/")
	}
}

func TestEngine_Builtin_Markdownify(t *testing.T) {
	fns := builtinFuncs("")
	fn, ok := fns["markdownify"].(func(string) (template.HTML, error))
	if !ok {
		t.Fatal("markdownify not found")
	}
	html, err := fn("**bold**")
	if err != nil {
		t.Fatalf("markdownify error: %v", err)
	}
	if !strings.Contains(string(html), "<strong>bold</strong>") {
		t.Errorf("markdownify: got %q", html)
	}
}

func TestEngine_CustomFunc(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "custom.html", `{{define "custom.html"}}{{shout .Config.Site.Title}}{{end}}`)
	e := NewEngine()
	customFuncs := template.FuncMap{
		"shout": func(s string) string { return strings.ToUpper(s) + "!" },
	}
	if err := e.Load(dir, customFuncs, ""); err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := renderStr(t, e, "custom.html", minSite("hello"))
	if got != "HELLO!" {
		t.Errorf("custom func: got %q, want %q", got, "HELLO!")
	}
}

func TestEngine_Load_SubDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "partials")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTmpl(t, dir, "index.html", `{{define "index.html"}}main{{end}}`)
	writeTmpl(t, sub, "partial.html", `{{define "partial.html"}}part{{end}}`)
	e := NewEngine()
	if err := e.Load(dir, nil, ""); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := renderStr(t, e, "index.html", minSite("")); got != "main" {
		t.Errorf("index: got %q", got)
	}
	if got := renderStr(t, e, "partial.html", minSite("")); got != "part" {
		t.Errorf("partial: got %q", got)
	}
}

func TestToSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Go Programming", "go-programming"},
		{"hello", "hello"},
		{"Web Dev 101", "web-dev-101"},
		{"UPPER CASE", "upper-case"},
	}
	for _, c := range cases {
		got := toSlug(c.in)
		if got != c.want {
			t.Errorf("toSlug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEngine_Builtin_PaginationPages(t *testing.T) {
	fns := builtinFuncs("")
	fn, ok := fns["paginationPages"].(func(int, int) []int)
	if !ok {
		t.Fatal("paginationPages not found in builtinFuncs")
	}

	// total <= 1 returns nil
	if got := fn(1, 0); got != nil {
		t.Errorf("paginationPages(1,0): got %v, want nil", got)
	}
	if got := fn(1, 1); got != nil {
		t.Errorf("paginationPages(1,1): got %v, want nil", got)
	}

	// small range: all pages shown
	got := fn(2, 3)
	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Errorf("paginationPages(2,3): got %v, want %v", got, want)
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("paginationPages(2,3)[%d]: got %d, want %d", i, got[i], want[i])
			}
		}
	}

	// large range: ellipsis inserted
	got = fn(5, 10)
	if len(got) == 0 {
		t.Fatal("paginationPages(5,10): got empty slice")
	}
	if got[0] != 1 {
		t.Errorf("paginationPages(5,10): first page should be 1, got %d", got[0])
	}
	if got[len(got)-1] != 10 {
		t.Errorf("paginationPages(5,10): last page should be 10, got %d", got[len(got)-1])
	}
	hasEllipsis := false
	for _, p := range got {
		if p == -1 {
			hasEllipsis = true
			break
		}
	}
	if !hasEllipsis {
		t.Errorf("paginationPages(5,10): expected ellipsis (-1) in %v", got)
	}

	// current at start: ellipsis only at end
	got = fn(1, 10)
	if got[0] != 1 {
		t.Errorf("paginationPages(1,10): got[0]=%d", got[0])
	}
	if got[len(got)-1] != 10 {
		t.Errorf("paginationPages(1,10): last=%d", got[len(got)-1])
	}

	// no consecutive -1
	for i := 1; i < len(got); i++ {
		if got[i] == -1 && got[i-1] == -1 {
			t.Errorf("paginationPages(1,10): consecutive ellipsis at %d", i)
		}
	}
}

func TestEngine_Builtin_PageURL(t *testing.T) {
	fns := builtinFuncs("")
	fn, ok := fns["pageURL"].(func(string, int) string)
	if !ok {
		t.Fatal("pageURL not found in builtinFuncs")
	}

	cases := []struct {
		baseURL string
		p       int
		want    string
	}{
		{"", 1, "/"},
		{"", 0, "/"},
		{"/tags/go", 1, "/tags/go/"},
		{"/tags/go", 0, "/tags/go/"},
		{"/tags/go", 2, "/tags/go/page/2/"},
		{"/categories/web", 3, "/categories/web/page/3/"},
	}
	for _, c := range cases {
		got := fn(c.baseURL, c.p)
		if got != c.want {
			t.Errorf("pageURL(%q, %d): got %q, want %q", c.baseURL, c.p, got, c.want)
		}
	}
}

func TestToSlug_Empty(t *testing.T) {
	// Empty input must return "untitled" not an empty string,
	// so tagURL("", "") never produces "/tags//".
	if got := toSlug(""); got != "untitled" {
		t.Errorf("toSlug(%q) = %q, want %q", "", got, "untitled")
	}
}

func TestPaginationPages_InvalidCurrent(t *testing.T) {
	// current < 1 must return nil (invalid input guard, BUG-09).
	fns := builtinFuncs("")
	fn := fns["paginationPages"].(func(int, int) []int)
	if got := fn(0, 10); got != nil {
		t.Errorf("paginationPages(0,10): expected nil, got %v", got)
	}
	if got := fn(-1, 10); got != nil {
		t.Errorf("paginationPages(-1,10): expected nil, got %v", got)
	}
}
