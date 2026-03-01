package server

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// ─────────────────────────────────────────
// FileWatcher interface
// ─────────────────────────────────────────

// FileWatcher is the file change detection interface. The implementation uses fsnotify.
type FileWatcher interface {
	Add(path string) error
	Events() <-chan string
	Close() error
}

// ─────────────────────────────────────────
// FsnotifyWatcher: FileWatcher backed by fsnotify
// ─────────────────────────────────────────

// FsnotifyWatcher implements FileWatcher using github.com/fsnotify/fsnotify.
type FsnotifyWatcher struct {
	watcher *fsnotify.Watcher
	events  chan string
	done    chan struct{}
}

// NewFsnotifyWatcher creates and starts a new FsnotifyWatcher.
func NewFsnotifyWatcher() (*FsnotifyWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	fw := &FsnotifyWatcher{
		watcher: w,
		events:  make(chan string, 100),
		done:    make(chan struct{}),
	}
	go fw.loop()
	return fw, nil
}

func (fw *FsnotifyWatcher) loop() {
	defer close(fw.events)
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			// Filter: only care about write/create/remove events
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				select {
				case fw.events <- event.Name:
				default: // drop if buffer full
				}
			}
		case _, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
		case <-fw.done:
			return
		}
	}
}

// Add adds a path to be watched.
func (fw *FsnotifyWatcher) Add(path string) error { return fw.watcher.Add(path) }

// Events returns the channel that receives changed file paths.
func (fw *FsnotifyWatcher) Events() <-chan string { return fw.events }

// Close stops the watcher.
func (fw *FsnotifyWatcher) Close() error {
	close(fw.done)
	return fw.watcher.Close()
}

// ─────────────────────────────────────────
// SSE broadcaster
// ─────────────────────────────────────────

type sseBroadcaster struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func newSSEBroadcaster() *sseBroadcaster {
	return &sseBroadcaster{clients: make(map[chan string]struct{})}
}

func (b *sseBroadcaster) subscribe() chan string {
	ch := make(chan string, 1)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *sseBroadcaster) unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
}

func (b *sseBroadcaster) broadcast(msg string) {
	b.mu.Lock()
	clients := make([]chan string, 0, len(b.clients))
	for ch := range b.clients {
		clients = append(clients, ch)
	}
	b.mu.Unlock()
	for _, ch := range clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

// ─────────────────────────────────────────
// Script-injecting response writer
// ─────────────────────────────────────────

const sseScript = `<script>(function(){` +
	`var e=new EventSource("/__gohan/reload");` +
	`e.onmessage=function(){location.reload();};` +
	`})()</script>`

type injectingResponseWriter struct {
	wrapped    http.ResponseWriter
	buf        bytes.Buffer
	header     int
	isHTML     bool
	headerSent bool
}

func (w *injectingResponseWriter) Header() http.Header { return w.wrapped.Header() }

func (w *injectingResponseWriter) WriteHeader(code int) {
	ct := w.Header().Get("Content-Type")
	w.isHTML = strings.Contains(ct, "text/html")
	w.header = code
}

func (w *injectingResponseWriter) Write(b []byte) (int, error) {
	if !w.isHTML {
		return w.wrapped.Write(b)
	}
	return w.buf.Write(b)
}

func (w *injectingResponseWriter) flush() {
	if !w.isHTML {
		return
	}
	if w.header != 0 {
		w.wrapped.WriteHeader(w.header)
	}
	body := w.buf.Bytes()
	script := []byte(sseScript)
	if idx := bytes.Index(body, []byte("</body>")); idx >= 0 {
		injected := make([]byte, 0, len(body)+len(script))
		injected = append(injected, body[:idx]...)
		injected = append(injected, script...)
		injected = append(injected, body[idx:]...)
		body = injected
	} else {
		body = append(body, script...)
	}
	_, _ = w.wrapped.Write(body)
}

// injectingHandler wraps handler and injects the SSE script into HTML responses.
func injectingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iw := &injectingResponseWriter{wrapped: w}
		h.ServeHTTP(iw, r)
		iw.flush()
	})
}

// ─────────────────────────────────────────
// DevServer
// ─────────────────────────────────────────

// DevServer is a local HTTP development server with live reload.
type DevServer struct {
	Host        string
	Port        int
	OutDir      string
	Watcher     FileWatcher
	RebuildFunc func() error // called on file change; may be nil
}

// NewDevServer creates a new DevServer.
// rebuildFn is called when a watched file changes; pass nil to disable rebuild.
func NewDevServer(host string, port int, outDir string, rebuildFn func() error) *DevServer {
	return &DevServer{
		Host:        host,
		Port:        port,
		OutDir:      outDir,
		RebuildFunc: rebuildFn,
	}
}

// WatchDirs is the set of directories monitored for changes.
var WatchDirs = []string{"content", "themes", "assets"}

// Start starts the development server and blocks until it exits.
func (s *DevServer) Start() error {
	broadcaster := newSSEBroadcaster()

	// Set up HTTP mux
	mux := http.NewServeMux()

	// SSE endpoint: /__gohan/reload
	mux.HandleFunc("/__gohan/reload", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := broadcaster.subscribe()
		defer broadcaster.unsubscribe(ch)

		notify := r.Context().Done()
		for {
			select {
			case <-notify:
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	})

	// Static file server with script injection
	fileHandler := injectingHandler(http.FileServer(http.Dir(s.OutDir)))
	mux.Handle("/", fileHandler)

	// Start file watcher if available
	if s.Watcher == nil {
		// Try to create a real watcher; silently skip if unavailable
		if fw, err := NewFsnotifyWatcher(); err == nil {
			s.Watcher = fw
			defer fw.Close()
		}
	}
	if s.Watcher != nil {
		for _, dir := range WatchDirs {
			_ = s.Watcher.Add(dir) // ignore missing dirs
		}
		go s.watchLoop(broadcaster)
	}

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	return http.ListenAndServe(addr, mux)
}

// watchLoop listens for file change events and triggers rebuild + SSE broadcast.
func (s *DevServer) watchLoop(b *sseBroadcaster) {
	for path := range s.Watcher.Events() {
		if s.RebuildFunc != nil {
			_ = s.RebuildFunc() // best-effort; errors go to stderr via rebuild
		}
		b.broadcast(path)
	}
}
