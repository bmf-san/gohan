package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "warn: file watcher: %v\n", err)
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
	wrapped       http.ResponseWriter
	buf           bytes.Buffer
	header        int
	isHTML        bool
	headerWritten bool
}

func (w *injectingResponseWriter) Header() http.Header { return w.wrapped.Header() }

func (w *injectingResponseWriter) WriteHeader(code int) {
	ct := w.Header().Get("Content-Type")
	w.isHTML = strings.Contains(ct, "text/html")
	w.header = code
}

func (w *injectingResponseWriter) Write(b []byte) (int, error) {
	// http.FileServer may call Write without calling WriteHeader first (implicit
	// 200). Detect HTML from Content-Type at first Write if not yet determined.
	if w.header == 0 {
		ct := w.Header().Get("Content-Type")
		w.isHTML = strings.Contains(ct, "text/html")
		w.header = http.StatusOK
	}
	if !w.isHTML {
		// Propagate non-200 status codes (e.g. 404) that were stored by
		// WriteHeader but not yet forwarded to the underlying ResponseWriter.
		// Without this, net/http implicitly sends 200 on the first Write.
		if !w.headerWritten && w.header != 0 && w.header != http.StatusOK {
			w.wrapped.WriteHeader(w.header)
			w.headerWritten = true
		}
		return w.wrapped.Write(b)
	}
	return w.buf.Write(b)
}

func (w *injectingResponseWriter) flush() {
	if !w.isHTML {
		// Forward any stored status code that was never written to the underlying
		// ResponseWriter (e.g. 304 Not Modified, where Write is never called).
		// Without this, Go's net/http falls back to an implicit 200 OK with an
		// empty body, causing the browser to show a blank page on reload.
		if !w.headerWritten && w.header != 0 && w.header != http.StatusOK {
			w.wrapped.WriteHeader(w.header)
		}
		return
	}
	// Remove Content-Length set by http.FileServer: the injected SSE script
	// makes the actual body larger, so the original value is wrong.
	// Deleting it lets net/http use chunked transfer encoding instead.
	w.wrapped.Header().Del("Content-Length")
	if w.header != 0 {
		w.wrapped.WriteHeader(w.header)
		w.headerWritten = true
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

// noListFS wraps http.Dir and disables directory listings.
// Directories without an index.html return os.ErrNotExist so that
// http.FileServer responds with 404 instead of showing a file list.
type noListFS struct{ base http.Dir }

func (fs noListFS) Open(name string) (http.File, error) {
	f, err := fs.base.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if info.IsDir() {
		idx, err := fs.base.Open(strings.TrimSuffix(name, "/") + "/index.html")
		if err != nil {
			_ = f.Close()
			return nil, os.ErrNotExist
		}
		_ = idx.Close()
	}
	return f, nil
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
	RootDir     string // project root; WatchDirs are resolved relative to this when set
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
				_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	})

	// Static file server with script injection.
	// noListFS disables directory listings: directories without index.html return 404.
	fileHandler := injectingHandler(http.FileServer(noListFS{http.Dir(s.OutDir)}))
	mux.Handle("/", fileHandler)

	// Start file watcher if available
	if s.Watcher == nil {
		// Try to create a real watcher; silently skip if unavailable
		if fw, err := NewFsnotifyWatcher(); err == nil {
			s.Watcher = fw
			defer func() { _ = fw.Close() }()
		}
	}
	if s.Watcher != nil {
		for _, dir := range WatchDirs {
			path := dir
			if s.RootDir != "" {
				path = filepath.Join(s.RootDir, dir)
			}
			if err := s.Watcher.Add(path); err != nil {
				fmt.Fprintf(os.Stderr, "warn: watch %s: %v\n", path, err)
			}
		}
		go s.watchLoop(broadcaster)
	}

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	return http.ListenAndServe(addr, mux)
}

// watchLoop listens for file change events and triggers rebuild + SSE broadcast.
// Events within debounceDelay are coalesced into a single rebuild to avoid
// multiple rapid reloads when an editor emits several write/rename events for
// one logical save.
const debounceDelay = 100 * time.Millisecond

func (s *DevServer) watchLoop(b *sseBroadcaster) {
	var pending string
	timer := time.NewTimer(0)
	<-timer.C // drain the initial tick so it doesn't fire immediately
	for {
		select {
		case path, ok := <-s.Watcher.Events():
			if !ok {
				return
			}
			pending = path
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(debounceDelay)
		case <-timer.C:
			if pending == "" {
				continue
			}
			if s.RebuildFunc != nil {
				_ = s.RebuildFunc() // best-effort; errors go to stderr via rebuild
			}
			b.broadcast(pending)
			pending = ""
		}
	}
}
