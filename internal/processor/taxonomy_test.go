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
