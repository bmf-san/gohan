package parser

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/bmf-san/gohan/internal/highlight"
	"github.com/bmf-san/gohan/internal/mermaid"
)

// Converter converts Markdown source bytes into safe HTML.
// It is CommonMark-compliant and supports GitHub Flavored Markdown (GFM)
// extensions by default.
type Converter struct {
	md goldmark.Markdown
}

// converterConfig holds the options accumulated before building the Markdown engine.
type converterConfig struct {
	gfm          bool
	unsafeHTML   bool
	highlighting *highlight.Config
	mermaid      bool
}

// ConverterOption is a functional option for NewConverter.
type ConverterOption func(*converterConfig)

// WithGFM enables the GitHub Flavored Markdown extension set (tables,
// strikethrough, task lists, and auto-links).  Enabled by default via
// NewConverter.
func WithGFM() ConverterOption {
	return func(c *converterConfig) { c.gfm = true }
}

// WithUnsafeHTML allows raw HTML pass-through in Markdown source.  By
// default raw HTML is escaped.  Use only when the content source is trusted.
func WithUnsafeHTML() ConverterOption {
	return func(c *converterConfig) { c.unsafeHTML = true }
}

// WithHighlighting enables chroma syntax highlighting for fenced code blocks.
// cfg specifies the chroma theme and line-number settings.
func WithHighlighting(cfg highlight.Config) ConverterOption {
	return func(c *converterConfig) { c.highlighting = &cfg }
}

// WithMermaid enables client-side Mermaid diagram rendering.
// Fenced code blocks tagged "mermaid" are output as <div class="mermaid">.
func WithMermaid() ConverterOption {
	return func(c *converterConfig) { c.mermaid = true }
}

// NewConverter builds a Converter with the supplied options.  When no options
// are given, GFM extensions are enabled and raw HTML is escaped.
func NewConverter(opts ...ConverterOption) *Converter {
	cfg := &converterConfig{gfm: true}
	for _, o := range opts {
		o(cfg)
	}

	var mdOpts []goldmark.Option

	if cfg.gfm {
		mdOpts = append(mdOpts,
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(
				goldmarkparser.WithAutoHeadingID(),
			),
		)
	}

	if cfg.unsafeHTML {
		mdOpts = append(mdOpts, goldmark.WithRendererOptions(html.WithUnsafe()))
	}

	if cfg.highlighting != nil {
		h := highlight.New(*cfg.highlighting)
		mdOpts = append(mdOpts, goldmark.WithExtensions(h.Extension()))
	}

	if cfg.mermaid {
		mdOpts = append(mdOpts, goldmark.WithExtensions(mermaid.Extension()))
	}

	return &Converter{md: goldmark.New(mdOpts...)}
}

// Convert transforms src Markdown bytes into an HTML string.  The returned
// value is marked safe for use with html/template without additional escaping.
func (c *Converter) Convert(src []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := c.md.Convert(src, &buf); err != nil {
		return "", fmt.Errorf("markdown: convert: %w", err)
	}
	return template.HTML(buf.String()), nil //nolint:gosec // goldmark output is safe HTML
}
