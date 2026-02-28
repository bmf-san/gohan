// Package server implements the local development HTTP server with live reload.
package server

import (
"github.com/bmf-san/gohan/internal/model"
)

// Server is the local development HTTP server.  It serves static files from
// the build output directory and notifies connected browsers of file changes
// via a WebSocket-based live reload mechanism.
type Server interface {
// Start launches the HTTP server and file watcher.  It blocks until the
// DevServer configuration is exhausted or an unrecoverable error occurs.
Start(cfg model.DevServer) error

// Stop gracefully stops the HTTP server and the file watcher.
Stop() error

// Reload triggers a live-reload notification to all connected clients.
Reload() error
}
