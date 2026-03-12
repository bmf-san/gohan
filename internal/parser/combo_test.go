package parser_test

import (
"strings"
"testing"

"github.com/bmf-san/gohan/internal/highlight"
"github.com/bmf-san/gohan/internal/parser"
)

func TestConverter_MermaidPlusHighlight_CodeBlockNotDropped(t *testing.T) {
hlCfg := highlight.Config{Theme: "github", LineNumbers: false}
conv := parser.NewConverter(parser.WithGFM(), parser.WithMermaid(), parser.WithHighlighting(hlCfg))

src := "```yaml\nfoo: bar\n```\n"
out, err := conv.Convert([]byte(src))
if err != nil {
t.Fatal(err)
}
t.Logf("yaml output: %s", string(out))
if !strings.Contains(string(out), "<pre") {
t.Errorf("expected <pre> tag for yaml code block, got: %q", string(out))
}
}
