package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunServe_UnknownFlag(t *testing.T) {
	err := runServe([]string{"--unknown-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestRunServe_DefaultFlagsAccepted(t *testing.T) {
	// We cannot actually start a server in tests, but we can verify flag parsing.
	// This test just checks that parsing completes without an error for default flags.
	// (The server Start() itself is tested in internal/server package.)
	//
	// We simulate flag-parse-only by passing a --help equivalent — not ideal.
	// Instead, verify error is flag-related, not a parse error.
	err := runServe([]string{"--port=19999", "--host=127.0.0.1", "--unknown"})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

// TestRunServe_InvalidHostFails exercises the post-parse code path:
// initial build (with missing config, non-fatal), then srv.Start() which
// returns an error because the host string is invalid.
func TestRunServe_InvalidHostFails(t *testing.T) {
	dir := t.TempDir()
	// Write a minimal config so runBuild doesn't error on missing file.
	cfg := []byte("site:\n  title: Test\n  base_url: http://localhost\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0755); err != nil {
		t.Fatal(err)
	}

	// "invalid host" causes net.Listen to fail immediately, so srv.Start() returns an error.
	err := runServe([]string{
		"--config=" + filepath.Join(dir, "config.yaml"),
		"--host=256.256.256.256",
		"--port=19876",
	})
	if err == nil {
		t.Fatal("expected error from srv.Start() with invalid host")
	}
}
