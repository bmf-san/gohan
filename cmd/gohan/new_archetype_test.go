package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNew_CustomArchetypeFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("archetypes", 0o755); err != nil {
		t.Fatal(err)
	}
	tpl := "---\ntitle: \"{{ .Title }}\"\ndate: {{ .Date }}\nslug: {{ .Slug }}\ntype: tutorial\n---\n\n## Overview\n"
	if err := os.WriteFile(filepath.Join("archetypes", "tutorial.md"), []byte(tpl), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runNew([]string{"--type=tutorial", "--title=Hello", "intro"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join("content", "tutorial", "intro.md"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, `title: "Hello"`) {
		t.Errorf("title not rendered:\n%s", s)
	}
	if !strings.Contains(s, "slug: intro") {
		t.Errorf("slug not rendered:\n%s", s)
	}
	if !strings.Contains(s, "## Overview") {
		t.Errorf("body not rendered:\n%s", s)
	}
}

func TestRunNew_ArchetypeFlagOverridesType(t *testing.T) {
	tmpDir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("archetypes", 0o755); err != nil {
		t.Fatal(err)
	}
	tpl := "---\ntitle: \"{{ .Title }}\"\nlayout: news\n---\n"
	if err := os.WriteFile(filepath.Join("archetypes", "news.md"), []byte(tpl), 0o644); err != nil {
		t.Fatal(err)
	}

	// --type defaults to post (content/posts) but archetype=news overrides the template.
	if err := runNew([]string{"--archetype=news", "breaking"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join("content", "posts", "breaking.md"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if !strings.Contains(string(data), "layout: news") {
		t.Errorf("expected archetype 'news' to be applied:\n%s", string(data))
	}
}

func TestRunNew_UnknownArchetypeErrors(t *testing.T) {
	tmpDir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	err := runNew([]string{"--archetype=does-not-exist", "x"})
	if err == nil {
		t.Fatal("expected error for missing archetype")
	}
}
