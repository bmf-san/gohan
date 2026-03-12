// Package mermaid provides a goldmark extension that transforms fenced code
// blocks tagged with "mermaid" into browser-renderable <div class="mermaid">
// elements for client-side rendering via the Mermaid.js CDN.
//
// Implementation note: instead of registering a renderer for
// ast.KindFencedCodeBlock (which would conflict with chroma or other
// code-block renderers), this extension uses an AST transformer that replaces
// mermaid FencedCodeBlock nodes with a custom KindMermaidBlock node before
// rendering.  The renderer is then registered only for KindMermaidBlock,
// leaving KindFencedCodeBlock untouched for other extensions to handle.
package mermaid

import (
	"bytes"
	"html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	goldmarkparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
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

// KindMermaidBlock is the AST node kind for mermaid diagram blocks.
var KindMermaidBlock = ast.NewNodeKind("MermaidBlock")

// mermaidNode is a custom AST node that holds the pre-collected mermaid source.
type mermaidNode struct {
	ast.BaseBlock
	source string
}

func (n *mermaidNode) Kind() ast.NodeKind { return KindMermaidBlock }
func (n *mermaidNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

// mermaidTransformer walks the parsed AST and replaces every FencedCodeBlock
// whose language is "mermaid" with a mermaidNode before rendering.
// This avoids registering for KindFencedCodeBlock and conflicting with
// syntax-highlighting extensions such as chroma.
type mermaidTransformer struct{}

func (t *mermaidTransformer) Transform(doc *ast.Document, reader text.Reader, _ goldmarkparser.Context) {
	src := reader.Source()

	// Collect targets first; do not modify the tree while walking.
	var targets []*ast.FencedCodeBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		fcb, ok := n.(*ast.FencedCodeBlock)
		if ok && string(fcb.Language(src)) == "mermaid" {
			targets = append(targets, fcb)
		}
		return ast.WalkContinue, nil
	})

	for _, fcb := range targets {
		parent := fcb.Parent()
		if parent == nil {
			continue
		}
		var buf bytes.Buffer
		lines := fcb.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value(src))
		}
		mn := &mermaidNode{source: buf.String()}
		parent.InsertBefore(parent, fcb, mn)
		parent.RemoveChild(parent, fcb)
	}
}

// mermaidRenderer renders mermaidNode as <div class="mermaid">…</div>.
type mermaidRenderer struct{}

func (r *mermaidRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMermaidBlock, r.renderMermaidBlock)
}

func (r *mermaidRenderer) renderMermaidBlock(
	w util.BufWriter, _ []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	mn := node.(*mermaidNode)
	_, _ = w.WriteString(`<div class="mermaid">`)
	_, _ = w.WriteString(html.EscapeString(mn.source))
	_, _ = w.WriteString(`</div>`)
	return ast.WalkSkipChildren, nil
}

// Extension returns a goldmark.Extender that transforms fenced "mermaid"
// code blocks into <div class="mermaid"> elements using an AST transformer,
// avoiding any conflict with KindFencedCodeBlock renderers (e.g. chroma).
func Extension() goldmark.Extender {
	return &mermaidExtender{}
}

type mermaidExtender struct{}

func (e *mermaidExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		goldmarkparser.WithASTTransformers(
			util.Prioritized(&mermaidTransformer{}, 100),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&mermaidRenderer{}, 500),
		),
	)
}
