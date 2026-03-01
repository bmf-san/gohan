package server

import (
	"testing"
)

func TestNewDevServer(t *testing.T) {
	srv := NewDevServer("127.0.0.1", 1313, "public")
	if srv.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", srv.Host)
	}
	if srv.Port != 1313 {
		t.Errorf("expected port 1313, got %d", srv.Port)
	}
	if srv.OutDir != "public" {
		t.Errorf("expected outDir public, got %s", srv.OutDir)
	}
}
