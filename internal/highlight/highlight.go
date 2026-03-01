// Package highlight provides chroma-based syntax highlighting
// for fenced code blocks in goldmark-rendered Markdown.
package highlight

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Config holds syntax highlighting settings derived from config.yaml.
type Config struct {
	// Theme is a chroma style name (e.g. "github", "monokai", "dracula").
	// Defaults to "github".
	Theme string
	// LineNumbers enables line number display.
	LineNumbers bool
}

// DefaultConfig returns the default highlighting configuration.
func DefaultConfig() Config {
	return Config{Theme: "github", LineNumbers: false}
}

// Highlighter renders fenced code blocks with chroma.
type Highlighter struct {
	cfg Config
}

// New creates a Highlighter with the provided config.
func New(cfg Config) *Highlighter {
	return &Highlighter{cfg: cfg}
}

// Highlight returns a syntax-highlighted HTML fragment for the given code and language.
// If the language is unknown or empty, it falls back to a plain <pre><code> block.
func (h *Highlighter) Highlight(code, lang string) (string, error) {
	var lexer = lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get(h.cfg.Theme)
	if style == nil {
		style = styles.Fallback
	}

	opts := []chromahtml.Option{
		chromahtml.WithClasses(false), // inline styles for portability
	}
	if h.cfg.LineNumbers {
		opts = append(opts, chromahtml.WithLineNumbers(true))
	}
	formatter := chromahtml.New(opts...)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return plainBlock(code), fmt.Errorf("tokenise: %w", err)
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return plainBlock(code), fmt.Errorf("format: %w", err)
	}
	return buf.String(), nil
}

func plainBlock(code string) string {
	return "<pre><code>" + html.EscapeString(code) + "</code></pre>"
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Goldmark extension
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Extension returns a goldmark.Extender that replaces the default fenced-code-
// block renderer with chroma-based highlighting.
func (h *Highlighter) Extension() goldmark.Extender {
	return &chromaExtender{h: h}
}

type chromaExtender struct{ h *Highlighter }

func (e *chromaExtender) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&codeBlockRenderer{h: e.h}, 200),
		),
	)
}

// codeBlockRenderer replaces goldmark's default FencedCodeBlock renderer.
type codeBlockRenderer struct{ h *Highlighter }

func (r *codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *codeBlockRenderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.FencedCodeBlock)

	// Collect raw code
	var sb strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		sb.Write(line.Value(source))
	}

	lang := string(n.Language(source))
	highlighted, err := r.h.Highlight(sb.String(), lang)
	if err != nil {
		// fallback – already returns plain HTML from plainBlock
		_, _ = fmt.Fprint(w, highlighted)
		return ast.WalkSkipChildren, nil
	}
	_, _ = fmt.Fprint(w, highlighted)
	return ast.WalkSkipChildren, nil
}
