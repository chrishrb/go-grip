package internal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

const testDebounce = 20 * time.Millisecond

// fakeSource is an EventSource driven by the test.
type fakeSource struct {
	events chan fsnotify.Event
	errs   chan error
	addErr error

	mu      sync.Mutex
	added   []string
	removed []string
}

func newFakeSource() *fakeSource {
	return &fakeSource{
		events: make(chan fsnotify.Event),
		errs:   make(chan error),
	}
}

func (f *fakeSource) Add(dir string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.added = append(f.added, dir)
	return f.addErr
}

func (f *fakeSource) Remove(dir string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removed = append(f.removed, dir)
	return nil
}

func (f *fakeSource) Events() <-chan fsnotify.Event { return f.events }
func (f *fakeSource) Errors() <-chan error          { return f.errs }

func (f *fakeSource) calls() (added, removed []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.added...), append([]string(nil), f.removed...)
}

// runHub starts a hub on a fake source for the duration of the test.
func runHub(t *testing.T) (*Hub, *fakeSource) {
	t.Helper()
	src := newFakeSource()
	hub := NewHub(src, testDebounce, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return hub, src
}

// osPath builds an absolute OS path from its elements.
func osPath(parts ...string) string {
	return filepath.Join(append([]string{string(filepath.Separator)}, parts...)...)
}

func expectNotify(t *testing.T, notify <-chan struct{}) {
	t.Helper()
	select {
	case <-notify:
	case <-time.After(2 * time.Second):
		t.Fatal("expected a notification")
	}
}

func expectNoNotify(t *testing.T, notify <-chan struct{}) {
	t.Helper()
	select {
	case <-notify:
		t.Fatal("unexpected notification")
	case <-time.After(5 * testDebounce):
	}
}

func TestHubMatchesFile(t *testing.T) {
	t.Parallel()
	hub, src := runHub(t)

	notify, cancel := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md")}})
	defer cancel()

	src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Write}
	expectNotify(t, notify)
	expectNoNotify(t, notify)
}

func TestHubIgnoresOtherFile(t *testing.T) {
	t.Parallel()
	hub, src := runHub(t)

	notify, cancel := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md")}})
	defer cancel()

	src.events <- fsnotify.Event{Name: osPath("r", "b.md"), Op: fsnotify.Write}
	src.events <- fsnotify.Event{Name: osPath("r", "a.go"), Op: fsnotify.Write}
	src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Chmod}
	expectNoNotify(t, notify)
}

func TestHubDirEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []fsnotify.Event
		want   bool
	}{
		{"create", []fsnotify.Event{{Name: osPath("r", "new.md"), Op: fsnotify.Create}}, true},
		{"remove", []fsnotify.Event{{Name: osPath("r", "old.md"), Op: fsnotify.Remove}}, true},
		{"rename", []fsnotify.Event{{Name: osPath("r", "old.md"), Op: fsnotify.Rename}}, true},
		{"temp names ignored", []fsnotify.Event{
			{Name: osPath("r", ".#a.md"), Op: fsnotify.Create},
			{Name: osPath("r", "a.md~"), Op: fsnotify.Create},
			{Name: osPath("r", ".a.md.swp"), Op: fsnotify.Write},
			{Name: osPath("r", ".a.md.swx"), Op: fsnotify.Create},
			{Name: osPath("r", "4913"), Op: fsnotify.Create},
		}, false},
		{"nested ignored", []fsnotify.Event{{Name: osPath("r", "sub", "x.md"), Op: fsnotify.Create}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hub, src := runHub(t)

			notify, cancel := hub.Subscribe(WatchSet{Dirs: []string{osPath("r")}})
			defer cancel()

			for _, ev := range tt.events {
				src.events <- ev
			}
			if tt.want {
				expectNotify(t, notify)
			} else {
				expectNoNotify(t, notify)
			}
		})
	}
}

func TestHubDebounce(t *testing.T) {
	t.Parallel()
	src := newFakeSource()
	hub := NewHub(src, 100*time.Millisecond, nil)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go hub.Run(ctx)

	notify, cancel := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md")}})
	defer cancel()

	for range 5 {
		src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Write}
	}
	expectNotify(t, notify)
	select {
	case <-notify:
		t.Fatal("expected exactly one notification for a burst")
	case <-time.After(300 * time.Millisecond):
	}
}

func TestHubRefCountsWatches(t *testing.T) {
	t.Parallel()
	hub, src := runHub(t)

	_, cancelA := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md"), osPath("r", "img", "x.png")}})
	_, cancelB := hub.Subscribe(WatchSet{Dirs: []string{osPath("r")}})

	added, removed := src.calls()
	if len(added) != 2 || len(removed) != 0 {
		t.Fatalf("after two subscribers: added %q, removed %q", added, removed)
	}

	cancelA()
	cancelA() // idempotent
	_, removed = src.calls()
	if !reflect.DeepEqual(removed, []string{osPath("r", "img")}) {
		t.Fatalf("after first cancel: removed %q, want only the img folder", removed)
	}

	cancelB()
	_, removed = src.calls()
	if len(removed) != 2 || removed[1] != osPath("r") {
		t.Fatalf("after last cancel: removed %q", removed)
	}
}

func TestHubAddErrorStillSubscribes(t *testing.T) {
	t.Parallel()
	src := newFakeSource()
	src.addErr = errors.New("no space left on device")
	hub := NewHub(src, testDebounce, nil)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go hub.Run(ctx)

	notify, cancel := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md")}})
	defer cancel()
	if notify == nil {
		t.Fatal("expected a channel even when the watch could not be added")
	}
	if len(hub.failed) != 1 {
		t.Fatalf("expected the failed folder to be recorded once, got %v", hub.failed)
	}
}

func TestHubFsnotifyAtomicSave(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "a.md")
	if err := os.WriteFile(target, []byte("# one\n"), 0o644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}

	src, err := newFsnotifySource()
	if err != nil {
		t.Fatalf("new watcher: %v", err)
	}
	t.Cleanup(func() { _ = src.w.Close() })

	hub := NewHub(src, testDebounce, nil)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go hub.Run(ctx)

	notify, cancel := hub.Subscribe(WatchSet{Files: []string{target}})
	defer cancel()

	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, []byte("# two\n"), 0o644); err != nil {
		t.Fatalf("write a.md.tmp: %v", err)
	}
	if err := os.Rename(tmp, target); err != nil {
		t.Fatalf("rename: %v", err)
	}
	expectNotify(t, notify)
}
