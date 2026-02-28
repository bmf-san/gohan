package diff

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

func TestWriteReadManifest(t *testing.T) {
	dir := t.TempDir()
	m := &model.BuildManifest{
		Version:   "1",
		BuildTime: time.Now().UTC().Truncate(time.Second),
		FileHashes: map[string]string{"a.md": "abc"},
	}
	if err := WriteManifest(dir, m); err != nil { t.Fatalf("WriteManifest: %v", err) }
	got, err := ReadManifest(dir)
	if err != nil { t.Fatalf("ReadManifest: %v", err) }
	if got.Version != m.Version { t.Errorf("version mismatch") }
	if got.FileHashes["a.md"] != "abc" { t.Error("hash mismatch") }
}

func TestReadManifest_NotExist(t *testing.T) {
	dir := t.TempDir()
	m, err := ReadManifest(dir)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if m != nil { t.Error("expected nil") }
}

func TestWriteReadCachedHTML(t *testing.T) {
	dir := t.TempDir()
	html := "<h1>hello</h1>"
	if err := WriteCachedHTML(dir, "my-post", html); err != nil { t.Fatalf("write: %v", err) }
	got, err := ReadCachedHTML(dir, "my-post")
	if err != nil { t.Fatalf("read: %v", err) }
	if got != html { t.Errorf("got %q, want %q", got, html) }
}

func TestReadCachedHTML_NotExist(t *testing.T) {
	dir := t.TempDir()
	got, err := ReadCachedHTML(dir, "nope")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if got != "" { t.Error("expected empty string") }
}

func TestClearCache(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "cache")
	if err := os.MkdirAll(sub, 0755); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(sub, "x.json"), []byte("{}"), 0644); err != nil { t.Fatal(err) }
	if err := ClearCache(sub); err != nil { t.Fatalf("ClearCache: %v", err) }
	if _, err := os.Stat(sub); !os.IsNotExist(err) { t.Error("expected removed") }
}

func TestCheckConfigChange_NilManifest(t *testing.T) {
	if !CheckConfigChange(nil, "hash") { t.Error("expected true for nil") }
}

func TestCheckConfigChange_Same(t *testing.T) {
	m := &model.BuildManifest{FileHashes: map[string]string{"__config__": "h1"}}
	if CheckConfigChange(m, "h1") { t.Error("expected false") }
}

func TestCheckConfigChange_Different(t *testing.T) {
	m := &model.BuildManifest{FileHashes: map[string]string{"__config__": "old"}}
	if !CheckConfigChange(m, "new") { t.Error("expected true") }
}

func TestNewManifest(t *testing.T) {
	m := NewManifest("myhash")
	if m.Version != "1" { t.Errorf("version: %s", m.Version) }
	if m.FileHashes["__config__"] != "myhash" { t.Errorf("config hash: %v", m.FileHashes) }
}
