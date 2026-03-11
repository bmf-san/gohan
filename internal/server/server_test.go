package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDevServer(t *testing.T) {
	srv := NewDevServer("127.0.0.1", 1313, "public", nil)
	if srv.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", srv.Host)
	}
	if srv.Port != 1313 {
		t.Errorf("expected port 1313, got %d", srv.Port)
	}
	if srv.OutDir != "public" {
		t.Errorf("expected outDir public, got %s", srv.OutDir)
	}
	if srv.RootDir != "" {
		t.Errorf("expected empty RootDir by default, got %s", srv.RootDir)
	}
}

func TestSSEBroadcaster(t *testing.T) {
	b := newSSEBroadcaster()
	ch := b.subscribe()

	b.broadcast("reload")

	select {
	case msg := <-ch:
		if msg != "reload" {
			t.Errorf("expected reload, got %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for SSE broadcast")
	}

	b.unsubscribe(ch)
}

func TestInjectScript_WithBody(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Hello</body></html>"))
	})

	handler := injectingHandler(inner)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	body := rec.Body.String()
	if !strings.Contains(body, sseScript) {
		t.Errorf("expected SSE script injected, body:\n%s", body)
	}
	if !strings.Contains(body, "Hello") {
		t.Error("original content must be preserved")
	}
}

func TestInjectScript_NoBody(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>no body tag</html>"))
	})

	handler := injectingHandler(inner)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	body := rec.Body.String()
	if !strings.Contains(body, sseScript) {
		t.Errorf("expected SSE script appended when no </body>, body:\n%s", body)
	}
}

func TestInjectScript_NotHTML(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("body { color: red; }"))
	})

	handler := injectingHandler(inner)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/style.css", nil))

	body := rec.Body.String()
	if strings.Contains(body, "<script>") {
		t.Error("must not inject script into non-HTML response")
	}
}

func TestInjectScript_NonHTMLStatusPropagated(t *testing.T) {
	// When http.FileServer returns a 404 (text/plain, not text/html),
	// the status code must be forwarded to the client, not silently become 200.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 page not found", http.StatusNotFound)
	})

	handler := injectingHandler(inner)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/missing/", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestInjectingResponseWriter_Buffer(t *testing.T) {
	rec := httptest.NewRecorder()
	iw := &injectingResponseWriter{wrapped: rec}
	iw.Header().Set("Content-Type", "text/html")
	iw.WriteHeader(http.StatusOK)
	_, _ = iw.Write([]byte("<html><body>test</body></html>"))
	iw.flush()

	if !bytes.Contains(rec.Body.Bytes(), []byte(sseScript)) {
		t.Error("expected SSE script in flushed response")
	}
}

// mockWatcher is a FileWatcher stub for testing.
type mockWatcher struct {
	events chan string
}

func (m *mockWatcher) Add(_ string) error    { return nil }
func (m *mockWatcher) Events() <-chan string { return m.events }
func (m *mockWatcher) Close() error          { close(m.events); return nil }

func TestWatchLoop_Debounce(t *testing.T) {
	// Verify that multiple rapid events are coalesced into a single rebuild.
	watcher := &mockWatcher{events: make(chan string, 20)}

	var rebuildCount atomic.Int32
	var broadcastCount atomic.Int32

	b := newSSEBroadcaster()
	ch := b.subscribe()
	go func() {
		for range ch {
			broadcastCount.Add(1)
		}
	}()

	srv := &DevServer{
		Watcher: watcher,
		RebuildFunc: func() error {
			rebuildCount.Add(1)
			return nil
		},
	}

	go srv.watchLoop(b)

	// Send 5 events in rapid succession (well within debounce window).
	for i := 0; i < 5; i++ {
		watcher.events <- "content/post.md"
	}

	// Wait longer than the debounce delay for the single rebuild to complete.
	time.Sleep(debounceDelay*3 + 50*time.Millisecond)

	if got := rebuildCount.Load(); got != 1 {
		t.Errorf("expected 1 rebuild, got %d (debounce not working)", got)
	}
	if got := broadcastCount.Load(); got != 1 {
		t.Errorf("expected 1 broadcast, got %d", got)
	}
	b.unsubscribe(ch)
}
