package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")
	data := []byte("<html>hello</html>")

	if err := writeFileAtomic(path, data, 0o644); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("content mismatch: got %q, want %q", got, data)
	}
}

func TestWriteFileAtomic_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")

	// 初回書き込み
	if err := writeFileAtomic(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("first write: %v", err)
	}
	// 上書き
	want := []byte("second")
	if err := writeFileAtomic(path, want, 0o644); err != nil {
		t.Fatalf("second write: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(want) {
		t.Fatalf("overwrite: got %q, want %q", got, want)
	}
}

func TestWriteFileAtomic_NoTempLeftBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")

	if err := writeFileAtomic(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}

	// 一時ファイル (.gohan-tmp-*) が残っていないこと
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "out.html" {
			t.Errorf("unexpected file left behind: %s", e.Name())
		}
	}
}

func TestWriteFileAtomic_Permissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")

	if err := writeFileAtomic(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions: got %04o, want %04o", got, 0o600)
	}
}

func TestWriteFileAtomic_InvalidDir(t *testing.T) {
	// 存在しないディレクトリへの書き込みはエラーになること
	err := writeFileAtomic("/nonexistent-dir/out.html", []byte("data"), 0o644)
	if err == nil {
		t.Fatal("expected error for nonexistent dir, got nil")
	}
}
