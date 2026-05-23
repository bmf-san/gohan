package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/bmf-san/gohan/internal/config"
	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/parser"
)

// runCheck implements `gohan check`, a lightweight linter that validates
// content without performing a full build. It reports:
//   - Duplicate slugs within the same output directory.
//   - Articles missing required front matter (currently: title and date).
//   - translation_key values that only have a single article (no actual
//     translation pair).
//
// Exit code is 0 when no problems are found, 1 when warnings are reported.
// Pure warnings policy: this command never modifies the filesystem.
func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "Path to config file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	cfgAbs := *configPath
	if !filepath.IsAbs(cfgAbs) {
		cfgAbs = filepath.Join(rootDir, cfgAbs)
	}
	if _, err := os.Stat(cfgAbs); err != nil {
		return fmt.Errorf("config file not found: %s", cfgAbs)
	}

	loader := config.New(rootDir)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	contentDir := filepath.Join(rootDir, cfg.Build.ContentDir)
	p := parser.NewFileParser(cfg.Build.ExcludeFiles...)
	articles, err := p.ParseAll(contentDir)
	if err != nil {
		return fmt.Errorf("parse content: %w", err)
	}

	issues := lintArticles(articles, contentDir)
	writeCheckReport(os.Stdout, issues)
	if len(issues) > 0 {
		return fmt.Errorf("%d issue(s) found", len(issues))
	}
	return nil
}

// checkIssue is a single linter finding.
type checkIssue struct {
	File    string // relative path under contentDir
	Kind    string // "missing-title" | "missing-date" | "duplicate-slug" | "orphan-translation-key"
	Message string
}

func lintArticles(articles []*model.Article, contentDir string) []checkIssue {
	var issues []checkIssue

	// Group by directory + slug to detect duplicates.
	type key struct{ dir, slug string }
	bySlug := make(map[key][]string)

	// translation_key occurrence counts.
	tkCount := make(map[string][]string)

	for _, a := range articles {
		rel, err := filepath.Rel(contentDir, a.FilePath)
		if err != nil {
			rel = a.FilePath
		}

		if a.FrontMatter.Title == "" {
			issues = append(issues, checkIssue{
				File:    rel,
				Kind:    "missing-title",
				Message: "front matter is missing required field 'title'",
			})
		}
		if a.FrontMatter.Date.IsZero() {
			issues = append(issues, checkIssue{
				File:    rel,
				Kind:    "missing-date",
				Message: "front matter is missing required field 'date'",
			})
		}

		slug := a.FrontMatter.Slug
		if slug == "" {
			// Fall back to filename without extension, mirroring the generator.
			base := filepath.Base(a.FilePath)
			slug = base[:len(base)-len(filepath.Ext(base))]
		}
		dir := filepath.Dir(rel)
		k := key{dir: dir, slug: slug}
		bySlug[k] = append(bySlug[k], rel)

		if tk := a.FrontMatter.TranslationKey; tk != "" {
			tkCount[tk] = append(tkCount[tk], rel)
		}
	}

	// Report duplicate slugs.
	dupKeys := make([]key, 0)
	for k, files := range bySlug {
		if len(files) > 1 {
			dupKeys = append(dupKeys, k)
		}
	}
	sort.Slice(dupKeys, func(i, j int) bool {
		if dupKeys[i].dir != dupKeys[j].dir {
			return dupKeys[i].dir < dupKeys[j].dir
		}
		return dupKeys[i].slug < dupKeys[j].slug
	})
	for _, k := range dupKeys {
		files := bySlug[k]
		sort.Strings(files)
		issues = append(issues, checkIssue{
			File:    files[0],
			Kind:    "duplicate-slug",
			Message: fmt.Sprintf("duplicate slug %q in %q also used by: %v", k.slug, k.dir, files[1:]),
		})
	}

	// Report orphan translation keys (key referenced by only one article).
	orphanKeys := make([]string, 0)
	for tk, files := range tkCount {
		if len(files) == 1 {
			orphanKeys = append(orphanKeys, tk)
		}
	}
	sort.Strings(orphanKeys)
	for _, tk := range orphanKeys {
		issues = append(issues, checkIssue{
			File:    tkCount[tk][0],
			Kind:    "orphan-translation-key",
			Message: fmt.Sprintf("translation_key %q has no matching translation in other locales", tk),
		})
	}

	return issues
}

func writeCheckReport(w io.Writer, issues []checkIssue) {
	if len(issues) == 0 {
		_, _ = fmt.Fprintln(w, "check: no issues found")
		return
	}
	// Stable order: by file, then kind.
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].File != issues[j].File {
			return issues[i].File < issues[j].File
		}
		return issues[i].Kind < issues[j].Kind
	})
	for _, it := range issues {
		_, _ = fmt.Fprintf(w, "%s: [%s] %s\n", it.File, it.Kind, it.Message)
	}
	_, _ = fmt.Fprintf(w, "\ncheck: %d issue(s)\n", len(issues))
}
