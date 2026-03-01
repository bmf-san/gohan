package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
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
