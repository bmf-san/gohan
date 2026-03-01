package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNew_MissingSlug(t *testing.T) {
	err := runNew([]string{})
	if err == nil {
		t.Fatal("expected error for missing slug")
	}
}

func TestRunNew_UnknownType(t *testing.T) {
	err := runNew([]string{"--type=article", "my-slug"})
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestRunNew_CreatePost(t *testing.T) {
	tmpDir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	err := runNew([]string{"my-first-post"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join("content", "posts", "my-first-post.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "title:") {
		t.Error("missing title in front matter")
	}
	if !strings.Contains(content, "draft: true") {
		t.Error("missing draft in front matter")
	}
	if !strings.Contains(content, "tags: []") {
		t.Error("missing tags in front matter")
	}
	if !strings.Contains(content, "categories: []") {
		t.Error("missing categories in front matter")
	}
}

func TestRunNew_CreatePage(t *testing.T) {
	tmpDir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	err := runNew([]string{"--type=page", "--title=About Me", "about"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join("content", "pages", "about.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if !strings.Contains(string(data), `"About Me"`) {
		t.Errorf("expected title \"About Me\" in front matter, got:\n%s", string(data))
	}
}

func TestRunNew_ExistingFileError(t *testing.T) {
	tmpDir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(tmpDir)

	// Create first time
	if err := runNew([]string{"duplicate-slug"}); err != nil {
		t.Fatalf("first creation failed: %v", err)
	}
	// Second creation should fail
	if err := runNew([]string{"duplicate-slug"}); err == nil {
		t.Fatal("expected error for existing file")
	}
}

func TestRunNew_SlugToTitle(t *testing.T) {
	cases := []struct {
		slug, want string
	}{
		{"my-first-post", "My First Post"},
		{"hello_world", "Hello World"},
		{"simple", "Simple"},
	}
	for _, tc := range cases {
		got := slugToTitle(tc.slug)
		if got != tc.want {
			t.Errorf("slugToTitle(%q) = %q, want %q", tc.slug, got, tc.want)
		}
	}
}
