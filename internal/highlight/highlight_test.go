package highlight

import (
	"strings"
	"testing"
)

func TestHighlight_Go(t *testing.T) {
	h := New(DefaultConfig())
	code := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n"
	out, err := h.Highlight(code, "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<pre") {
		t.Error("expected <pre> in output")
	}
	if !strings.Contains(out, "fmt") {
		t.Error("expected source content in output")
	}
}

func TestHighlight_Python(t *testing.T) {
	h := New(DefaultConfig())
	code := "def hello():\n    print('hello')\n"
	out, err := h.Highlight(code, "python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "hello") {
		t.Error("expected source content in output")
	}
}

func TestHighlight_Shell(t *testing.T) {
	h := New(DefaultConfig())
	code := "#!/bin/bash\necho hello\n"
	out, err := h.Highlight(code, "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "echo") {
		t.Error("expected source content in output")
	}
}

func TestHighlight_UnknownLanguage(t *testing.T) {
	h := New(DefaultConfig())
	code := "some unknown lang code\n"
	out, err := h.Highlight(code, "totally-unknown-xyz")
	// Should not error; falls back to plain rendering
	if err != nil {
		t.Fatalf("unexpected error for unknown language: %v", err)
	}
	if !strings.Contains(out, "some unknown lang code") {
		t.Errorf("expected original code in fallback output, got: %s", out)
	}
}

func TestHighlight_EmptyLanguage(t *testing.T) {
	h := New(DefaultConfig())
	code := "hello world\n"
	out, err := h.Highlight(code, "")
	if err != nil {
		t.Fatalf("unexpected error for empty language: %v", err)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected original code in output, got: %s", out)
	}
}

func TestHighlight_LineNumbers(t *testing.T) {
	h := New(Config{Theme: "github", LineNumbers: true})
	code := "line one\nline two\n"
	out, err := h.Highlight(code, "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// chroma with line numbers wraps in a table; just verify content exists
	if !strings.Contains(out, "line one") {
		t.Errorf("expected source content in line-number output, got: %s", out)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Theme != "github" {
		t.Errorf("expected default theme github, got %s", cfg.Theme)
	}
	if cfg.LineNumbers {
		t.Error("expected default line numbers false")
	}
}
