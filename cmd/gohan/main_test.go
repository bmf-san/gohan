package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestPrintUsage_WritesToStderr(t *testing.T) {
	// Redirect stderr and capture output.
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	printUsage()

	w.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	for _, want := range []string{"build", "new", "serve", "version"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("printUsage output missing %q:\n%s", want, out)
		}
	}
}

func TestVersionVars_Defaults(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
	if commit == "" {
		t.Error("commit should not be empty")
	}
	if date == "" {
		t.Error("date should not be empty")
	}
}

// testdataDir returns the path to cmd/gohan/testdata/.
// Go's test runner sets cwd to the package directory, so "testdata" works.
func testdataDir(t *testing.T) string {
	t.Helper()
	return "testdata"
}

func TestRunBuild_FullBuild(t *testing.T) {
	// Copy testdata to a temp dir so the output doesn't pollute the source tree.
	src := testdataDir(t)
	dir := t.TempDir()
	if err := copyDir(src, dir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	// Use a relative output name; build.go does filepath.Join(rootDir, outputDir)
	// so "public" resolves to <dir>/public.
	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--output=public",
	})
	if err != nil {
		t.Fatalf("full build: %v", err)
	}

	outDir := filepath.Join(dir, "public")

	// index.html must exist
	idx := filepath.Join(outDir, "index.html")
	if _, statErr := os.Stat(idx); statErr != nil {
		t.Errorf("index.html not created: %v", statErr)
	}

	// sitemap.xml
	sm := filepath.Join(outDir, "sitemap.xml")
	if _, statErr := os.Stat(sm); statErr != nil {
		t.Errorf("sitemap.xml not created: %v", statErr)
	}
}

func TestRunBuild_FullBuild_ParallelOverride(t *testing.T) {
	src := testdataDir(t)
	dir := t.TempDir()
	if err := copyDir(src, dir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	err := runBuild([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--output=out",
		"--parallel=2",
	})
	if err != nil {
		t.Fatalf("full build with --parallel=2: %v", err)
	}
}

func TestRunBuild_IncrementalAfterFull(t *testing.T) {
	src := testdataDir(t)
	dir := t.TempDir()
	if err := copyDir(src, dir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	cfgFlag := "--config=" + filepath.Join(dir, "config.yaml")

	// First: full build (relative output so rootDir / "public" is correct)
	if err := runBuild([]string{cfgFlag, "--output=public", "--full"}); err != nil {
		t.Fatalf("full build: %v", err)
	}
	// Second: incremental build (uses cached manifest)
	if err := runBuild([]string{cfgFlag, "--output=public"}); err != nil {
		t.Fatalf("incremental build: %v", err)
	}
}

// copyDir recursively copies src directory to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
