package highlight

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

// TestExtension_RegisterAndRender exercises Extension(), Extend(),
// RegisterFuncs() and renderFencedCodeBlock() via a real goldmark conversion.
func TestExtension_RegisterAndRender(t *testing.T) {
	h := New(DefaultConfig())
	md := goldmark.New(goldmark.WithExtensions(h.Extension()))

	src := []byte("```go\nfmt.Println(\"hello\")\n```\n")
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		t.Fatalf("goldmark.Convert: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<pre") {
		t.Errorf("expected <pre> tag in highlighted output, got:\n%s", out)
	}
	if !strings.Contains(out, "Println") {
		t.Errorf("expected source code content in output, got:\n%s", out)
	}
}

func TestExtension_WithLineNumbers(t *testing.T) {
	h := New(Config{Theme: "github", LineNumbers: true})
	md := goldmark.New(goldmark.WithExtensions(h.Extension()))

	src := []byte("```python\nprint('hi')\n```\n")
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		t.Fatalf("goldmark.Convert: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "print") {
		t.Errorf("expected code content in output, got:\n%s", out)
	}
}

func TestExtension_UnknownLanguage(t *testing.T) {
	h := New(DefaultConfig())
	md := goldmark.New(goldmark.WithExtensions(h.Extension()))

	src := []byte("```unknownxyz\nsome code\n```\n")
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		t.Fatalf("goldmark.Convert: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "some code") {
		t.Errorf("expected fallback content in output, got:\n%s", out)
	}
}

func TestExtension_NoLanguage(t *testing.T) {
	h := New(DefaultConfig())
	md := goldmark.New(goldmark.WithExtensions(h.Extension()))

	src := []byte("```\nhello world\n```\n")
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		t.Fatalf("goldmark.Convert: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected content in output, got:\n%s", out)
	}
}

func TestPlainBlock_EscapesHTML(t *testing.T) {
	out := plainBlock("<script>alert('xss')</script>")
	if strings.Contains(out, "<script>") {
		t.Error("expected HTML-escaped output, got raw <script>")
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Errorf("expected &lt;script&gt; in output, got:\n%s", out)
	}
	if !strings.Contains(out, "<pre><code>") {
		t.Errorf("expected <pre><code> wrapper, got:\n%s", out)
	}
}
