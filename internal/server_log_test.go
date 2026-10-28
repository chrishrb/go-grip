package internal

import (
	"bufio"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestRequestLogQuietAtInfo(t *testing.T) {
	t.Parallel()
	dir := writeTree(t, map[string]string{"a.md": "# A\n"})
	buf := &syncBuffer{}
	server := NewServer("localhost", 6419, false, false, false, NewParser(), NewLogger(buf, slog.LevelInfo))
	handler := server.newHandler(http.Dir(dir))

	for _, p := range []string{"/a.md", "/", "/static/js/live-reload.js", "/missing.txt"} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, p, nil))
	}
	if out := buf.String(); out != "" {
		t.Fatalf("expected no request output at info, got:\n%s", out)
	}
}

func TestRequestLogAtDebug(t *testing.T) {
	t.Parallel()
	dir := writeTree(t, map[string]string{"a.md": "# A\n"})
	buf := &syncBuffer{}
	server := NewServer("localhost", 6419, false, false, false, NewParser(), NewLogger(buf, slog.LevelDebug))
	handler := server.newHandler(http.Dir(dir))

	tests := []struct {
		path   string
		status int
	}{
		{"/a.md", http.StatusOK},
		{"/missing.txt", http.StatusNotFound},
		{"/static/js/live-reload.js", http.StatusOK},
		{"/sub", http.StatusNotFound},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rec.Code != tt.status {
			t.Fatalf("%s: status %d, want %d", tt.path, rec.Code, tt.status)
		}
		re := regexp.MustCompile(`DEBUG request method=GET path=` + regexp.QuoteMeta(tt.path) +
			` status=` + strconv.Itoa(tt.status) + ` bytes=\d+ dur=[0-9.]+[µnm]?s\n`)
		if !re.MatchString(buf.String()) {
			t.Fatalf("%s: no request line with status %d in:\n%s", tt.path, tt.status, buf.String())
		}
	}
}

func TestStatusRecorder(t *testing.T) {
	t.Parallel()

	inner := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: inner}
	rec.WriteHeader(http.StatusTeapot)
	rec.WriteHeader(http.StatusOK) // superfluous; the first code wins
	_, _ = rec.Write([]byte("hello"))
	rec.Flush()
	if rec.status != http.StatusTeapot || rec.bytes != 5 {
		t.Fatalf("recorded status %d, bytes %d", rec.status, rec.bytes)
	}
	if !inner.Flushed {
		t.Fatal("Flush was not forwarded")
	}
	var _ http.Flusher = rec

	implicit := &statusRecorder{ResponseWriter: httptest.NewRecorder()}
	_, _ = implicit.Write([]byte("x"))
	if implicit.status != http.StatusOK {
		t.Fatalf("implicit status %d, want 200", implicit.status)
	}
}

func TestEventsStreamWithDebugLogging(t *testing.T) {
	t.Parallel()

	dir := writeTree(t, map[string]string{"docs/a.md": "# A\n"})
	hub, src, hubLog := loggedHub(t, testDebounce, slog.LevelDebug)
	buf := &syncBuffer{}
	server := NewServer("localhost", 6419, false, false, true, NewParser(), NewLogger(buf, slog.LevelDebug))
	server.hub = hub
	ts := httptest.NewServer(server.newHandler(http.Dir(dir)))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+eventsPath+"?path="+url.QueryEscape("/docs/a.md"), nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	reader := bufio.NewReader(resp.Body)
	if line, err := reader.ReadString('\n'); err != nil || line != ": connected\n" {
		t.Fatalf("expected the connected comment first, got %q (%v)", line, err)
	}
	waitForLog(t, buf, "sse connect page=/docs/a.md")
	waitForLog(t, buf, "watch set page=/docs/a.md files=1 dirs=0 took=")

	src.events <- fsnotify.Event{Name: filepath.Join(dir, "docs", "a.md"), Op: fsnotify.Write}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if line == "event: reload\n" {
			break
		}
	}
	waitForLog(t, buf, "sse reload sent page=/docs/a.md")
	waitForLog(t, hubLog, "debounce fire sub=1 page=/docs/a.md coalesced=1")

	cancel()
	waitForLog(t, buf, "sse disconnect page=/docs/a.md reloads=1 dur=")
	waitForLog(t, buf, "request method=GET path="+eventsPath+" status=200")
}

func TestEventsRejectionLogsWarning(t *testing.T) {
	t.Parallel()
	dir := writeTree(t, map[string]string{"a.md": "# A\n"})
	hub, _, _ := loggedHub(t, testDebounce, slog.LevelInfo)
	buf := &syncBuffer{}
	server := NewServer("localhost", 6419, false, false, true, NewParser(), NewLogger(buf, slog.LevelWarn))
	server.hub = hub
	handler := server.newHandler(http.Dir(dir))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, eventsPath+"?path=/../x.md", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
	if got, want := buf.String(), "WARN live reload request rejected page=/../x.md\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
