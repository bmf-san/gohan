// Package mermaid provides a goldmark extension that transforms fenced code
// blocks tagged with "mermaid" into browser-renderable <div class="mermaid">
// elements for client-side rendering via the Mermaid.js CDN.
package mermaid

import (
	"bytes"
	"html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// ScriptTag is the Mermaid CDN <script> snippet injected into HTML pages that
// contain at least one mermaid diagram block.
const ScriptTag = `<script type="module">` +
	`import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs';` +
	`mermaid.initialize({startOnLoad:true});` +
	`</script>`

// MermaidMarker is the CSS class used to identify mermaid diagrams in HTML.
const MermaidMarker = `class="mermaid"`

// InjectScript inserts ScriptTag into body before </body>.
// It returns the modified HTML; if </body> is not present the script is appended.
// Callers should only call this when the HTML contains MermaidMarker.
func InjectScript(htmlDoc []byte) []byte {
	script := []byte(ScriptTag)
	if idx := bytes.Index(htmlDoc, []byte("</body>")); idx >= 0 {
		out := make([]byte, 0, len(htmlDoc)+len(script))
		out = append(out, htmlDoc[:idx]...)
		out = append(out, script...)
		out = append(out, htmlDoc[idx:]...)
		return out
	}
	return append(htmlDoc, script...)
}

// Extension returns a goldmark.Extender that replaces the FencedCodeBlock
// renderer for blocks whose info/language string is "mermaid".
func Extension() goldmark.Extender {
	return &mermaidExtender{}
}

type mermaidExtender struct{}

func (e *mermaidExtender) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&mermaidRenderer{}, 199), // higher priority than chroma (200)
		),
	)
}

// mermaidRenderer handles FencedCodeBlock nodes with lang == "mermaid".
// All other code blocks are passed to the default renderer.
type mermaidRenderer struct{}

func (r *mermaidRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *mermaidRenderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.FencedCodeBlock)
	lang := string(n.Language(source))
	if lang != "mermaid" {
		// Not a mermaid block — let the next renderer handle it.
		return ast.WalkContinue, nil
	}

	// Collect diagram source
	var buf bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.Write(line.Value(source))
	}

	// Output as <div class="mermaid"> for client-side rendering
	_, _ = w.WriteString(`<div class="mermaid">`)
	_, _ = w.WriteString(html.EscapeString(buf.String()))
	_, _ = w.WriteString(`</div>`)
	return ast.WalkSkipChildren, nil
}
