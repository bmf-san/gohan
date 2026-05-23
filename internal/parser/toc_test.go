package parser

import (
	"testing"
)

func TestExtractTOC_FlatHeadings(t *testing.T) {
	src := []byte("# A\n\ntext\n\n# B\n\nmore text\n")
	got := ExtractTOC(src)
	if len(got) != 2 {
		t.Fatalf("got %d roots, want 2", len(got))
	}
	if got[0].Text != "A" || got[0].Level != 1 || got[0].ID != "a" {
		t.Errorf("first entry = %+v", got[0])
	}
	if got[1].Text != "B" || got[1].Level != 1 || got[1].ID != "b" {
		t.Errorf("second entry = %+v", got[1])
	}
}

func TestExtractTOC_Nested(t *testing.T) {
	src := []byte("# A\n\n## A1\n\n### A1a\n\n## A2\n\n# B\n")
	got := ExtractTOC(src)
	if len(got) != 2 {
		t.Fatalf("got %d roots, want 2", len(got))
	}
	a := got[0]
	if a.Text != "A" {
		t.Fatalf("root[0] text = %q", a.Text)
	}
	if len(a.Children) != 2 {
		t.Fatalf("A children = %d, want 2", len(a.Children))
	}
	if a.Children[0].Text != "A1" {
		t.Errorf("A.children[0] = %q", a.Children[0].Text)
	}
	if len(a.Children[0].Children) != 1 || a.Children[0].Children[0].Text != "A1a" {
		t.Errorf("A1 children = %+v", a.Children[0].Children)
	}
	if a.Children[1].Text != "A2" {
		t.Errorf("A.children[1] = %q", a.Children[1].Text)
	}
}

func TestExtractTOC_EmptyDocument(t *testing.T) {
	if got := ExtractTOC([]byte("just a paragraph\n")); got != nil {
		t.Errorf("want nil, got %+v", got)
	}
}

func TestExtractTOC_StartsAtDeepLevel(t *testing.T) {
	src := []byte("### deep\n\n## mid\n")
	got := ExtractTOC(src)
	if len(got) != 2 {
		t.Fatalf("want 2 roots, got %d: %+v", len(got), got)
	}
	if got[0].Level != 3 || got[1].Level != 2 {
		t.Errorf("levels = %d, %d", got[0].Level, got[1].Level)
	}
}

func TestExtractTOC_InlineFormatting(t *testing.T) {
	src := []byte("# Hello *world* `code`\n")
	got := ExtractTOC(src)
	if len(got) != 1 {
		t.Fatalf("want 1, got %d", len(got))
	}
	// Text content should include all inline text segments.
	if got[0].Text == "" {
		t.Errorf("text is empty")
	}
}
