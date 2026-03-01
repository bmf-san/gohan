package mermaid

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkparser "github.com/yuin/goldmark/parser"
)

func newMD() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM, Extension()),
		goldmark.WithParserOptions(goldmarkparser.WithAutoHeadingID()),
	)
}

func TestMermaidBlock_Passthrough(t *testing.T) {
	md := newMD()
	src := "```mermaid\ngraph TD\n  A --> B\n```\n"
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `class="mermaid"`) {
		t.Errorf("expected class=mermaid in output, got:\n%s", out)
	}
	if !strings.Contains(out, "A --&gt; B") || !strings.Contains(out, "graph TD") {
		// HTML-escaped content expected
		t.Errorf("expected diagram content in output, got:\n%s", out)
	}
	if strings.Contains(out, "<pre>") {
		t.Errorf("expected <div>, not <pre> for mermaid block, got:\n%s", out)
	}
}

func TestMermaidBlock_NotAffectOtherLang(t *testing.T) {
	md := newMD()
	src := "```go\npackage main\n```\n"
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `class="mermaid"`) {
		t.Errorf("go block should not be wrapped in mermaid div, got:\n%s", out)
	}
}

func TestInjectScript_WithBody(t *testing.T) {
	input := []byte("<html><body>content</body></html>")
	out := InjectScript(input)
	if !bytes.Contains(out, []byte(ScriptTag)) {
		t.Error("expected script tag in output")
	}
	if !bytes.Contains(out, []byte("content")) {
		t.Error("original content must be preserved")
	}
	// Script should appear before </body>
	scriptIdx := bytes.Index(out, []byte(ScriptTag))
	bodyIdx := bytes.Index(out, []byte("</body>"))
	if scriptIdx > bodyIdx {
		t.Error("script should be injected before </body>")
	}
}

func TestInjectScript_NoBody(t *testing.T) {
	input := []byte("<html>no body tag</html>")
	out := InjectScript(input)
	if !bytes.Contains(out, []byte(ScriptTag)) {
		t.Error("expected script appended when no </body>")
	}
}

func TestMermaidMarker(t *testing.T) {
	if MermaidMarker == "" {
		t.Error("MermaidMarker must not be empty")
	}
}
