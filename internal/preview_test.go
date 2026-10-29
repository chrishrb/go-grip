package internal

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func getBody(t *testing.T, addr string) string {
	t.Helper()
	resp, err := http.Get(addr)
	if err != nil {
		t.Fatalf("GET %s: %v", addr, err)
	}
	defer resp.Body.Close() //nolint:errcheck
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", addr, resp.StatusCode)
	}
	return string(body)
}

func hostOf(t *testing.T, addr string) string {
	t.Helper()
	u, err := url.Parse(addr)
	if err != nil {
		t.Fatalf("parse %s: %v", addr, err)
	}
	return u.Host
}

func TestPreviewServesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "plan.md"), "# The Plan\n")

	p := NewPreviewer(false, false, NewParser(), nil)
	addr, err := p.Preview(filepath.Join(dir, "plan.md"), false)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if !strings.HasPrefix(addr, "http://localhost:") || !strings.HasSuffix(addr, "/plan.md") {
		t.Fatalf("unexpected URL %q", addr)
	}
	if body := getBody(t, addr); !strings.Contains(body, "The Plan") {
		t.Fatalf("expected rendered heading, got %q", body)
	}
}

func TestPreviewReusesServedRoots(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeFile(t, filepath.Join(base, "a", "one.md"), "# One\n")
	writeFile(t, filepath.Join(base, "a", "two.md"), "# Two\n")
	writeFile(t, filepath.Join(base, "a", "sub", "three.md"), "# Three\n")
	writeFile(t, filepath.Join(base, "b", "four.md"), "# Four\n")

	p := NewPreviewer(false, false, NewParser(), nil)
	preview := func(rel string) string {
		t.Helper()
		addr, err := p.Preview(filepath.Join(base, rel), false)
		if err != nil {
			t.Fatalf("Preview %s: %v", rel, err)
		}
		getBody(t, addr)
		return addr
	}

	one := preview("a/one.md")
	two := preview("a/two.md")
	three := preview("a/sub/three.md")
	four := preview("b/four.md")

	if hostOf(t, one) != hostOf(t, two) {
		t.Errorf("same folder should share a server: %s vs %s", one, two)
	}
	if hostOf(t, one) != hostOf(t, three) || !strings.HasSuffix(three, "/sub/three.md") {
		t.Errorf("subfolder should reuse the parent's server: %s vs %s", one, three)
	}
	if hostOf(t, one) == hostOf(t, four) {
		t.Errorf("sibling folder should get its own server: %s vs %s", one, four)
	}
}

func TestPreviewFolderKeepsTrailingSlash(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeFile(t, filepath.Join(base, "docs", "README.md"), "# Docs\n")

	p := NewPreviewer(false, false, NewParser(), nil)
	if _, err := p.Preview(base, false); err != nil {
		t.Fatalf("Preview root: %v", err)
	}
	addr, err := p.Preview(filepath.Join(base, "docs"), false)
	if err != nil {
		t.Fatalf("Preview docs: %v", err)
	}
	if !strings.HasSuffix(addr, "/docs/") {
		t.Fatalf("expected folder URL ending in /docs/, got %q", addr)
	}
	getBody(t, addr)
}

func TestPreviewEscapesNames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "my plan #1.md"), "# Escaped\n")

	p := NewPreviewer(false, false, NewParser(), nil)
	addr, err := p.Preview(filepath.Join(dir, "my plan #1.md"), false)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if body := getBody(t, addr); !strings.Contains(body, "Escaped") {
		t.Fatalf("expected rendered heading at %s, got %q", addr, body)
	}
}

func TestPreviewMissingPath(t *testing.T) {
	t.Parallel()

	p := NewPreviewer(false, false, NewParser(), nil)
	if _, err := p.Preview(filepath.Join(t.TempDir(), "nope.md"), false); err == nil {
		t.Fatal("expected an error for a missing path")
	}
	if len(p.roots) != 0 {
		t.Fatalf("expected no server to start, got %v", p.roots)
	}
}
