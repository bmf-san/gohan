package server

import (
	"fmt"
	"net/http"
)

// FileWatcher is the file change detection interface. The implementation uses fsnotify.
type FileWatcher interface {
	Add(path string) error
	Events() <-chan string
	Close() error
}

// DevServer is a local HTTP development server.
type DevServer struct {
	Host    string
	Port    int
	OutDir  string
	Watcher FileWatcher
}

// NewDevServer creates a new DevServer.
func NewDevServer(host string, port int, outDir string) *DevServer {
	return &DevServer{
		Host:   host,
		Port:   port,
		OutDir: outDir,
	}
}

// Start starts the development server.
// Phase 9 expands this with fsnotify file watching and SSE live reload.
func (s *DevServer) Start() error {
	fs := http.FileServer(http.Dir(s.OutDir))
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), mux)
}
