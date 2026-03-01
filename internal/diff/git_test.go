package diff

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmf-san/gohan/internal/model"
)

func fileHash(t *testing.T, path string) string {
	t.Helper()
	h, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}
	return h
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "gohan-diff-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestIsGitRepo_NonGit(t *testing.T) {
	dir := t.TempDir()
	if IsGitRepo(dir) {
		t.Error("expected false for plain temp dir")
	}
}

func TestIsGitRepo_GitDir(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if !IsGitRepo(root) {
		t.Errorf("expected true for project root %s", root)
	}
}

func TestHash_ValidFile(t *testing.T) {
	content := "hello gohan"
	path := writeTemp(t, content)
	defer func() { _ = os.Remove(path) }()
	got, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}
	sum := sha256.Sum256([]byte(content))
	want := hex.EncodeToString(sum[:])
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestHash_MissingFile(t *testing.T) {
	_, err := hashFile("/no/such/file.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDetect_NilManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("A"), 0644); err != nil {
		t.Fatal(err)
	}
	eng := NewGitDiffEngine(dir)
	cs, err := eng.Detect(nil)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(cs.AddedFiles) != 1 {
		t.Errorf("expected 1 added file, got %d", len(cs.AddedFiles))
	}
}

func TestDetect_NoChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "b.md")
	if err := os.WriteFile(path, []byte("B"), 0644); err != nil {
		t.Fatal(err)
	}
	h := fileHash(t, path)
	manifest := &model.BuildManifest{FileHashes: map[string]string{"b.md": h}}
	eng := NewGitDiffEngine(dir)
	cs, err := eng.Detect(manifest)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(cs.AddedFiles)+len(cs.ModifiedFiles)+len(cs.DeletedFiles) != 0 {
		t.Error("expected empty ChangeSet")
	}
}

func TestDetect_ModifiedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.md")
	if err := os.WriteFile(path, []byte("C"), 0644); err != nil {
		t.Fatal(err)
	}
	manifest := &model.BuildManifest{FileHashes: map[string]string{"c.md": "stale-hash"}}
	eng := NewGitDiffEngine(dir)
	cs, err := eng.Detect(manifest)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(cs.ModifiedFiles) != 1 {
		t.Errorf("expected 1 modified file, got %d", len(cs.ModifiedFiles))
	}
}

func TestDetect_DeletedFile(t *testing.T) {
	dir := t.TempDir()
	manifest := &model.BuildManifest{FileHashes: map[string]string{"gone.md": "some-hash"}}
	eng := NewGitDiffEngine(dir)
	cs, err := eng.Detect(manifest)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(cs.DeletedFiles) != 1 {
		t.Errorf("expected 1 deleted file, got %d", len(cs.DeletedFiles))
	}
}

func TestParseNameStatus(t *testing.T) {
	output := "M\tinternal/foo.go\nA\tinternal/bar.go\nD\tinternal/old.go\n"
	cs := parseNameStatus(output)
	if len(cs.ModifiedFiles) != 1 {
		t.Errorf("modified: %v", cs.ModifiedFiles)
	}
	if len(cs.AddedFiles) != 1 {
		t.Errorf("added: %v", cs.AddedFiles)
	}
	if len(cs.DeletedFiles) != 1 {
		t.Errorf("deleted: %v", cs.DeletedFiles)
	}
}
