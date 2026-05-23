package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	if err := runInit([]string{dir}); err != nil {
		t.Fatalf("init: %v", err)
	}
	mustExist := []string{
		"config.yaml",
		"content/posts/.gitkeep",
		"content/pages/.gitkeep",
		"archetypes/post.md",
		"archetypes/page.md",
		"README.md",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
		}
	}
}

func TestRunInit_NonEmptyDirRefused(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runInit([]string{dir})
	if err == nil {
		t.Fatal("expected error for non-empty dir")
	}
}

func TestRunInit_ForceAllowsNonEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit([]string{"--force", dir}); err != nil {
		t.Fatalf("init --force: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); err != nil {
		t.Errorf("config.yaml missing after --force: %v", err)
	}
	// Pre-existing file is preserved.
	if _, err := os.Stat(filepath.Join(dir, "other")); err != nil {
		t.Errorf("pre-existing file removed: %v", err)
	}
}

func TestRunInit_ForceDoesNotClobberExistingFile(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfg, []byte("existing: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit([]string{"--force", dir}); err != nil {
		t.Fatalf("init --force: %v", err)
	}
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing: true\n" {
		t.Errorf("config.yaml was overwritten: %q", string(data))
	}
}
