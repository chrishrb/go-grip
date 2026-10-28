package internal

import (
	"bufio"
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestDirectoryListingIgnoresCacheValidators(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("expected Cache-Control to disable storage, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), "README.md") {
		t.Fatalf("expected directory listing body to mention README.md, got %q", recorder.Body.String())
	}
}

func TestRegularFileStillSupportsConditionalRequests(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "plain.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write plain.txt: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/plain.txt", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotModified {
		t.Fatalf("expected status %d, got %d", http.StatusNotModified, recorder.Code)
	}
}

func TestMarkdownResponsesDisableCaching(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/README.md", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("expected Cache-Control to disable storage, got %q", got)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html" {
		t.Fatalf("expected text/html response, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), "Hello") {
		t.Fatalf("expected rendered markdown response to contain document content, got %q", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "<title>README</title>") {
		t.Fatalf("expected filename-based HTML title, got %q", recorder.Body.String())
	}
}

func TestHandlerTraversalRejected(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, "secret.md"), []byte("top secret text\n"), 0o644); err != nil {
		t.Fatalf("write secret.md: %v", err)
	}
	rootDir := filepath.Join(parent, "root")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(rootDir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/%2e%2e/secret.md", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "top secret text") {
		t.Fatalf("response leaked a file outside the root: %q", recorder.Body.String())
	}
}

func TestStaticServesVendoredMathJaxFonts(t *testing.T) {
	t.Parallel()

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(t.TempDir()))

	tests := []struct {
		path        string
		contentType string
		contains    string
	}{
		{"/static/fonts/mathjax-newcm-font/chtml/woff2/mjx-ncm-zero.woff2", "font/woff2", "wOF2"},
		{"/static/fonts/mathjax-newcm-font/chtml/dynamic/double-struck.js", "text/javascript", "MathJaxNewcmFont"},
	}
	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", tt.path, http.StatusOK, recorder.Code)
		}
		if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, tt.contentType) {
			t.Fatalf("%s: expected Content-Type %q, got %q", tt.path, tt.contentType, got)
		}
		if !strings.Contains(recorder.Body.String(), tt.contains) {
			t.Fatalf("%s: expected body to contain %q", tt.path, tt.contains)
		}
	}

	// The embedded static tree must not expose anything outside it.
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/static/fonts/%2e%2e/%2e%2e/embed.go", nil))
	if recorder.Code == http.StatusOK || strings.Contains(recorder.Body.String(), "go:embed") {
		t.Fatalf("static path escaped the embedded tree: status %d", recorder.Code)
	}
}

func TestHandlerReadErrorReturns500(t *testing.T) {
	t.Parallel()

	if os.Getuid() == 0 {
		t.Skip("root can read files with mode 0o000")
	}

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte("# Fine\n"), 0o644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}
	locked := filepath.Join(tmpDir, "locked.md")
	if err := os.WriteFile(locked, []byte("# Locked\n"), 0o000); err != nil {
		t.Fatalf("write locked.md: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o644) })

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/locked.md", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if got := strings.TrimSpace(recorder.Body.String()); got != "Internal Server Error" {
		t.Fatalf("expected body %q, got %q", "Internal Server Error", got)
	}

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/a.md", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the handler to keep serving with status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestHandlerUppercaseMarkdownExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "B.MD"), []byte("# Upper Case\n"), 0o644); err != nil {
		t.Fatalf("write B.MD: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/B.MD", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html" {
		t.Fatalf("expected text/html response, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), "Upper Case</h1>") {
		t.Fatalf("expected rendered markdown, got %q", recorder.Body.String())
	}
}

func TestFormatFilenameTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{name: "humanizes separators", filename: "my-guide_v2.md", want: "My Guide V2"},
		{name: "uses basename", filename: "/docs/getting-started.md", want: "Getting Started"},
		{name: "preserves acronym", filename: "README.md", want: "README"},
		{name: "strips uppercase extension", filename: "release_notes.MD", want: "Release Notes"},
		{name: "collapses separators and whitespace", filename: "  release---notes__v2.md", want: "Release Notes V2"},
		{name: "supports unicode", filename: "überblick-plan.md", want: "Überblick Plan"},
		{name: "falls back for empty stem", filename: ".md", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatFilenameTitle(tt.filename); got != tt.want {
				t.Fatalf("formatFilenameTitle(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestFilenameTitleResponses(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	files := []string{"my-guide_v2.md", "unsafe-<script>.md", ".md"}
	for _, filename := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("# Hello\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", filename, err)
		}
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "humanized title", path: "/my-guide_v2.md", want: "<title>My Guide V2</title>"},
		{name: "escaped title", path: "/unsafe-%3Cscript%3E.md", want: "<title>Unsafe &lt;script&gt;</title>"},
		{name: "empty title fallback", path: "/.md", want: "<title>" + defaultHTMLTitle + "</title>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			if !strings.Contains(recorder.Body.String(), tt.want) {
				t.Fatalf("expected response to contain %q, got %q", tt.want, recorder.Body.String())
			}
		})
	}
}

func TestFolderPageRedirectsWithoutSlash(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	tests := []struct {
		path string
		want string
	}{
		{path: "/sub", want: "/sub/"},
		{path: "/sub?from=x", want: "/sub/?from=x"},
	}
	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if recorder.Code != http.StatusMovedPermanently {
			t.Fatalf("%s: expected status %d, got %d", tt.path, http.StatusMovedPermanently, recorder.Code)
		}
		if got := recorder.Header().Get("Location"); got != tt.want {
			t.Fatalf("%s: expected Location %q, got %q", tt.path, tt.want, got)
		}
	}
}

func TestFolderPageRendersTemplate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "sub", "inner"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"a.md", "sub/b.txt", "<img src=x onerror=alert(1)>.md"} {
		if err := os.WriteFile(filepath.Join(tmpDir, filepath.FromSlash(name)), []byte("x\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	tests := []struct {
		name    string
		path    string
		want    []string
		notWant []string
	}{
		{
			name: "root",
			path: "/",
			want: []string{
				"<title>" + defaultHTMLTitle + "</title>",
				`<link rel="stylesheet" href="/static/css/dir-listing.css" />`,
				`<script src="/static/js/dir-listing.js"></script>`,
				`<table class="grip-listing">`,
				`<a href="sub/">sub</a>`,
				`<a href="a.md">a.md</a>`,
				`&lt;img src=x onerror=alert(1)&gt;.md`,
			},
			notWant: []string{"grip-listing-up", "<img src=x"},
		},
		{
			name: "nested",
			path: "/sub/",
			want: []string{
				"<title>Sub</title>",
				`<a href="../?from=sub">..</a>`,
				`<a href="inner/">inner</a>`,
				`<a href="b.txt">b.txt</a>`,
			},
			notWant: []string{"a.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != "text/html" {
				t.Fatalf("expected text/html, got %q", got)
			}
			if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
				t.Fatalf("expected Cache-Control to disable storage, got %q", got)
			}
			body := recorder.Body.String()
			for _, w := range tt.want {
				if !strings.Contains(body, w) {
					t.Errorf("body missing %q\n%s", w, body)
				}
			}
			for _, nw := range tt.notWant {
				if strings.Contains(body, nw) {
					t.Errorf("body unexpectedly contains %q", nw)
				}
			}
		})
	}
}

func TestMarkdownPageDoesNotLoadListingAssets(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte("# A\n"), 0o644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/a.md", nil))
	if strings.Contains(recorder.Body.String(), "dir-listing") {
		t.Fatalf("markdown page must not load listing assets")
	}
}

func TestMarkdownPageLinksBackToFolder(t *testing.T) {
	t.Parallel()

	tmpDir := writeTree(t, map[string]string{
		"a.md":       "# A\n",
		"a/b.md":     "# B\n",
		"sub/c d.md": "# C\n",
		"x:y/p:q.md": "# P\n",
		"q&<x>/f.md": "# F\n",
	})

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	tests := []struct {
		name      string
		path      string
		wantHref  string
		wantLabel string
	}{
		{"root file", "/a.md", `href="/?from=a.md"`, `<span>/</span>`},
		{"nested file", "/a/b.md", `href="/a/?from=b.md"`, `<span>/a/</span>`},
		{"space in name", "/sub/c%20d.md", `href="/sub/?from=c+d.md"`, `<span>/sub/</span>`},
		{"colon in names", "/x%3Ay/p%3Aq.md", `href="/x%3Ay/?from=p%3Aq.md"`, `<span>/x:y/</span>`},
		{"html in folder name", "/q&%3Cx%3E/f.md", `href="/q&amp;%3Cx%3E/?from=f.md"`, `<span>/q&amp;&lt;x&gt;/</span>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			body := recorder.Body.String()
			for _, w := range []string{
				`<a id="grip-back" ` + tt.wantHref,
				tt.wantLabel,
				`<link rel="stylesheet" href="/static/css/back-to-folder.css" />`,
				`<script src="/static/js/back-to-folder.js"></script>`,
			} {
				if !strings.Contains(body, w) {
					t.Errorf("body missing %q\n%s", w, body)
				}
			}
			if strings.Contains(body, "<x>") {
				t.Errorf("folder name rendered unescaped")
			}
			nav := strings.Index(body, `<nav class="grip-back"`)
			content := strings.Index(body, "<h1")
			if nav < 0 || content < 0 || nav > content {
				t.Errorf("expected the back link above the content")
			}
		})
	}
}

func TestFolderPageHasNoBackLink(t *testing.T) {
	t.Parallel()

	tmpDir := writeTree(t, map[string]string{
		"sub/b.md":      "# B\n",
		"sub/README.md": "# Readme\n",
	})

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	for _, p := range []string{"/", "/sub/"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, p, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", p, http.StatusOK, recorder.Code)
		}
		if body := recorder.Body.String(); strings.Contains(body, "grip-back") || strings.Contains(body, "back-to-folder") {
			t.Errorf("%s: folder page must not render the back link or load its assets", p)
		}
	}
}

func TestFolderPageRendersReadme(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "docs", "readme.md"), []byte("# Docs Home\n"), 0o644); err != nil {
		t.Fatalf("write readme.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(tmpDir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/docs/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	table := strings.Index(body, "</table>")
	article := strings.Index(body, `<article class="grip-listing-readme markdown-body">`)
	if table < 0 || article < 0 || article < table {
		t.Fatalf("expected README article after the table, got %q", body)
	}
	if !strings.Contains(body[article:], "Docs Home</h1>") {
		t.Fatalf("expected rendered README, got %q", body[article:])
	}
}

// readdirErrorFS opens folders normally but fails to list those under
// /locked, as a folder without read permission would. A real chmod cannot
// produce this: on Linux an unreadable folder already fails to open.
type readdirErrorFS struct{ http.FileSystem }

type readdirErrorFile struct{ http.File }

func (f readdirErrorFile) Readdir(int) ([]fs.FileInfo, error) { return nil, fs.ErrPermission }

func (e readdirErrorFS) Open(name string) (http.File, error) {
	f, err := e.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(name, "/locked") {
		return readdirErrorFile{f}, nil
	}
	return f, nil
}

func TestFolderPageListErrorReturns500(t *testing.T) {
	t.Parallel()

	fsys := readdirErrorFS{http.FS(fstest.MapFS{
		"a.md":        {Data: []byte("# Fine\n")},
		"locked/b.md": {Data: []byte("# B\n")},
		"open/c.txt":  {Data: []byte("c\n")},
	})}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(fsys)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/locked/", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/open/", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the handler to keep serving with status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestFolderPageStaysInsideRoot(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, "outside-secret.md"), []byte("secret\n"), 0o644); err != nil {
		t.Fatalf("write outside-secret.md: %v", err)
	}
	rootDir := filepath.Join(parent, "root")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "inside.md"), []byte("# In\n"), 0o644); err != nil {
		t.Fatalf("write inside.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(rootDir))

	for _, p := range []string{"/%2e%2e/", "/%2e%2e", "/..%2f", "/sub/%2e%2e/%2e%2e/"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, p, nil))
		if strings.Contains(recorder.Body.String(), "outside-secret.md") {
			t.Fatalf("%s: folder page listed an entry outside the root (status %d)", p, recorder.Code)
		}
	}
}

func TestFolderPageListsSymlinkByNameOnly(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outside, "outside-secret.md"), []byte("secret\n"), 0o644); err != nil {
		t.Fatalf("write outside-secret.md: %v", err)
	}
	rootDir := filepath.Join(parent, "root")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(rootDir, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(rootDir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(body, ">link</a>") {
		t.Fatalf("expected the symlink entry in the root listing, got %q", body)
	}
	if strings.Contains(body, "outside-secret.md") {
		t.Fatalf("root listing must not include the symlink target's contents")
	}
}

// writeTree creates files (slash-separated paths) under a new temp folder.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", name, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// reloadServer returns a server with live reload on, backed by a hub on a
// fake event source.
func reloadServer(t *testing.T) (*Server, *fakeSource) {
	t.Helper()
	hub, src := runHub(t)
	server := NewServer("localhost", 6419, false, false, true, NewParser(), nil)
	server.hub = hub
	return server, src
}

func TestEventsStreamsReload(t *testing.T) {
	t.Parallel()

	dir := writeTree(t, map[string]string{"docs/a.md": "# A\n"})
	server, src := reloadServer(t)
	ts := httptest.NewServer(server.newHandler(http.Dir(dir)))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/__grip/events?path="+url.QueryEscape("/docs/a.md"), nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("expected text/event-stream, got %q", got)
	}

	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil || line != ": connected\n" {
		t.Fatalf("expected the connected comment first, got %q (%v)", line, err)
	}
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("read blank line: %v", err)
	}

	// An unrelated file in the same folder must not reload the page.
	src.events <- fsnotify.Event{Name: filepath.Join(dir, "docs", "b.md"), Op: fsnotify.Write}
	src.events <- fsnotify.Event{Name: filepath.Join(dir, "docs", "a.md"), Op: fsnotify.Write}

	got := make(chan string, 1)
	go func() {
		var buf strings.Builder
		for range 2 {
			l, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			buf.WriteString(l)
		}
		got <- buf.String()
	}()
	select {
	case s := <-got:
		if s != "event: reload\ndata: 1\n" {
			t.Fatalf("expected a reload event, got %q", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the reload event")
	}

	// Disconnecting releases the folder watch.
	cancel()
	deadline := time.Now().Add(2 * time.Second)
	for {
		_, removed := src.calls()
		if len(removed) == 1 && removed[0] == filepath.Join(dir, "docs") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected the docs folder watch to be released, removed %q", removed)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestEventsRejectsBadPath(t *testing.T) {
	t.Parallel()

	parent := writeTree(t, map[string]string{
		"x.md":           "outside\n",
		"root/a.md":      "# A\n",
		"root/notes.txt": "plain\n",
	})
	server, _ := reloadServer(t)
	handler := server.newHandler(http.Dir(filepath.Join(parent, "root")))

	tests := []struct {
		name  string
		query string
	}{
		{"encoded traversal", "?path=/%2e%2e/x.md"},
		{"plain traversal", "?path=/../x.md"},
		{"inner traversal", "?path=/sub/../a.md"},
		{"missing parameter", ""},
		{"relative path", "?path=a.md"},
		{"not found", "?path=/nope.md"},
		{"not markdown", "?path=/notes.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/__grip/events"+tt.query, nil))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
		})
	}
}

func TestEventsDisabledWithoutReload(t *testing.T) {
	t.Parallel()

	dir := writeTree(t, map[string]string{"a.md": "# A\n"})
	server := NewServer("localhost", 6419, false, false, false, NewParser(), nil)
	handler := server.newHandler(http.Dir(dir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/__grip/events?path=/a.md", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	for _, page := range []string{"/a.md", "/"} {
		recorder = httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, page, nil))
		if strings.Contains(recorder.Body.String(), "live-reload.js") {
			t.Fatalf("%s: reload script must not load with --no-reload", page)
		}
	}
}

func TestReloadPagesLoadScriptAndDisableCaching(t *testing.T) {
	t.Parallel()

	dir := writeTree(t, map[string]string{"a.md": "# A\n", "img.png": "png"})
	server, _ := reloadServer(t)
	handler := server.newHandler(http.Dir(dir))

	for _, page := range []string{"/a.md", "/"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, page, nil))
		if !strings.Contains(recorder.Body.String(), `<script src="/static/js/live-reload.js"></script>`) {
			t.Fatalf("%s: expected the reload script tag", page)
		}
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/img.png", nil))
	if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("expected Cache-Control no-cache on other files, got %q", got)
	}
}

func TestWatchSetFor(t *testing.T) {
	t.Parallel()

	parent := writeTree(t, map[string]string{
		"root/docs/a.md":      "# A\n\n![x](img/x.png) ![w](https://example.com/w.png)\n",
		"root/docs/README.md": "![r](r.png)\n",
		"root/docs/esc.md":    "![e](../../outside.png) ![ok](../top.png) ![abs](/abs.png)\n",
	})
	absRoot := filepath.Join(parent, "root")
	root := NewRoot(http.Dir(absRoot))
	server := NewServer("localhost", 6419, false, false, true, NewParser(), nil)
	abs := func(parts ...string) string { return filepath.Join(append([]string{absRoot}, parts...)...) }

	tests := []struct {
		name string
		path string
		want WatchSet
	}{
		{"markdown with image", "/docs/a.md", WatchSet{
			Files: []string{abs("docs", "a.md"), abs("docs", "img", "x.png")},
		}},
		{"folder with readme", "/docs/", WatchSet{
			Files: []string{abs("docs", "r.png")},
			Dirs:  []string{abs("docs")},
		}},
		{"ref escaping root is dropped", "/docs/esc.md", WatchSet{
			Files: []string{abs("docs", "esc.md"), abs("top.png"), abs("abs.png")},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := server.watchSetFor(root, absRoot, tt.path)
			if !ok {
				t.Fatalf("watchSetFor(%q) rejected the path", tt.path)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("watchSetFor(%q) = %+v, want %+v", tt.path, got, tt.want)
			}
		})
	}
}
