package processor

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmf-san/gohan/internal/model"
)

// updateGolden controls whether failing golden HTML tests overwrite the
// snapshot files instead of failing. Run with:
//
//	go test ./internal/processor -update
//
// after intentional rendering changes, then commit the new testdata files.
var updateGolden = flag.Bool("update", false, "update golden HTML snapshots in testdata/golden/")

// TestProcess_GoldenHTML asserts that the markdown→HTML rendering pipeline
// produces stable output for a curated set of inputs. The fixtures live under
// testdata/golden/<name>.md and the expected HTML in testdata/golden/<name>.html.
//
// This guards against accidental regressions in Goldmark configuration, code
// highlighting, mermaid handling, or summary/word-count logic.
func TestProcess_GoldenHTML(t *testing.T) {
	dir := filepath.Join("testdata", "golden")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read golden dir: %v", err)
	}

	var cases []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		cases = append(cases, strings.TrimSuffix(e.Name(), ".md"))
	}
	if len(cases) == 0 {
		t.Fatalf("no golden fixtures found under %s", dir)
	}

	p := NewSiteProcessor()
	cfg := model.Config{}

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			mdPath := filepath.Join(dir, name+".md")
			htmlPath := filepath.Join(dir, name+".html")

			raw, err := os.ReadFile(mdPath)
			if err != nil {
				t.Fatalf("read %s: %v", mdPath, err)
			}

			// Normalize line endings so the test is stable on Windows.
			rawStr := strings.ReplaceAll(string(raw), "\r\n", "\n")

			// Strip a leading YAML front matter block if present so the test
			// only exercises the body→HTML conversion.
			body := stripFrontMatter(rawStr)

			articles := []*model.Article{{
				FilePath:   mdPath,
				RawContent: body,
			}}
			processed, err := p.Process(articles, cfg)
			if err != nil {
				t.Fatalf("Process: %v", err)
			}
			if len(processed) != 1 {
				t.Fatalf("expected 1 processed article, got %d", len(processed))
			}
			got := strings.TrimRight(string(processed[0].HTMLContent), "\n") + "\n"

			if *updateGolden {
				if err := os.WriteFile(htmlPath, []byte(got), 0o644); err != nil {
					t.Fatalf("update golden: %v", err)
				}
				return
			}

			wantRaw, err := os.ReadFile(htmlPath)
			if err != nil {
				t.Fatalf("read %s: %v (run with -update to create)", htmlPath, err)
			}
			// Normalize line endings so the test is stable on Windows where
			// git may check files out with CRLF.
			want := strings.ReplaceAll(string(wantRaw), "\r\n", "\n")
			if got != want {
				t.Errorf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s",
					name, got, want)
			}
		})
	}
}

// stripFrontMatter removes a leading "---\n...\n---\n" block, if present.
func stripFrontMatter(s string) string {
	if !strings.HasPrefix(s, "---\n") {
		return s
	}
	end := strings.Index(s[4:], "\n---\n")
	if end < 0 {
		return s
	}
	return s[4+end+5:]
}
