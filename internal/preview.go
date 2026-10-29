package internal

import (
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

// Previewer serves Markdown files on demand, for callers such as the MCP
// server that preview files from many folders over one process lifetime.
// Each previewed folder gets its own localhost server on a free port, unless
// a folder above it is already served.
type Previewer struct {
	mu     sync.Mutex
	server *Server
	roots  map[string]string // absolute served dir -> base URL without trailing slash
}

// NewPreviewer returns a previewer that listens on localhost only. A nil
// logger discards all output.
func NewPreviewer(boundingBox bool, enableReload bool, parser *Parser, logger *slog.Logger) *Previewer {
	return &Previewer{
		server: NewServer("localhost", 0, boundingBox, false, enableReload, parser, logger),
		roots:  map[string]string{},
	}
}

// Preview makes path (a Markdown file or a folder) reachable in the browser
// and returns its URL. With open set it also opens the URL.
func (p *Previewer) Preview(path string, open bool) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	dir, startPath, err := ResolveTarget(abs)
	if err != nil {
		return "", err
	}

	p.mu.Lock()
	root, base, ok := p.servedRoot(dir)
	if !ok {
		addr, _, err := p.server.Start(dir)
		if err != nil {
			p.mu.Unlock()
			return "", err
		}
		root, base = dir, strings.TrimSuffix(addr, "/")
		p.roots[root] = base
	}
	p.mu.Unlock()

	addr := base + urlPathFor(root, abs)
	if startPath == "/" && !strings.HasSuffix(addr, "/") {
		// a folder: keep the slash so relative links on its page resolve
		addr += "/"
	}
	if open {
		if err := Open(addr); err != nil {
			p.server.log.Warn("❌ Error opening browser", "err", err)
		}
	}
	return addr, nil
}

// servedRoot returns the served root that contains dir, if any. Callers hold
// p.mu.
func (p *Previewer) servedRoot(dir string) (root, base string, ok bool) {
	for r, b := range p.roots {
		rel, err := filepath.Rel(r, dir)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			// Prefer the deepest root, so URLs stay short and stable.
			if !ok || len(r) > len(root) {
				root, base, ok = r, b, true
			}
		}
	}
	return root, base, ok
}

// urlPathFor returns the escaped URL path of target inside root.
func urlPathFor(root, target string) string {
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." {
		return "/"
	}
	segments := strings.Split(filepath.ToSlash(rel), "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return "/" + strings.Join(segments, "/")
}
