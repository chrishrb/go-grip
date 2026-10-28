package internal

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// loggedHub starts a hub on a fake source that logs to a buffer at level.
func loggedHub(t *testing.T, debounce time.Duration, level slog.Level) (*Hub, *fakeSource, *syncBuffer) {
	t.Helper()
	buf := &syncBuffer{}
	src := newFakeSource()
	hub := NewHub(src, debounce, NewLogger(buf, level))
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
	return hub, src, buf
}

func TestHubDebugLogsCoalescedBurst(t *testing.T) {
	t.Parallel()
	hub, src, buf := loggedHub(t, 100*time.Millisecond, slog.LevelDebug)

	notify, cancel := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md")}, Page: "/a.md"})
	defer cancel()

	for range 5 {
		src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Write}
	}
	expectNotify(t, notify)
	waitForLog(t, buf, "debounce fire sub=1 page=/a.md coalesced=5")

	out := buf.String()
	if n := strings.Count(out, "debounce start"); n != 1 {
		t.Fatalf("expected one debounce start for the burst, got %d:\n%s", n, out)
	}
	if n := strings.Count(out, "debounce fire"); n != 1 {
		t.Fatalf("expected one debounce fire for the burst, got %d:\n%s", n, out)
	}
	if !strings.Contains(out, "watch added dir="+osPath("r")+" refs=1") {
		t.Fatalf("expected the folder watch to be logged:\n%s", out)
	}
}

func TestHubDebugLogsIgnoredEvents(t *testing.T) {
	t.Parallel()
	hub, src, buf := loggedHub(t, testDebounce, slog.LevelDebug)

	_, cancel := hub.Subscribe(WatchSet{Dirs: []string{osPath("r")}, Page: "/"})
	defer cancel()

	src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Chmod}
	src.events <- fsnotify.Event{Name: osPath("r", ".a.md.swp"), Op: fsnotify.Write}
	src.events <- fsnotify.Event{Name: osPath("elsewhere", "b.md"), Op: fsnotify.Write}

	waitForLog(t, buf, "reason=chmod")
	waitForLog(t, buf, `reason="temp name"`)
	waitForLog(t, buf, `reason="no subscriber matched"`)
}

func TestHubDebugLogsDroppedNotification(t *testing.T) {
	t.Parallel()
	hub, src, buf := loggedHub(t, testDebounce, slog.LevelDebug)

	_, cancel := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md")}, Page: "/a.md"})
	defer cancel()

	// Nobody reads notify, so the second fire finds the slot taken.
	src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Write}
	waitForLog(t, buf, "debounce fire")
	src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Write}
	waitForLog(t, buf, "notification dropped, one already pending sub=1 page=/a.md coalesced=1")
}

func TestHubInfoLogsNoEvents(t *testing.T) {
	t.Parallel()
	hub, src, buf := loggedHub(t, testDebounce, slog.LevelInfo)

	notify, cancel := hub.Subscribe(WatchSet{Files: []string{osPath("r", "a.md")}})
	defer cancel()
	src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Write}
	src.events <- fsnotify.Event{Name: osPath("r", "a.md"), Op: fsnotify.Chmod}
	expectNotify(t, notify)
	cancel()
	if out := buf.String(); out != "" {
		t.Fatalf("expected no output at info, got:\n%s", out)
	}
}
