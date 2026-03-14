package main

import (
	"flag"
	"fmt"
	"io"
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

	// Validate slug: no path separators, no whitespace
	if strings.ContainsAny(slug, "/\\ \t") {
		return fmt.Errorf("slug must not contain path separators or whitespace: %q", slug)
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

	// BUG-D: use O_CREATE|O_EXCL for an atomic create-or-fail, eliminating the
	// TOCTOU race between the old os.Stat check and os.WriteFile.
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
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
