package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// LogLevels lists the accepted --log-level values, most to least severe.
var LogLevels = []string{"error", "warn", "info", "debug"}

// ParseLogLevel maps a --log-level value (case-insensitive) to a slog level.
func ParseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "error":
		return slog.LevelError, nil
	case "warn":
		return slog.LevelWarn, nil
	case "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	}
	return 0, fmt.Errorf("invalid log level %q: want one of %s", s, strings.Join(LogLevels, ", "))
}

// NewLogger returns a logger writing human-readable lines to w. Info lines
// are the bare message plus attributes, so the startup banner reads as it
// always has; other levels are prefixed with the level. At debug every line
// also starts with a millisecond timestamp, to show event and reload cadence.
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(&lineHandler{
		w:     w,
		mu:    &sync.Mutex{},
		level: level,
		stamp: level <= slog.LevelDebug,
	})
}

// discardLogger is used when a constructor is given a nil logger.
func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func orDiscard(l *slog.Logger) *slog.Logger {
	if l == nil {
		return discardLogger()
	}
	return l
}

// lineHandler is a small slog.Handler producing one plain line per record:
//
//	[15:04:05.000 ][LEVEL ]message key=value ...
type lineHandler struct {
	w     io.Writer
	mu    *sync.Mutex
	level slog.Level
	stamp bool

	attrs  string // preformatted attributes from WithAttrs
	prefix string // group prefix from WithGroup, "a.b."
}

func (h *lineHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *lineHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer
	if h.stamp {
		t := r.Time
		if t.IsZero() {
			t = time.Now()
		}
		buf.WriteString(t.Format("15:04:05.000"))
		buf.WriteByte(' ')
	}
	if r.Level != slog.LevelInfo {
		buf.WriteString(r.Level.String())
		buf.WriteByte(' ')
	}
	buf.WriteString(r.Message)
	buf.WriteString(h.attrs)
	r.Attrs(func(a slog.Attr) bool {
		appendAttr(&buf, h.prefix, a)
		return true
	})
	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *lineHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var buf bytes.Buffer
	for _, a := range attrs {
		appendAttr(&buf, h.prefix, a)
	}
	h2 := *h
	h2.attrs = h.attrs + buf.String()
	return &h2
}

func (h *lineHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.prefix = h.prefix + name + "."
	return &h2
}

func appendAttr(buf *bytes.Buffer, prefix string, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		p := prefix
		if a.Key != "" {
			p += a.Key + "."
		}
		for _, ga := range a.Value.Group() {
			appendAttr(buf, p, ga)
		}
		return
	}
	buf.WriteByte(' ')
	buf.WriteString(prefix)
	buf.WriteString(a.Key)
	buf.WriteByte('=')
	buf.WriteString(quoteIfNeeded(a.Value.String()))
}

func quoteIfNeeded(s string) string {
	// Invalid UTF-8 ranges as U+FFFD, which IsPrint accepts; quote it so a
	// raw byte such as 0x9b (C1 CSI) never reaches the terminal.
	if s == "" || !utf8.ValidString(s) {
		return strconv.Quote(s)
	}
	for _, r := range s {
		if unicode.IsSpace(r) || r == '"' || r == '=' || !unicode.IsPrint(r) {
			return strconv.Quote(s)
		}
	}
	return s
}
