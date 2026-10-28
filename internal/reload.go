package internal

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchSet is what one page depends on.
type WatchSet struct {
	Files []string // absolute OS paths; any event on them matches
	Dirs  []string // absolute OS paths; any non-temp event on a direct child matches
	Page  string   // the page's URL path; used in log lines only
}

// EventSource is a non-recursive folder watcher. *fsnotify.Watcher is
// adapted to it by newFsnotifySource.
type EventSource interface {
	Add(dir string) error
	Remove(dir string) error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// fsnotifySource adapts *fsnotify.Watcher, whose channels are fields, to
// EventSource.
type fsnotifySource struct {
	w *fsnotify.Watcher
}

func newFsnotifySource() (*fsnotifySource, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fsnotifySource{w: w}, nil
}

func (s *fsnotifySource) Add(dir string) error          { return s.w.Add(dir) }
func (s *fsnotifySource) Remove(dir string) error       { return s.w.Remove(dir) }
func (s *fsnotifySource) Events() <-chan fsnotify.Event { return s.w.Events }
func (s *fsnotifySource) Errors() <-chan error          { return s.w.Errors }

// Hub fans file system events out to the subscribers they are relevant to.
// It keeps one non-recursive watch per folder that at least one subscriber
// needs, and debounces notifications per subscriber.
type Hub struct {
	src      EventSource
	debounce time.Duration
	log      *slog.Logger
	nextID   atomic.Uint64

	// watchMu serializes folder reference counting and the matching
	// src.Add/src.Remove calls. It is never held together with mu.
	watchMu sync.Mutex
	refs    map[string]int
	failed  map[string]bool

	mu   sync.Mutex
	subs map[*subscriber]struct{}
}

type subscriber struct {
	id      uint64
	page    string
	files   map[string]bool
	dirs    map[string]bool
	folders []string
	notify  chan struct{}

	// Guarded by Hub.mu.
	timer   *time.Timer
	gen     uint64
	pending int // events coalesced into the next fire
}

// NewHub returns a hub reading from src. Call Run to start processing events.
// A nil logger discards all output.
func NewHub(src EventSource, debounce time.Duration, logger *slog.Logger) *Hub {
	return &Hub{
		src:      src,
		debounce: debounce,
		log:      orDiscard(logger),
		refs:     make(map[string]int),
		failed:   make(map[string]bool),
		subs:     make(map[*subscriber]struct{}),
	}
}

// Run processes events until ctx is done or the event channel is closed.
func (h *Hub) Run(ctx context.Context) {
	events := h.src.Events()
	errs := h.src.Errors()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			h.handle(ev)
		case err, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			h.log.Error("❌ file watcher", "err", err)
		}
	}
}

// Subscribe registers ws and returns a channel that receives a value, at
// most once per debounce window, when a relevant file changes. cancel
// unregisters the subscriber and releases its folder watches; it is safe to
// call more than once.
func (h *Hub) Subscribe(ws WatchSet) (notify <-chan struct{}, cancel func()) {
	sub := &subscriber{
		id:     h.nextID.Add(1),
		page:   ws.Page,
		files:  make(map[string]bool),
		dirs:   make(map[string]bool),
		notify: make(chan struct{}, 1),
	}
	folders := make(map[string]bool)
	for _, f := range ws.Files {
		f = filepath.Clean(f)
		sub.files[f] = true
		folders[filepath.Dir(f)] = true
	}
	for _, d := range ws.Dirs {
		d = filepath.Clean(d)
		sub.dirs[d] = true
		folders[d] = true
	}
	for d := range folders {
		sub.folders = append(sub.folders, d)
	}

	h.acquire(sub.folders)

	h.mu.Lock()
	h.subs[sub] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	return sub.notify, func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subs, sub)
			if sub.timer != nil {
				sub.timer.Stop()
			}
			h.mu.Unlock()
			h.release(sub.folders)
		})
	}
}

func (h *Hub) acquire(folders []string) {
	h.watchMu.Lock()
	defer h.watchMu.Unlock()
	for _, d := range folders {
		h.refs[d]++
		if h.refs[d] != 1 {
			h.log.Debug("watch shared", "dir", d, "refs", h.refs[d])
			continue
		}
		if err := h.src.Add(d); err != nil {
			if !h.failed[d] {
				h.failed[d] = true
				h.log.Warn("❌ cannot watch folder, live reload is off for it", "dir", d, "err", err)
			}
			continue
		}
		h.log.Debug("watch added", "dir", d, "refs", 1)
	}
}

func (h *Hub) release(folders []string) {
	h.watchMu.Lock()
	defer h.watchMu.Unlock()
	for _, d := range folders {
		h.refs[d]--
		if h.refs[d] > 0 {
			h.log.Debug("watch released", "dir", d, "refs", h.refs[d])
			continue
		}
		delete(h.refs, d)
		// Removing a folder whose Add failed reports an error; ignore it.
		_ = h.src.Remove(d)
		h.log.Debug("watch removed", "dir", d, "refs", 0)
	}
}

// debugging reports whether debug logging is on. Checking it first keeps the
// disabled hot path to a single branch with no allocation.
func (h *Hub) debugging() bool {
	return h.log.Enabled(context.Background(), slog.LevelDebug)
}

func (h *Hub) handle(ev fsnotify.Event) {
	debug := h.debugging()
	// Attribute-only changes (chmod, touch on some systems, indexers) do not
	// change what a page shows.
	if ev.Op == fsnotify.Chmod {
		if debug {
			h.log.Debug("fs event ignored", "op", ev.Op.String(), "path", ev.Name, "reason", "chmod")
		}
		return
	}
	name := filepath.Clean(ev.Name)
	parent := filepath.Dir(name)
	temp := isTempName(filepath.Base(name))

	// Collected under the lock and logged after releasing it, so a slow
	// stderr never holds up the hub.
	var started []*subscriber
	matched := 0

	h.mu.Lock()
	for sub := range h.subs {
		if !sub.files[name] && (temp || !sub.dirs[parent]) {
			continue
		}
		matched++
		if h.schedule(sub) && debug {
			started = append(started, sub)
		}
	}
	h.mu.Unlock()

	if !debug {
		return
	}
	switch {
	case matched > 0:
		h.log.Debug("fs event", "op", ev.Op.String(), "path", name, "subscribers", matched)
	case temp:
		h.log.Debug("fs event ignored", "op", ev.Op.String(), "path", name, "reason", "temp name")
	default:
		h.log.Debug("fs event ignored", "op", ev.Op.String(), "path", name, "reason", "no subscriber matched")
	}
	for _, sub := range started {
		h.log.Debug("debounce start", "sub", sub.id, "page", sub.page, "wait", h.debounce)
	}
}

// schedule (re)starts sub's debounce timer and reports whether this event
// opened a new debounce window. The caller holds h.mu.
func (h *Hub) schedule(sub *subscriber) bool {
	sub.gen++
	sub.pending++
	gen := sub.gen
	if sub.timer != nil {
		sub.timer.Stop()
	}
	sub.timer = time.AfterFunc(h.debounce, func() { h.fire(sub, gen) })
	return sub.pending == 1
}

func (h *Hub) fire(sub *subscriber, gen uint64) {
	h.mu.Lock()
	_, live := h.subs[sub]
	current := sub.gen == gen
	coalesced := 0
	if current {
		coalesced = sub.pending
		sub.pending = 0
	}
	h.mu.Unlock()
	if !live || !current {
		return
	}
	select {
	case sub.notify <- struct{}{}:
		if h.debugging() {
			h.log.Debug("debounce fire", "sub", sub.id, "page", sub.page, "coalesced", coalesced)
		}
	default:
		if h.debugging() {
			h.log.Debug("notification dropped, one already pending", "sub", sub.id, "page", sub.page, "coalesced", coalesced)
		}
	}
}

// isTempName reports editor temp and swap file names: Emacs lock files
// (.#name), backup files (name~), Vim swap files (.swp, .swx) and Vim's
// write-permission probe (4913).
func isTempName(base string) bool {
	return strings.HasPrefix(base, ".#") ||
		strings.HasSuffix(base, "~") ||
		strings.HasSuffix(base, ".swp") ||
		strings.HasSuffix(base, ".swx") ||
		base == "4913"
}
