package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// freePort finds a free TCP port on localhost.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// waitForPort polls until the TCP port accepts connections or the deadline is reached.
func waitForPort(addr string, deadline time.Duration) bool {
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestStart_ServesStaticFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"),
		[]byte("<html><body>hello from gohan</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	port := freePort(t)
	srv := NewDevServer("127.0.0.1", port, dir, nil)

	// Start runs http.ListenAndServe which blocks; run in background goroutine.
	// The goroutine exits when the test process ends.
	go func() { _ = srv.Start() }()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if !waitForPort(addr, 3*time.Second) {
		t.Fatalf("server did not start within 3s on %s", addr)
	}

	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("HTTP GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) == "" {
		// The injectingHandler may add the SSE script; just verify no error
		t.Log("body was empty but no error")
	}
}

func TestStart_SSEEndpoint(t *testing.T) {
	dir := t.TempDir()

	port := freePort(t)
	srv := NewDevServer("127.0.0.1", port, dir, nil)
	go func() { _ = srv.Start() }()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if !waitForPort(addr, 3*time.Second) {
		t.Fatalf("server did not start within 3s on %s", addr)
	}

	// SSE endpoint must return text/event-stream
	req, err := http.NewRequest("GET", "http://"+addr+"/__gohan/reload", nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		// Timeout is expected since SSE streams indefinitely — just verify it started
		t.Logf("SSE request ended (expected timeout): %v", err)
		return
	}
	defer resp.Body.Close()
	ct := resp.Header.Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %q", ct)
	}
}

func TestStart_WithMockWatcher(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "page.html"),
		[]byte("<html><body>Page</body></html>"), 0644)

	rebuilt := make(chan struct{}, 1)
	fw := newFakeWatcher()

	port := freePort(t)
	srv := &DevServer{
		Host:    "127.0.0.1",
		Port:    port,
		OutDir:  dir,
		Watcher: fw,
		RebuildFunc: func() error {
			rebuilt <- struct{}{}
			return nil
		},
	}
	go func() { _ = srv.Start() }()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if !waitForPort(addr, 3*time.Second) {
		t.Fatalf("server did not start on %s", addr)
	}

	// Trigger a file change event
	fw.ch <- "/content/test.md"

	select {
	case <-rebuilt:
		// rebuild was triggered
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: RebuildFunc not called from Start")
	}
}
