package internal

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer is a bytes.Buffer safe for a logger writing from one goroutine
// while the test reads from another.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// waitForLog polls buf until it contains want.
func waitForLog(t *testing.T, buf *syncBuffer, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(buf.String(), want) {
		if time.Now().After(deadline) {
			t.Fatalf("expected log to contain %q, got:\n%s", want, buf.String())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestParseLogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"error", slog.LevelError},
		{"warn", slog.LevelWarn},
		{"info", slog.LevelInfo},
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{" Info ", slog.LevelInfo},
	}
	for _, tt := range tests {
		got, err := ParseLogLevel(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("ParseLogLevel(%q) = %v, %v; want %v", tt.in, got, err, tt.want)
		}
	}
	for _, bad := range []string{"", "verbose", "warning", "trace", "1"} {
		_, err := ParseLogLevel(bad)
		if err == nil {
			t.Errorf("ParseLogLevel(%q): expected an error", bad)
			continue
		}
		if !strings.Contains(err.Error(), "error, warn, info, debug") {
			t.Errorf("ParseLogLevel(%q): error %q should list the valid levels", bad, err)
		}
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	t.Parallel()

	var info bytes.Buffer
	l := NewLogger(&info, slog.LevelInfo)
	l.Debug("hidden event", "path", "/a.md")
	l.Info("🚀 Starting server: http://localhost:1/")
	l.Error("❌ file watcher", "err", errors.New("boom"))
	want := "🚀 Starting server: http://localhost:1/\nERROR ❌ file watcher err=boom\n"
	if info.String() != want {
		t.Fatalf("info logger wrote %q, want %q", info.String(), want)
	}

	var debug bytes.Buffer
	d := NewLogger(&debug, slog.LevelDebug)
	d.Debug("fs event", "path", "/a b.md", "subscribers", 2)
	got := debug.String()
	// "15:04:05.000 DEBUG fs event ..."
	if len(got) < 13 || got[2] != ':' || got[12] != ' ' {
		t.Fatalf("expected a timestamp prefix at debug, got %q", got)
	}
	if !strings.HasSuffix(got, `DEBUG fs event path="/a b.md" subscribers=2`+"\n") {
		t.Fatalf("unexpected debug line %q", got)
	}

	var errOnly bytes.Buffer
	e := NewLogger(&errOnly, slog.LevelError)
	e.Info("banner")
	e.Warn("warning")
	if errOnly.Len() != 0 {
		t.Fatalf("error logger wrote %q", errOnly.String())
	}
}

func TestLoggerQuotesUnsafeValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"/plain.md", "/plain.md"},
		{"/a\nb", `"/a\nb"`},
		{"/a\x1b[31m", `"/a\x1b[31m"`},
		{"/\x9b31m", `"/\x9b31m"`}, // invalid UTF-8: a raw C1 CSI byte
		{"/caf\xc3", `"/caf\xc3"`},
		{"/café", "/café"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		NewLogger(&buf, slog.LevelInfo).Info("m", "path", tt.in)
		if got, want := buf.String(), "m path="+tt.want+"\n"; got != want {
			t.Errorf("path %q: got %q, want %q", tt.in, got, want)
		}
		if bytes.IndexByte(buf.Bytes(), 0x9b) >= 0 || bytes.IndexByte(buf.Bytes(), 0x1b) >= 0 {
			t.Errorf("path %q: raw control byte in output %q", tt.in, buf.String())
		}
	}
}

func TestLoggerAttrsAndGroups(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := NewLogger(&buf, slog.LevelInfo).With("a", 1).WithGroup("g")
	l.Info("msg", "b", "", slog.Group("h", "c", true))
	if got, want := buf.String(), "msg a=1 g.b=\"\" g.h.c=true\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
