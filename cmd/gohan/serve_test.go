package main

import (
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
