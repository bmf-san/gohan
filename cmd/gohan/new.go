package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	title := fs.String("title", "", "article title (defaults to slug)")
	articleType := fs.String("type", "post", "article type: post or page")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: gohan new [--title=\"...\" ] [--type=post|page] <slug>")
	}
	slug := fs.Arg(0)

	// Validate slug: no path separators, no spaces
	if strings.ContainsAny(slug, "/\\") {
		return fmt.Errorf("slug must not contain path separators: %q", slug)
	}

	// Determine directory
	var dir string
	switch *articleType {
	case "post", "":
		dir = filepath.Join("content", "posts")
	case "page":
		dir = filepath.Join("content", "pages")
	default:
		return fmt.Errorf("unknown type %q: must be post or page", *articleType)
	}

	// Resolve title
	articleTitle := *title
	if articleTitle == "" {
		// Convert slug to title: replace hyphens/underscores with spaces, title-case
		articleTitle = slugToTitle(slug)
	}

	filePath := filepath.Join(dir, slug+".md")

	// Error if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file already exists: %s", filePath)
	}

	// Create directory if needed
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Build front matter
	today := time.Now().Format("2006-01-02")
	content := fmt.Sprintf(`---
title: %q
date: %s
draft: true
tags: []
categories: []
---

`, articleTitle, today)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("created: %s\n", filePath)
	return nil
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
