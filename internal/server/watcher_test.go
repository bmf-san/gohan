package server

import (
	"testing"
	"time"
)

// fakeWatcher is a mock FileWatcher for unit tests.
type fakeWatcher struct {
	ch     chan string
	closed bool
}

func newFakeWatcher() *fakeWatcher {
	return &fakeWatcher{ch: make(chan string, 10)}
}

func (f *fakeWatcher) Add(path string) error   { return nil }
func (f *fakeWatcher) Events() <-chan string    { return f.ch }
func (f *fakeWatcher) Close() error            { close(f.ch); return nil }

func TestNewFsnotifyWatcher_CreatesWatcher(t *testing.T) {
	fw, err := NewFsnotifyWatcher()
	if err != nil {
		t.Fatalf("NewFsnotifyWatcher: %v", err)
	}
	defer fw.Close()

	if fw.Events() == nil {
		t.Error("expected non-nil events channel")
	}
}

func TestFsnotifyWatcher_AddAndClose(t *testing.T) {
	fw, err := NewFsnotifyWatcher()
	if err != nil {
		t.Fatalf("NewFsnotifyWatcher: %v", err)
	}
	// Add a real path; errors are acceptable (e.g. missing dir).
	_ = fw.Add(".")
	if err := fw.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestWatchLoop_BroadcastsEvent(t *testing.T) {
	b := newSSEBroadcaster()
	ch := b.subscribe()
	defer b.unsubscribe(ch)

	fw := newFakeWatcher()
	srv := &DevServer{
		Watcher:     fw,
		RebuildFunc: func() error { return nil },
	}

	go srv.watchLoop(b)

	fw.ch <- "/content/posts/hello.md"

	select {
	case msg := <-ch:
		if msg != "/content/posts/hello.md" {
			t.Errorf("expected path broadcasted, got: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: watchLoop did not broadcast event")
	}
}

func TestWatchLoop_CallsRebuildFunc(t *testing.T) {
	b := newSSEBroadcaster()

	rebuilt := make(chan struct{}, 1)
	fw := newFakeWatcher()
	srv := &DevServer{
		Watcher: fw,
		RebuildFunc: func() error {
			rebuilt <- struct{}{}
			return nil
		},
	}

	go srv.watchLoop(b)

	fw.ch <- "somefile.md"

	select {
	case <-rebuilt:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: RebuildFunc not called")
	}
}

func TestWatchLoop_NilRebuildFunc(t *testing.T) {
	b := newSSEBroadcaster()
	ch := b.subscribe()
	defer b.unsubscribe(ch)

	fw := newFakeWatcher()
	srv := &DevServer{
		Watcher:     fw,
		RebuildFunc: nil, // should not panic
	}

	go srv.watchLoop(b)

	fw.ch <- "somefile.md"

	select {
	case <-ch:
		// success — event was broadcasted even without rebuild func
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: watchLoop with nil RebuildFunc")
	}
}

func TestWatchLoop_ExitsOnChannelClose(t *testing.T) {
	b := newSSEBroadcaster()
	fw := newFakeWatcher()
	srv := &DevServer{Watcher: fw}

	done := make(chan struct{})
	go func() {
		srv.watchLoop(b)
		close(done)
	}()

	fw.Close() // closes the events channel → watchLoop should return

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: watchLoop did not exit after channel close")
	}
}
