package internal

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ResolveTarget maps the CLI argument to the served directory (an OS path)
// and the URL path the browser opens first.
//
// An empty argument serves the working directory. A directory argument
// serves that directory and opens its root. A file argument serves the
// file's directory and opens the file. A path that does not exist is an
// error wrapping fs.ErrNotExist.
func ResolveTarget(arg string) (dir, startPath string, err error) {
	if arg == "" {
		return ".", "/", nil
	}

	cleaned := filepath.Clean(arg)
	info, err := os.Stat(cleaned)
	if err != nil {
		return "", "", err
	}

	if info.IsDir() {
		return cleaned, "/", nil
	}

	return filepath.Dir(cleaned), "/" + url.PathEscape(filepath.Base(cleaned)), nil
}

// Kind is the classification of a URL path inside a Root.
type Kind int

const (
	KindNotFound Kind = iota
	KindMarkdown
	KindDir
	KindOther
)

// Entry describes one item of a directory listing.
type Entry struct {
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// Root is the single point of disk access for the server. Every path is
// resolved through the wrapped http.FileSystem, so requests stay inside the
// served directory.
type Root struct {
	fsys http.FileSystem
}

// NewRoot wraps fsys, typically http.Dir(dir) in production and
// http.FS(fstest.MapFS{...}) in tests.
func NewRoot(fsys http.FileSystem) *Root {
	return &Root{fsys: fsys}
}

func isMarkdownName(name string) bool {
	return strings.EqualFold(path.Ext(name), ".md")
}

// Classify reports what urlPath refers to. Markdown is a regular file with
// a case-insensitive .md extension; a folder named x.md is KindDir.
func (r *Root) Classify(urlPath string) Kind {
	f, err := r.fsys.Open(urlPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return KindNotFound
		}
		// The path exists but cannot be opened (e.g. permission denied).
		// Markdown keeps its route so the read failure surfaces as a 500;
		// everything else is left to http.FileServer as before.
		if isMarkdownName(urlPath) {
			return KindMarkdown
		}
		return KindOther
	}
	//nolint:errcheck
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return KindNotFound
	}
	if info.IsDir() {
		return KindDir
	}
	if isMarkdownName(urlPath) {
		return KindMarkdown
	}
	return KindOther
}

// List returns every entry of the folder at urlPath, dotfiles included:
// folders first, then files, each ordered by case-insensitive name.
func (r *Root) List(urlPath string) ([]Entry, error) {
	f, err := r.fsys.Open(urlPath)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer f.Close()

	infos, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(infos))
	for _, info := range infos {
		entries = append(entries, Entry{
			Name:    info.Name(),
			IsDir:   info.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		li, lj := strings.ToLower(entries[i].Name), strings.ToLower(entries[j].Name)
		if li != lj {
			return li < lj
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

// ReadFile returns the contents of the file at urlPath.
func (r *Root) ReadFile(urlPath string) ([]byte, error) {
	f, err := r.fsys.Open(urlPath)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer f.Close()

	return io.ReadAll(f)
}

// Readme returns the name of the folder's README.md, matched in any letter
// case. An exact "README.md" wins over other spellings.
func (r *Root) Readme(dirPath string) (name string, ok bool) {
	entries, err := r.List(dirPath)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		if !e.IsDir && e.Name == "README.md" {
			return e.Name, true
		}
	}
	for _, e := range entries {
		if !e.IsDir && strings.EqualFold(e.Name, "README.md") {
			return e.Name, true
		}
	}
	return "", false
}
