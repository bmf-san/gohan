//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestRunBuild_ConcurrentBuildSkipped verifies that when another process already
// holds the exclusive build lock, runBuild returns nil immediately (skips).
func TestRunBuild_ConcurrentBuildSkipped(t *testing.T) {
	dir := t.TempDir()
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-acquire the exclusive lock that runBuild will try to obtain.
	lockDir := filepath.Join(dir, ".gohan")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lockFile, err := os.OpenFile(filepath.Join(lockDir, "build.lock"), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lockFile.Close() }()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("could not acquire pre-lock: %v", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) //nolint:errcheck

	// runBuild should detect the held lock and skip without error.
	if err := runBuild([]string{"--config=" + filepath.Join(dir, "config.yaml")}); err != nil {
		t.Fatalf("expected nil when lock held, got: %v", err)
	}
}
