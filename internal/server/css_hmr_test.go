package server

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsCSSPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"assets/style.css", true},
		{"assets/STYLE.CSS", true},
		{"themes/default/static/css/main.css", true},
		{"content/post.md", false},
		{"templates/index.html", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isCSSPath(c.path); got != c.want {
			t.Errorf("isCSSPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestWatchLoop_CSSOnlyBroadcastsCSS(t *testing.T) {
	watcher := &mockWatcher{events: make(chan string, 20)}
	b := newSSEBroadcaster()
	ch := b.subscribe()
	var got atomic.Value
	got.Store("")
	done := make(chan struct{})
	go func() {
		msg, ok := <-ch
		if ok {
			got.Store(msg)
		}
		close(done)
	}()

	srv := &DevServer{Watcher: watcher}
	go srv.watchLoop(b)

	watcher.events <- "assets/style.css"
	watcher.events <- "assets/extra.css"

	select {
	case <-done:
	case <-time.After(debounceDelay*5 + 100*time.Millisecond):
		t.Fatal("timed out waiting for broadcast")
	}
	if v, _ := got.Load().(string); v != "css" {
		t.Errorf("expected broadcast %q for CSS-only changes, got %q", "css", v)
	}
	b.unsubscribe(ch)
}

func TestWatchLoop_MixedBroadcastsPath(t *testing.T) {
	watcher := &mockWatcher{events: make(chan string, 20)}
	b := newSSEBroadcaster()
	ch := b.subscribe()
	var got atomic.Value
	got.Store("")
	done := make(chan struct{})
	go func() {
		msg, ok := <-ch
		if ok {
			got.Store(msg)
		}
		close(done)
	}()

	srv := &DevServer{Watcher: watcher}
	go srv.watchLoop(b)

	watcher.events <- "assets/style.css"
	watcher.events <- "content/post.md"

	select {
	case <-done:
	case <-time.After(debounceDelay*5 + 100*time.Millisecond):
		t.Fatal("timed out waiting for broadcast")
	}
	if v, _ := got.Load().(string); v == "css" || v == "" {
		t.Errorf("expected non-CSS path broadcast for mixed changes, got %q", v)
	}
	b.unsubscribe(ch)
}

func TestSSEScript_HandlesCSSMessage(t *testing.T) {
	// Sanity check the injected client script contains the CSS branch.
	if !strings.Contains(sseScript, "ev.data===\"css\"") {
		t.Errorf("sseScript missing CSS branch:\n%s", sseScript)
	}
	if !strings.Contains(sseScript, "stylesheet") {
		t.Errorf("sseScript missing stylesheet selector:\n%s", sseScript)
	}
	if !strings.Contains(sseScript, "location.reload()") {
		t.Errorf("sseScript missing full-reload fallback:\n%s", sseScript)
	}
}
