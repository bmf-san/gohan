package processor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

func writeTaxFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writeTaxFile: %v", err)
	}
}

func TestLoadTaxonomyRegistry_Valid(t *testing.T) {
	dir := t.TempDir()
	writeTaxFile(t, dir, "tags.yaml", "- name: go\n  description: Go language\n- name: ssg\n")
	writeTaxFile(t, dir, "categories.yaml", "- name: tutorials\n")
	reg, err := LoadTaxonomyRegistry(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Tags) != 2 {
		t.Errorf("Tags: got %d, want 2", len(reg.Tags))
	}
	if reg.Tags[0].Name != "go" {
		t.Errorf("Tags[0].Name: got %q", reg.Tags[0].Name)
	}
	if len(reg.Categories) != 1 {
		t.Errorf("Categories: got %d, want 1", len(reg.Categories))
	}
}

func TestLoadTaxonomyRegistry_MissingFiles(t *testing.T) {
	dir := t.TempDir()
	reg, err := LoadTaxonomyRegistry(dir)
	if err != nil {
		t.Fatalf("unexpected error for missing files: %v", err)
	}
	if len(reg.Tags) != 0 || len(reg.Categories) != 0 {
		t.Error("expected empty registry for missing files")
	}
}

func TestLoadTaxonomyRegistry_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeTaxFile(t, dir, "tags.yaml", "invalid: {yaml: [unclosed")
	_, err := LoadTaxonomyRegistry(dir)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestValidateArticleTaxonomies_AllValid(t *testing.T) {
	reg := &model.TaxonomyRegistry{
		Tags:       []model.Taxonomy{{Name: "go"}, {Name: "ssg"}},
		Categories: []model.Taxonomy{{Name: "tutorials"}},
	}
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", []string{"go"}, []string{"tutorials"}, time.Time{})},
	}
	errs := ValidateArticleTaxonomies(articles, reg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateArticleTaxonomies_UnknownTag(t *testing.T) {
	reg := &model.TaxonomyRegistry{
		Tags: []model.Taxonomy{{Name: "go"}},
	}
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", []string{"unknown-tag"}, nil, time.Time{})},
	}
	errs := ValidateArticleTaxonomies(articles, reg)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateArticleTaxonomies_UnknownCategory(t *testing.T) {
	reg := &model.TaxonomyRegistry{
		Categories: []model.Taxonomy{{Name: "news"}},
	}
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", nil, []string{"unknown-cat"}, time.Time{})},
	}
	errs := ValidateArticleTaxonomies(articles, reg)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateOutputPaths_NoDuplicates(t *testing.T) {
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/a/index.html"},
		{Article: *testArticle("b.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/b/index.html"},
	}
	if errs := ValidateOutputPaths(articles); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateOutputPaths_Duplicate(t *testing.T) {
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/same/index.html"},
		{Article: *testArticle("b.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/same/index.html"},
	}
	errs := ValidateOutputPaths(articles)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateOutputPaths_MultipleCollisions(t *testing.T) {
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/same/index.html"},
		{Article: *testArticle("b.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/same/index.html"},
		{Article: *testArticle("c.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/other/index.html"},
		{Article: *testArticle("d.md", "", "", nil, nil, time.Time{}), OutputPath: "public/posts/other/index.html"},
	}
	errs := ValidateOutputPaths(articles)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestBuildTagIndex(t *testing.T) {
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", []string{"go", "ssg"}, nil, time.Time{})},
		{Article: *testArticle("b.md", "", "", []string{"go"}, nil, time.Time{})},
	}
	idx := BuildTagIndex(articles)
	if len(idx["go"]) != 2 {
		t.Errorf("tag go: got %d articles, want 2", len(idx["go"]))
	}
	if len(idx["ssg"]) != 1 {
		t.Errorf("tag ssg: got %d articles, want 1", len(idx["ssg"]))
	}
}

func TestBuildCategoryIndex(t *testing.T) {
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("a.md", "", "", nil, []string{"news", "web"}, time.Time{})},
		{Article: *testArticle("b.md", "", "", nil, []string{"news"}, time.Time{})},
	}
	idx := BuildCategoryIndex(articles)
	if len(idx["news"]) != 2 {
		t.Errorf("cat news: got %d articles, want 2", len(idx["news"]))
	}
	if len(idx["web"]) != 1 {
		t.Errorf("cat web: got %d articles, want 1", len(idx["web"]))
	}
}

// --- Locale-aware taxonomy tests ---

func TestLoadLocaleAwareTaxonomyRegistries_GlobalOnly(t *testing.T) {
	dir := t.TempDir()
	writeTaxFile(t, dir, "tags.yaml", "- name: go\n")
	writeTaxFile(t, dir, "categories.yaml", "- name: tutorials\n")

	regs, err := LoadLocaleAwareTaxonomyRegistries(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if regs[""] == nil || len(regs[""].Tags) != 1 {
		t.Error("global registry should have 1 tag")
	}
}

func TestLoadLocaleAwareTaxonomyRegistries_LocaleFile(t *testing.T) {
	dir := t.TempDir()
	writeTaxFile(t, dir, "tags.yaml", "- name: go\n")
	if err := os.MkdirAll(filepath.Join(dir, "ja"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTaxFile(t, filepath.Join(dir, "ja"), "tags.yaml", "- name: golang\n- name: 書評\n")

	regs, err := LoadLocaleAwareTaxonomyRegistries(dir, []string{"en", "ja"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "en" has no locale file → falls back to global
	if len(regs["en"].Tags) != 1 || regs["en"].Tags[0].Name != "go" {
		t.Errorf("en should fall back to global: got %v", regs["en"].Tags)
	}
	// "ja" has its own file
	if len(regs["ja"].Tags) != 2 {
		t.Errorf("ja should have 2 tags, got %d", len(regs["ja"].Tags))
	}
}

func TestLoadLocaleAwareTaxonomyRegistries_FallbackWhenNoFiles(t *testing.T) {
	dir := t.TempDir()
	regs, err := LoadLocaleAwareTaxonomyRegistries(dir, []string{"en", "ja"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, locale := range []string{"", "en", "ja"} {
		if regs[locale] == nil {
			t.Errorf("locale %q: registry should not be nil", locale)
			continue
		}
		if len(regs[locale].Tags) != 0 || len(regs[locale].Categories) != 0 {
			t.Errorf("locale %q: expected empty registry", locale)
		}
	}
}

func TestMergeTaxonomyRegistries_Deduplication(t *testing.T) {
	regs := map[string]*model.TaxonomyRegistry{
		"en": {Tags: []model.Taxonomy{{Name: "go"}, {Name: "arch"}}},
		"ja": {Tags: []model.Taxonomy{{Name: "go"}, {Name: "アーキテクチャ"}}},
	}
	merged := MergeTaxonomyRegistries(regs)
	if len(merged.Tags) != 3 {
		t.Errorf("expected 3 unique tags, got %d: %v", len(merged.Tags), merged.Tags)
	}
}

func TestValidateArticleTaxonomiesLocale_PerLocale(t *testing.T) {
	regs := map[string]*model.TaxonomyRegistry{
		"": {
			Tags:       []model.Taxonomy{{Name: "go"}},
			Categories: []model.Taxonomy{{Name: "tools"}},
		},
		"en": {
			Tags:       []model.Taxonomy{{Name: "go"}, {Name: "arch"}},
			Categories: []model.Taxonomy{{Name: "tools"}},
		},
		"ja": {
			Tags:       []model.Taxonomy{{Name: "go"}, {Name: "アーキテクチャ"}},
			Categories: []model.Taxonomy{{Name: "ツール"}},
		},
	}

	articles := []*model.ProcessedArticle{
		// EN article using EN tag → OK
		{Article: *testArticle("en/a.md", "", "", []string{"arch"}, []string{"tools"}, time.Time{}), Locale: "en"},
		// JA article using JA tag → OK
		{Article: *testArticle("ja/b.md", "", "", []string{"アーキテクチャ"}, []string{"ツール"}, time.Time{}), Locale: "ja"},
		// EN article using JA-only tag → error
		{Article: *testArticle("en/c.md", "", "", []string{"アーキテクチャ"}, nil, time.Time{}), Locale: "en"},
	}

	errs := ValidateArticleTaxonomiesLocale(articles, regs)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateArticleTaxonomiesLocale_FallbackToGlobal(t *testing.T) {
	regs := map[string]*model.TaxonomyRegistry{
		"": {Tags: []model.Taxonomy{{Name: "go"}}},
	}
	// article with locale "fr" — no "fr" key, falls back to ""
	articles := []*model.ProcessedArticle{
		{Article: *testArticle("fr/a.md", "", "", []string{"go"}, nil, time.Time{}), Locale: "fr"},
		{Article: *testArticle("fr/b.md", "", "", []string{"unknown"}, nil, time.Time{}), Locale: "fr"},
	}
	errs := ValidateArticleTaxonomiesLocale(articles, regs)
	if len(errs) != 1 {
		t.Errorf("expected 1 error (unknown tag), got %d: %v", len(errs), errs)
	}
}
