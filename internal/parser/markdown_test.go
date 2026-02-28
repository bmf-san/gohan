package parser_test

import (
	"strings"
	"testing"

	"github.com/bmf-san/gohan/internal/parser"
)

func TestConverter_Heading(t *testing.T) {
	c := parser.NewConverter()
	got, err := c.Convert([]byte("# Hello\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "Hello") {
		t.Errorf("expected Hello in output, got: %s", got)
	}
}

func TestConverter_Paragraph(t *testing.T) {
	c := parser.NewConverter()
	got, err := c.Convert([]byte("Hello, world.\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "<p>Hello, world.</p>") {
		t.Errorf("expected paragraph, got: %s", got)
	}
}

func TestConverter_Link(t *testing.T) {
	c := parser.NewConverter()
	got, err := c.Convert([]byte("[Go](https://go.dev)\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "https://go.dev") {
		t.Errorf("expected link href, got: %s", html)
	}
}

func TestConverter_CodeBlock(t *testing.T) {
	c := parser.NewConverter()
	src := "```go\nfmt.Println()\n```\n"
	got, err := c.Convert([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "<pre><code") {
		t.Errorf("expected code block, got: %s", got)
	}
}

func TestConverter_GFM_Strikethrough(t *testing.T) {
	c := parser.NewConverter()
	got, err := c.Convert([]byte("~~strike~~\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "<del>strike</del>") {
		t.Errorf("expected strikethrough, got: %s", got)
	}
}

func TestConverter_GFM_Table(t *testing.T) {
	c := parser.NewConverter()
	src := "| A | B |\n|---|---|\n| 1 | 2 |\n"
	got, err := c.Convert([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "<table>") {
		t.Errorf("expected table, got: %s", html)
	}
}
func TestConverter_HTMLEscapedByDefault(t *testing.T) {
	c := parser.NewConverter()
	src := "<script>alert(1)</script>\n"
	got, err := c.Convert([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(string(got), "<script>") {
		t.Errorf("raw script tag should be escaped, got: %s", got)
	}
}

func TestConverter_UnsafeHTMLPassthrough(t *testing.T) {
	c := parser.NewConverter(parser.WithUnsafeHTML())
	src := "<b>raw</b>\n"
	got, err := c.Convert([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "<b>raw</b>") {
		t.Errorf("raw HTML not passed through, got: %s", got)
	}
}

func TestConverter_WithGFMOption(t *testing.T) {
	c := parser.NewConverter(parser.WithGFM())
	got, err := c.Convert([]byte("~~strike~~\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "<del>") {
		t.Errorf("WithGFM option: expected strikethrough, got: %s", got)
	}
}

func TestConverter_EmptyInput(t *testing.T) {
	c := parser.NewConverter()
	got, err := c.Convert([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(string(got)) != "" {
		t.Errorf("expected empty output, got: %q", got)
	}
}

func TestConverter_HeadingAutoID(t *testing.T) {
	c := parser.NewConverter()
	got, err := c.Convert([]byte("## My Section\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "id=") {
		t.Errorf("expected id on heading, got: %s", got)
	}
}
