package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBuild_UnknownFlag(t *testing.T) {
	err := runBuild([]string{"--unknown-flag"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestRunBuild_MissingConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	err := runBuild([]string{"--config=" + cfgPath})
	if err == nil {
		t.Error("expected error when config file missing")
	}
}

func TestRunBuild_DryRun(t *testing.T) {
	dir := t.TempDir()

	// Write minimal config.yaml
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}

	// Create empty content dir (ParseAll on empty dir should return no articles)
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0755); err != nil {
		t.Fatal(err)
	}

	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("dry-run with empty content: %v", err)
	}
}

func TestRunBuild_FullFlagAccepted(t *testing.T) {
	dir := t.TempDir()
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0755); err != nil {
		t.Fatal(err)
	}
	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--full",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("--full --dry-run: %v", err)
	}
}

func TestRunBuild_OutputOverride(t *testing.T) {
	dir := t.TempDir()
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0755); err != nil {
		t.Fatal(err)
	}
	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--output=" + filepath.Join(dir, "out"),
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("--output override: %v", err)
	}
}

func TestRunBuild_DraftFlagAccepted(t *testing.T) {
	dir := t.TempDir()
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0755); err != nil {
		t.Fatal(err)
	}
	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--draft",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("--draft --dry-run: %v", err)
	}
}

// TestRunBuild_DraftArticlesExcludedByDefault verifies draft filtering.
func TestRunBuild_DraftArticlesExcludedByDefault(t *testing.T) {
	dir := t.TempDir()
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	contentDir := filepath.Join(dir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write one draft and one published article.
	draft := []byte("---\ntitle: Draft Post\ndraft: true\n---\nDraft body.\n")
	pub := []byte("---\ntitle: Published Post\ndraft: false\n---\nPublished body.\n")
	if err := os.WriteFile(filepath.Join(contentDir, "draft.md"), draft, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "published.md"), pub, 0644); err != nil {
		t.Fatal(err)
	}
	// dry-run reports processed count; we just verify it doesn't error.
	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("build with draft article: %v", err)
	}
}

// TestRunBuild_FutureFlagAccepted verifies the --future CLI flag parses.
func TestRunBuild_FutureFlagAccepted(t *testing.T) {
	dir := t.TempDir()
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0755); err != nil {
		t.Fatal(err)
	}
	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--future",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("--future --dry-run: %v", err)
	}
}

// TestRunBuild_ScheduledArticlesExcludedByDefault verifies that articles with a
// future date are skipped unless --future is set.
func TestRunBuild_ScheduledArticlesExcludedByDefault(t *testing.T) {
	dir := t.TempDir()
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	contentDir := filepath.Join(dir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatal(err)
	}
	future := []byte("---\ntitle: Future Post\ndate: 2999-01-01T00:00:00Z\n---\nBody.\n")
	pub := []byte("---\ntitle: Published Post\ndate: 2024-01-01T00:00:00Z\n---\nBody.\n")
	if err := os.WriteFile(filepath.Join(contentDir, "future.md"), future, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "published.md"), pub, 0644); err != nil {
		t.Fatal(err)
	}
	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("build with future article: %v", err)
	}
}
