package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// runInit scaffolds a new gohan project directory containing a minimal
// `config.yaml`, the standard content folders, and starter archetypes.
//
// Usage:
//
//	gohan init [--force] [<dir>]
//
// When <dir> is omitted the current working directory is used. The command
// refuses to write into a non-empty directory unless --force is given.
func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "allow scaffolding into a non-empty directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	target := "."
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve target dir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	if !*force {
		entries, readErr := os.ReadDir(abs)
		if readErr != nil {
			return fmt.Errorf("read target dir: %w", readErr)
		}
		if len(entries) > 0 {
			return fmt.Errorf("target dir %q is not empty; pass --force to scaffold anyway", target)
		}
	}

	files := map[string]string{
		"config.yaml":            initConfigYAML,
		"content/posts/.gitkeep": "",
		"content/pages/.gitkeep": "",
		"archetypes/post.md":     initArchetypePost,
		"archetypes/page.md":     initArchetypePage,
		"README.md":              initReadme,
	}

	created := make([]string, 0, len(files))
	for rel, body := range files {
		path := filepath.Join(abs, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(rel), err)
		}
		// O_EXCL prevents accidentally clobbering existing files when --force is set.
		f, openErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if openErr != nil {
			if os.IsExist(openErr) {
				continue
			}
			return fmt.Errorf("create %s: %w", rel, openErr)
		}
		if _, werr := io.WriteString(f, body); werr != nil {
			_ = f.Close()
			return fmt.Errorf("write %s: %w", rel, werr)
		}
		if cerr := f.Close(); cerr != nil {
			return fmt.Errorf("close %s: %w", rel, cerr)
		}
		created = append(created, rel)
	}

	fmt.Printf("initialised gohan project at %s\n", abs)
	for _, rel := range created {
		fmt.Printf("  + %s\n", rel)
	}
	fmt.Println("next steps:")
	fmt.Println("  1. install a theme into themes/<name>/ (see https://github.com/bmf-san/gohan)")
	fmt.Println("  2. set theme.dir in config.yaml")
	fmt.Println("  3. run `gohan new my-first-post`")
	fmt.Println("  4. run `gohan build`")
	return nil
}

const initConfigYAML = `site:
  title: My Gohan Site
  description: A new site built with gohan.
  base_url: http://localhost:8080
  language: en

build:
  content_dir: content
  output_dir: public
  assets_dir: assets
  static_dir: static

theme:
  dir: themes/default

i18n:
  default_locale: en
  locales: [en]
`

const initArchetypePost = `---
title: {{ .Title }}
date: {{ .Date }}
draft: true
tags: []
categories: []
---

`

const initArchetypePage = `---
title: {{ .Title }}
date: {{ .Date }}
draft: true
---

`

const initReadme = `# My Gohan Site

This site was scaffolded by ` + "`gohan init`" + `.

## Quick start

1. Install a theme into ` + "`themes/<name>/`" + ` and update ` + "`theme.dir`" + ` in ` + "`config.yaml`" + `.
2. Create a new post: ` + "`gohan new my-first-post`" + `.
3. Build the site: ` + "`gohan build`" + `.
4. Preview locally: ` + "`gohan serve`" + `.

See https://github.com/bmf-san/gohan for documentation.
`
