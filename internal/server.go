package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	chroma_html "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/chrishrb/go-grip/defaults"
)

const defaultHTMLTitle = "go-grip - markdown preview"

type Server struct {
	parser       *Parser
	boundingBox  bool
	host         string
	port         int
	browser      bool
	enableReload bool
	hub          *Hub
	log          *slog.Logger
}

// reloadDebounce collapses the several events an editor writes per save into
// one browser reload.
const reloadDebounce = 100 * time.Millisecond

// eventsPath is the live reload stream endpoint.
const eventsPath = "/__grip/events"

// reloading reports whether live reload is on: requested and backed by a
// running hub.
func (s *Server) reloading() bool {
	return s.enableReload && s.hub != nil
}

// NewServer returns a server. A nil logger discards all output.
func NewServer(host string, port int, boundingBox bool, browser bool, enableReload bool, parser *Parser, logger *slog.Logger) *Server {
	return &Server{
		host:         host,
		port:         port,
		boundingBox:  boundingBox,
		browser:      browser,
		enableReload: enableReload,
		parser:       parser,
		log:          orDiscard(logger),
	}
}

func (s *Server) Serve(file string) error {
	directory, startPath, err := ResolveTarget(file)
	if err != nil {
		return err
	}

	if s.enableReload {
		src, err := newFsnotifySource()
		if err != nil {
			s.log.Warn("❌ Error starting file watcher, auto-reload disabled", "err", err)
			s.enableReload = false
		} else {
			s.hub = NewHub(src, reloadDebounce, s.log)
			go s.hub.Run(context.Background())
		}
	}

	handler := s.newHandler(http.Dir(directory))

	addr := fmt.Sprintf("http://%s:%d%s", s.host, s.port, startPath)

	s.log.Info("🚀 Starting server: " + addr)
	s.log.Debug("serving directory", "dir", directory)

	if s.browser {
		err := Open(addr)
		if err != nil {
			s.log.Warn("❌ Error opening browser", "err", err)
		}
	}

	if s.reloading() {
		s.log.Info("📡 Auto-reload enabled. Files will trigger browser refresh.")
	} else {
		s.log.Info("🔄 Auto-reload disabled. Use F5 to manually refresh.")
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler)
}

// debugging reports whether debug logging is on.
func (s *Server) debugging(ctx context.Context) bool {
	return s.log.Enabled(ctx, slog.LevelDebug)
}

// statusRecorder captures the status code and body size of a response. It
// forwards Flush, which serveEvents needs to stream.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (rec *statusRecorder) WriteHeader(code int) {
	if rec.status == 0 {
		rec.status = code
	}
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += int64(n)
	return n, err
}

func (rec *statusRecorder) Flush() {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}
	if f, ok := rec.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (rec *statusRecorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}

// logRequests logs each request's method, path, status and duration at
// debug level, after it completes. With debug off it adds one branch.
func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.debugging(r.Context()) {
			next.ServeHTTP(w, r)
			return
		}
		rec := &statusRecorder{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(rec, r)
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		s.log.Debug("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", rec.bytes,
			"dur", time.Since(start))
	})
}

func (s *Server) newHandler(fsys http.FileSystem) http.Handler {
	root := NewRoot(fsys)
	fileServer := http.FileServer(fsys)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(defaults.StaticFiles)))

	if s.reloading() {
		absRoot := ""
		if dir, ok := fsys.(http.Dir); ok {
			if abs, err := filepath.Abs(string(dir)); err == nil {
				absRoot = abs
			}
		}
		mux.HandleFunc(eventsPath, func(w http.ResponseWriter, r *http.Request) {
			s.serveEvents(w, r, root, absRoot)
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch root.Classify(r.URL.Path) {
		case KindMarkdown:
			s.serveMarkdown(w, r, root)
			return
		case KindDir:
			s.serveFolder(w, r, root)
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	if !s.reloading() {
		return s.logRequests(mux)
	}
	// While reload is on, make the browser revalidate everything, so a
	// reload never shows a heuristically cached image or stylesheet.
	return s.logRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		mux.ServeHTTP(w, r)
	}))
}

// serveEvents streams Server-Sent Events to one page: a "reload" event each
// time something that page shows has changed. The page names itself in the
// path query parameter.
func (s *Server) serveEvents(w http.ResponseWriter, r *http.Request, root *Root, absRoot string) {
	page := r.URL.Query().Get("path")
	start := time.Now()
	ws, ok := s.watchSetFor(root, absRoot, page)
	if !ok {
		s.log.Warn("live reload request rejected", "page", page)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		s.serverError(w, r, errors.New("response writer does not support streaming"))
		return
	}
	ws.Page = page
	debug := s.debugging(r.Context())
	if debug {
		s.log.Debug("watch set",
			"page", page,
			"files", len(ws.Files),
			"dirs", len(ws.Dirs),
			"took", time.Since(start))
	}

	notify, cancel := s.hub.Subscribe(ws)
	defer cancel()

	connected := time.Now()
	reloads := 0
	if debug {
		s.log.Debug("sse connect", "page", page)
		defer func() {
			s.log.Debug("sse disconnect", "page", page, "reloads", reloads, "dur", time.Since(connected))
		}()
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-notify:
			if _, err := io.WriteString(w, "event: reload\ndata: 1\n\n"); err != nil {
				if debug {
					s.log.Debug("sse reload write failed", "page", page, "err", err)
				}
				return
			}
			flusher.Flush()
			reloads++
			if debug {
				s.log.Debug("sse reload sent", "page", page)
			}
		}
	}
}

// watchSetFor computes what the page at urlPath depends on, as absolute OS
// paths under absRoot. A Markdown page depends on its file and the local
// images it references; a folder page on the folder's direct entries and the
// images its README references. urlPath is only a lookup key: a path with a
// ".." segment, or one that is neither Markdown nor a folder, is rejected.
func (s *Server) watchSetFor(root *Root, absRoot, urlPath string) (WatchSet, bool) {
	if absRoot == "" || !strings.HasPrefix(urlPath, "/") || strings.ContainsRune(urlPath, 0) {
		return WatchSet{}, false
	}
	if filepath.Separator != '/' && strings.ContainsRune(urlPath, filepath.Separator) {
		return WatchSet{}, false
	}
	for _, seg := range strings.Split(urlPath, "/") {
		if seg == ".." {
			return WatchSet{}, false
		}
	}
	clean := path.Clean(urlPath)
	own, ok := toOSPath(absRoot, clean)
	if !ok {
		return WatchSet{}, false
	}

	var ws WatchSet
	var mdPath, refDir string
	switch root.Classify(clean) {
	case KindMarkdown:
		ws.Files = append(ws.Files, own)
		mdPath, refDir = clean, path.Dir(clean)
	case KindDir:
		ws.Dirs = append(ws.Dirs, own)
		if name, ok := root.Readme(clean); ok {
			mdPath, refDir = path.Join(clean, name), clean
		}
	default:
		return WatchSet{}, false
	}

	if mdPath != "" {
		if content, err := root.ReadFile(mdPath); err == nil {
			for _, ref := range s.parser.LocalRefs(content) {
				rel, ok := resolveRef(refDir, ref)
				if !ok {
					continue
				}
				if p, ok := toOSPath(absRoot, "/"+rel); ok {
					ws.Files = append(ws.Files, p)
				}
			}
		}
	}
	return ws, true
}

// resolveRef resolves an image reference against the URL folder dir and
// returns it relative to the root, or false if it would leave the root.
func resolveRef(dir, ref string) (string, bool) {
	var rel string
	if strings.HasPrefix(ref, "/") {
		rel = path.Clean(strings.TrimLeft(ref, "/"))
	} else {
		rel = path.Join(strings.TrimLeft(dir, "/"), ref)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return rel, true
}

// toOSPath maps a cleaned URL path to an OS path under absRoot, or reports
// false if the result would not stay inside absRoot.
func toOSPath(absRoot, urlPath string) (string, bool) {
	p := filepath.Join(absRoot, filepath.FromSlash(strings.TrimPrefix(urlPath, "/")))
	rel, err := filepath.Rel(absRoot, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return p, true
}

func (s *Server) serveMarkdown(w http.ResponseWriter, r *http.Request, root *Root) {
	setNoCacheHeaders(w)

	content, err := root.ReadFile(r.URL.Path)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	htmlContent, err := s.parser.MdToHTML(content)
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	var buf bytes.Buffer
	err = renderTemplate(&buf, htmlStruct{
		Content:      string(htmlContent),
		BoundingBox:  s.boundingBox,
		CssCodeLight: getCssCode("github"),
		CssCodeDark:  getCssCode("github-dark"),
		Title:        html.EscapeString(s.pageTitle(r.URL.Path)),
		Reload:       s.reloading(),
		BackHref:     html.EscapeString(backHref(r.URL.Path)),
		BackLabel:    html.EscapeString(backLabel(r.URL.Path)),
	})
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(buf.Bytes())
}

func (s *Server) serveFolder(w http.ResponseWriter, r *http.Request, root *Root) {
	dirPath := r.URL.Path
	if !strings.HasSuffix(dirPath, "/") {
		// Collapse leading slashes so the redirect can never be read as a
		// protocol-relative URL to another host.
		target := url.URL{Path: "/" + strings.TrimLeft(dirPath, "/") + "/", RawQuery: r.URL.RawQuery}
		http.Redirect(w, r, target.String(), http.StatusMovedPermanently)
		return
	}

	setNoCacheHeaders(w)
	stripCacheValidators(r)

	entries, err := root.List(dirPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		s.serverError(w, r, err)
		return
	}

	var readme htmltemplate.HTML
	if name, ok := root.Readme(dirPath); ok {
		content, err := root.ReadFile(path.Join(dirPath, name))
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		rendered, err := s.parser.MdToHTML(content)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		// Goldmark output is trusted, as it is for Markdown pages.
		readme = htmltemplate.HTML(rendered) //nolint:gosec
	}

	listing, err := renderListing(dirPath, entries, readme, time.Now())
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	title := defaultHTMLTitle
	if dirPath != "/" {
		title = s.pageTitle(strings.TrimSuffix(dirPath, "/"))
	}

	var buf bytes.Buffer
	err = renderTemplate(&buf, htmlStruct{
		Content:      listing,
		BoundingBox:  s.boundingBox,
		CssCodeLight: getCssCode("github"),
		CssCodeDark:  getCssCode("github-dark"),
		Title:        html.EscapeString(title),
		IsListing:    true,
		Reload:       s.reloading(),
	})
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(buf.Bytes())
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("❌ request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type htmlStruct struct {
	Content      string
	BoundingBox  bool
	CssCodeLight string
	CssCodeDark  string
	Title        string
	IsListing    bool
	Reload       bool
	// BackHref and BackLabel describe the link from a Markdown page to its
	// folder. Both are HTML-escaped; they are empty on folder pages.
	BackHref  string
	BackLabel string
}

func (s *Server) pageTitle(filename string) string {
	title := formatFilenameTitle(filename)
	if title == "" {
		return defaultHTMLTitle
	}
	return title
}

func formatFilenameTitle(filename string) string {
	filename = path.Base(filename)
	extension := path.Ext(filename)
	if strings.EqualFold(extension, ".md") {
		filename = strings.TrimSuffix(filename, extension)
	}

	filename = strings.Map(func(r rune) rune {
		if r == '-' || r == '_' {
			return ' '
		}
		return r
	}, filename)

	words := strings.Fields(filename)
	for i, word := range words {
		first, size := utf8.DecodeRuneInString(word)
		words[i] = string(unicode.ToUpper(first)) + word[size:]
	}
	return strings.Join(words, " ")
}

func renderTemplate(w io.Writer, html htmlStruct) error {
	tmpl, err := template.ParseFS(defaults.Templates, "templates/layout.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, html)
}

func getCssCode(style string) string {
	buf := new(strings.Builder)
	formatter := chroma_html.New(chroma_html.WithClasses(true))
	s := styles.Get(style)
	_ = formatter.WriteCSS(buf, s)
	return buf.String()
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func stripCacheValidators(r *http.Request) {
	r.Header.Del("If-Modified-Since")
	r.Header.Del("If-None-Match")
}
