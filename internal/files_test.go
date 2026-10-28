package internal

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestResolveTarget(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	for _, name := range []string{"docs/guide.md", "docs/my notes.md", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(tmpDir, filepath.FromSlash(name)), []byte("x\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	t.Chdir(tmpDir)

	tests := []struct {
		name      string
		arg       string
		wantDir   string
		wantStart string
	}{
		{name: "no argument", arg: "", wantDir: ".", wantStart: "/"},
		{name: "markdown file", arg: "docs/guide.md", wantDir: "docs", wantStart: "/guide.md"},
		{name: "directory", arg: "docs", wantDir: "docs", wantStart: "/"},
		{name: "directory trailing slash", arg: "docs/", wantDir: "docs", wantStart: "/"},
		{name: "non-markdown file", arg: "notes.txt", wantDir: ".", wantStart: "/notes.txt"},
		{name: "name with space is escaped", arg: "docs/my notes.md", wantDir: "docs", wantStart: "/my%20notes.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, start, err := ResolveTarget(tt.arg)
			if err != nil {
				t.Fatalf("ResolveTarget(%q) error: %v", tt.arg, err)
			}
			if dir != tt.wantDir || start != tt.wantStart {
				t.Fatalf("ResolveTarget(%q) = (%q, %q), want (%q, %q)", tt.arg, dir, start, tt.wantDir, tt.wantStart)
			}
		})
	}

	t.Run("missing path errors", func(t *testing.T) {
		_, _, err := ResolveTarget("missing.md")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("expected fs.ErrNotExist, got %v", err)
		}
		if !strings.Contains(err.Error(), "missing.md") {
			t.Fatalf("expected error to name missing.md, got %q", err.Error())
		}
	})
}

func TestRootClassify(t *testing.T) {
	t.Parallel()

	root := NewRoot(http.FS(fstest.MapFS{
		"a.md":         {Data: []byte("# a\n")},
		"B.MD":         {Data: []byte("# b\n")},
		"sub/c.md":     {Data: []byte("# c\n")},
		"x.png":        {Data: []byte("png")},
		"x.md/keep.md": {Data: []byte("# keep\n")},
	}))

	tests := []struct {
		name string
		path string
		want Kind
	}{
		{name: "markdown", path: "/a.md", want: KindMarkdown},
		{name: "uppercase markdown", path: "/B.MD", want: KindMarkdown},
		{name: "directory", path: "/sub", want: KindDir},
		{name: "other file", path: "/x.png", want: KindOther},
		{name: "missing", path: "/nope", want: KindNotFound},
		{name: "folder named like markdown", path: "/x.md", want: KindDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := root.Classify(tt.path); got != tt.want {
				t.Fatalf("Classify(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}

func TestRootListOrder(t *testing.T) {
	t.Parallel()

	root := NewRoot(http.FS(fstest.MapFS{
		"b.md":       {Data: []byte("b")},
		"A.md":       {Data: []byte("a")},
		".hidden":    {Data: []byte("h")},
		"zed/x.txt":  {Data: []byte("x")},
		"Alpha/y.md": {Data: []byte("y")},
	}))

	entries, err := root.List("/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	got := make([]string, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.Name)
	}
	want := []string{"Alpha", "zed", ".hidden", "A.md", "b.md"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("List order = %v, want %v", got, want)
	}
	if !entries[0].IsDir || !entries[1].IsDir || entries[2].IsDir {
		t.Fatalf("unexpected IsDir flags: %+v", entries)
	}
}

func TestRootListMissing(t *testing.T) {
	t.Parallel()

	root := NewRoot(http.FS(fstest.MapFS{"a.md": {Data: []byte("a")}}))
	if _, err := root.List("/nope"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist, got %v", err)
	}
}

func TestRootReadme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    fstest.MapFS
		wantName string
		wantOK   bool
	}{
		{
			name:     "exact",
			files:    fstest.MapFS{"README.md": {Data: []byte("r")}, "other.md": {Data: []byte("o")}},
			wantName: "README.md",
			wantOK:   true,
		},
		{
			name:     "lowercase",
			files:    fstest.MapFS{"readme.MD": {Data: []byte("r")}},
			wantName: "readme.MD",
			wantOK:   true,
		},
		{
			name:  "absent",
			files: fstest.MapFS{"other.md": {Data: []byte("o")}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			name, ok := NewRoot(http.FS(tt.files)).Readme("/")
			if name != tt.wantName || ok != tt.wantOK {
				t.Fatalf("Readme(\"/\") = (%q, %v), want (%q, %v)", name, ok, tt.wantName, tt.wantOK)
			}
		})
	}
}

func TestRootReadFileTraversal(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, "secret.md"), []byte("top secret\n"), 0o644); err != nil {
		t.Fatalf("write secret.md: %v", err)
	}
	rootDir := filepath.Join(parent, "root")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	root := NewRoot(http.Dir(rootDir))
	for _, p := range []string{"/../secret.md", "/../../etc/passwd"} {
		data, err := root.ReadFile(p)
		if err == nil {
			t.Fatalf("ReadFile(%q) read outside the root: %q", p, data)
		}
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("ReadFile(%q) expected fs.ErrNotExist, got %v", p, err)
		}
	}
}
