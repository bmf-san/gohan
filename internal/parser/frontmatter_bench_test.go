package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkFileParser_ParseAll measures the cost of parsing a content tree of
// the given size. Tracked in CI to detect regressions in the hot parse path.
func BenchmarkFileParser_ParseAll(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("articles=%d", n), func(b *testing.B) {
			dir := b.TempDir()
			seedBenchmarkArticles(b, dir, n)
			p := NewFileParser()
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				articles, err := p.ParseAll(dir)
				if err != nil {
					b.Fatalf("ParseAll: %v", err)
				}
				if len(articles) != n {
					b.Fatalf("expected %d articles, got %d", n, len(articles))
				}
			}
		})
	}
}

func seedBenchmarkArticles(b *testing.B, dir string, n int) {
	b.Helper()
	const body = "# Heading\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n" +
		"## Subheading\n\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\n\n" +
		"```go\nfunc Hello() string { return \"world\" }\n```\n"
	for i := 0; i < n; i++ {
		path := filepath.Join(dir, fmt.Sprintf("post-%04d.md", i))
		content := fmt.Sprintf(
			"---\ntitle: Post %d\ndate: 2024-01-%02d\ntags: [a, b, c]\ndescription: bench post %d\n---\n%s",
			i, (i%28)+1, i, body,
		)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			b.Fatalf("write: %v", err)
		}
	}
}
