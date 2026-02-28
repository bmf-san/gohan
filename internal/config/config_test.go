package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmf-san/gohan/internal/config"
)

// writeConfig writes content to config.yaml inside dir and returns dir.
func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeConfig: %v", err)
	}
	return dir
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
site:
  title: "My Blog"
  description: "A personal blog"
  base_url: "https://example.com"
  language: "en"
build:
  content_dir: "content"
  output_dir: "public"
  assets_dir: "assets"
  parallelism: 4
theme:
  name: "default"
  dir: "themes/default"
`)

	cfg, err := config.New(dir).Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Site.Title != "My Blog" {
		t.Errorf("site.title: got %q, want %q", cfg.Site.Title, "My Blog")
	}
	if cfg.Site.BaseURL != "https://example.com" {
		t.Errorf("site.base_url: got %q, want %q", cfg.Site.BaseURL, "https://example.com")
	}
	if cfg.Build.Parallelism != 4 {
		t.Errorf("build.parallelism: got %d, want 4", cfg.Build.Parallelism)
	}
	if cfg.Theme.Name != "default" {
		t.Errorf("theme.name: got %q, want %q", cfg.Theme.Name, "default")
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
site:
  title: "My Blog"
  base_url: "https://example.com"
`)

	cfg, err := config.New(dir).Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Build.ContentDir != "content" {
		t.Errorf("build.content_dir default: got %q, want %q", cfg.Build.ContentDir, "content")
	}
	if cfg.Build.OutputDir != "public" {
		t.Errorf("build.output_dir default: got %q, want %q", cfg.Build.OutputDir, "public")
	}
	if cfg.Build.AssetsDir != "assets" {
		t.Errorf("build.assets_dir default: got %q, want %q", cfg.Build.AssetsDir, "assets")
	}
	if cfg.Build.Parallelism != 4 {
		t.Errorf("build.parallelism default: got %d, want 4", cfg.Build.Parallelism)
	}
	if cfg.Theme.Name != "default" {
		t.Errorf("theme.name default: got %q, want %q", cfg.Theme.Name, "default")
	}
	if cfg.Theme.Dir != "themes/default" {
		t.Errorf("theme.dir default: got %q, want %q", cfg.Theme.Dir, "themes/default")
	}
	if cfg.Site.Language != "en" {
		t.Errorf("site.language default: got %q, want %q", cfg.Site.Language, "en")
	}
}

func TestLoad_MissingTitle(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
site:
  base_url: "https://example.com"
`)

	_, err := config.New(dir).Load()
	if err == nil {
		t.Fatal("expected error for missing site.title, got nil")
	}
}

func TestLoad_MissingBaseURL(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
site:
  title: "My Blog"
`)

	_, err := config.New(dir).Load()
	if err == nil {
		t.Fatal("expected error for missing site.base_url, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	dir := t.TempDir() // no config.yaml written

	_, err := config.New(dir).Load()
	if err == nil {
		t.Fatal("expected error for missing config.yaml, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `this: is: not: valid: yaml: [`)

	_, err := config.New(dir).Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_CustomThemeDir(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
site:
  title: "My Blog"
  base_url: "https://example.com"
theme:
  name: "mytheme"
  dir: "custom/theme/path"
`)

	cfg, err := config.New(dir).Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme.Dir != "custom/theme/path" {
		t.Errorf("theme.dir: got %q, want %q", cfg.Theme.Dir, "custom/theme/path")
	}
}

func TestLoad_ThemeDirDefaultsDerivedFromName(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
site:
  title: "My Blog"
  base_url: "https://example.com"
theme:
  name: "mytheme"
`)

	cfg, err := config.New(dir).Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme.Dir != "themes/mytheme" {
		t.Errorf("theme.dir: got %q, want %q", cfg.Theme.Dir, "themes/mytheme")
	}
}
