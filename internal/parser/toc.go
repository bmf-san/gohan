package parser

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	goldmarkparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"github.com/bmf-san/gohan/internal/model"
)

// ExtractTOC parses Markdown source and returns the hierarchical table of
// contents built from headings. Heading IDs match those produced by the
// Markdown converter (goldmark with AutoHeadingID enabled).
//
// Headings at any depth are nested under the nearest preceding heading whose
// level is strictly smaller. When the first heading in the document is not at
// the top level (e.g. starts with H3), it becomes a top-level TOC entry.
func ExtractTOC(src []byte) []model.TOCEntry {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(goldmarkparser.WithAutoHeadingID()),
	)
	reader := text.NewReader(src)
	doc := md.Parser().Parse(reader)

	var flat []model.TOCEntry
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		id, _ := h.AttributeString("id")
		idStr := ""
		if b, ok := id.([]byte); ok {
			idStr = string(b)
		}
		flat = append(flat, model.TOCEntry{
			Level: h.Level,
			ID:    idStr,
			Text:  headingText(h, src),
		})
		return ast.WalkSkipChildren, nil
	})

	return nestEntries(flat)
}

// headingText extracts the plain-text content of a heading node.
func headingText(h *ast.Heading, src []byte) string {
	var buf bytes.Buffer
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(src))
			continue
		}
		// Fall back to walking the segments of any other inline node.
		_ = ast.Walk(c, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			if t, ok := n.(*ast.Text); ok {
				buf.Write(t.Segment.Value(src))
			}
			return ast.WalkContinue, nil
		})
	}
	return buf.String()
}

// nestEntries converts a flat slice of TOC entries into a hierarchical tree.
// Each entry is nested under the most recently seen entry with a smaller level.
func nestEntries(flat []model.TOCEntry) []model.TOCEntry {
	if len(flat) == 0 {
		return nil
	}
	var roots []model.TOCEntry
	// Stack of pointers to entries currently being filled in by deeper levels.
	type frame struct {
		level   int
		parent  *[]model.TOCEntry // slice owning the entry
		index   int               // entry index within *parent
	}
	var stack []frame
	// rootsRef lets the first frame point at the roots slice itself.
	rootsRef := &roots
	for _, e := range flat {
		// Pop frames until the top has a strictly smaller level.
		for len(stack) > 0 && stack[len(stack)-1].level >= e.Level {
			stack = stack[:len(stack)-1]
		}
		var owner *[]model.TOCEntry
		if len(stack) == 0 {
			owner = rootsRef
		} else {
			top := stack[len(stack)-1]
			owner = &(*top.parent)[top.index].Children
		}
		*owner = append(*owner, e)
		stack = append(stack, frame{level: e.Level, parent: owner, index: len(*owner) - 1})
	}
	return roots
}
