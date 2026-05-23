package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func runNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	title := fs.String("title", "", "article title (defaults to slug)")
	articleType := fs.String("type", "post", "article type / content section name (e.g. post, page, tutorial)")
	archetype := fs.String("archetype", "", "archetype template name under archetypes/ (defaults to --type value)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: gohan new [--title=\"...\"] [--type=<section>] [--archetype=<name>] <slug>")
	}
	slug := fs.Arg(0)

	// Validate slug: no path separators, no whitespace.
	if strings.ContainsAny(slug, "/\\ \t") {
		return fmt.Errorf("slug must not contain path separators or whitespace: %q", slug)
	}

	// Determine content section directory.
	// `post` and `page` are pluralised for backwards compatibility with the
	// legacy convention (content/posts, content/pages). Any other section
	// name is used verbatim under content/.
	var dir string
	switch *articleType {
	case "post", "":
		dir = filepath.Join("content", "posts")
	case "page":
		dir = filepath.Join("content", "pages")
	default:
		dir = filepath.Join("content", *articleType)
	}

	// Resolve title.
	articleTitle := *title
	if articleTitle == "" {
		articleTitle = slugToTitle(slug)
	}

	// Resolve archetype template name (falls back to the section type).
	archetypeName := *archetype
	if archetypeName == "" {
		archetypeName = *articleType
		if archetypeName == "" {
			archetypeName = "post"
		}
	}

	today := time.Now().Format("2006-01-02")
	content, err := renderArchetype(archetypeName, archetypeData{
		Title: articleTitle,
		Date:  today,
		Slug:  slug,
		Type:  *articleType,
	})
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, slug+".md")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Use O_CREATE|O_EXCL for an atomic create-or-fail, eliminating the
	// TOCTOU race between an os.Stat check and os.WriteFile.
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("file already exists: %s", filePath)
		}
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := io.WriteString(f, content); err != nil {
		_ = os.Remove(filePath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("created: %s\n", filePath)
	return nil
}

// archetypeData is the set of values exposed to archetype templates.
type archetypeData struct {
	Title string
	Date  string
	Slug  string
	Type  string
}

// renderArchetype loads archetypes/<name>.md from the current working
// directory and renders it through text/template using data. When no
// archetype file exists, a built-in default for the given name is used; if
// no such default exists either, an error is returned.
func renderArchetype(name string, data archetypeData) (string, error) {
	body, err := os.ReadFile(filepath.Join("archetypes", name+".md"))
	switch {
	case err == nil:
		tpl, perr := template.New(name).Parse(string(body))
		if perr != nil {
			return "", fmt.Errorf("parse archetype %s: %w", name, perr)
		}
		var sb strings.Builder
		if eerr := tpl.Execute(&sb, data); eerr != nil {
			return "", fmt.Errorf("render archetype %s: %w", name, eerr)
		}
		return sb.String(), nil
	case os.IsNotExist(err):
		// fall through to built-in defaults
	default:
		return "", fmt.Errorf("read archetype %s: %w", name, err)
	}

	switch name {
	case "post":
		return fmt.Sprintf("---\ntitle: %q\ndate: %s\ndraft: true\ntags: []\ncategories: []\n---\n\n", data.Title, data.Date), nil
	case "page":
		return fmt.Sprintf("---\ntitle: %q\ndate: %s\ndraft: true\n---\n\n", data.Title, data.Date), nil
	default:
		return "", fmt.Errorf("no archetype %q found under archetypes/ and no built-in default exists; create archetypes/%s.md to define one", name, name)
	}
}

// slugToTitle converts a hyphen/underscore-separated slug to a title-cased string.
func slugToTitle(slug string) string {
	slug = strings.ReplaceAll(slug, "-", " ")
	slug = strings.ReplaceAll(slug, "_", " ")
	words := strings.Fields(slug)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
