package internal

import (
	"html/template"
	"strings"
	"testing"
	"time"
)

func TestRenderListing(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 9, 13, 0, 0, 0, time.UTC)
	mtime := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	entries := []Entry{
		{Name: "sub", IsDir: true, ModTime: mtime},
		{Name: "my notes.md", Size: 1536, ModTime: mtime},
		{Name: "<img src=x onerror=alert(1)>.md", Size: 3, ModTime: mtime},
	}

	tests := []struct {
		name    string
		dirPath string
		readme  template.HTML
		want    []string
		notWant []string
	}{
		{
			name:    "root has no up row",
			dirPath: "/",
			want:    []string{`<nav class="grip-listing-path">/</nav>`, `<table class="grip-listing">`},
			notWant: []string{"grip-listing-up", ">..<"},
		},
		{
			name:    "nested up row carries from",
			dirPath: "/a/b/",
			want:    []string{`<tr class="grip-listing-row grip-listing-up"><td class="grip-listing-name"><a href="../?from=b">..</a></td>`},
		},
		{
			name:    "folder href has slash",
			dirPath: "/",
			want: []string{
				`<tr class="grip-listing-row" data-dir="true">`,
				`<a href="sub/">sub</a>`,
				`<td class="grip-listing-mtime"><time datetime="2026-10-09T10:00:00Z" title="2026-10-09 10:00:00 UTC">3 hours ago</time></td>`,
			},
		},
		{
			name:    "space escaped",
			dirPath: "/",
			want:    []string{`href="my%20notes.md"`, `>my notes.md</a>`, `<td class="grip-listing-size">1.5 KB</td>`},
		},
		{
			name:    "hostile name escaped",
			dirPath: "/",
			want:    []string{`&lt;img src=x onerror=alert(1)&gt;.md</a>`},
			notWant: []string{"<img"},
		},
		{
			name:    "readme article present",
			dirPath: "/docs/",
			readme:  template.HTML("<h1>Docs</h1>"),
			want:    []string{`</table>`, `<article class="grip-listing-readme markdown-body"><h1>Docs</h1></article>`},
		},
		{
			name:    "readme absent",
			dirPath: "/docs/",
			notWant: []string{"<article"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := renderListing(tt.dirPath, entries, tt.readme, now)
			if err != nil {
				t.Fatalf("renderListing: %v", err)
			}
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("output missing %q\n%s", w, got)
				}
			}
			for _, nw := range tt.notWant {
				if strings.Contains(got, nw) {
					t.Errorf("output unexpectedly contains %q\n%s", nw, got)
				}
			}
			if tt.readme != "" && strings.Index(got, "</table>") > strings.Index(got, "<article") {
				t.Errorf("readme must follow the table\n%s", got)
			}
		})
	}
}

func TestRenderListingSchemeLikeName(t *testing.T) {
	t.Parallel()

	got, err := renderListing("/", []Entry{{Name: "javascript:alert(1).md"}}, "", time.Now())
	if err != nil {
		t.Fatalf("renderListing: %v", err)
	}
	if !strings.Contains(got, `href="javascript%3Aalert%281%29.md"`) {
		t.Fatalf("expected a relative, escaped href, got\n%s", got)
	}
}

func TestBackHrefAndLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		urlPath   string
		wantHref  string
		wantLabel string
	}{
		{"/a.md", "/?from=a.md", "/"},
		{"/a/b.md", "/a/?from=b.md", "/a/"},
		{"/sub/c d.md", "/sub/?from=c+d.md", "/sub/"},
		{"/x:y/p:q.md", "/x%3Ay/?from=p%3Aq.md", "/x:y/"},
		{"/a&b/c?d#e.md", "/a&b/?from=c%3Fd%23e.md", "/a&b/"},
		{"/a/b/c.md", "/a/b/?from=c.md", "/a/b/"},
		// A path can never point above the served root.
		{"/../../x.md", "/?from=x.md", "/"},
		{"//a//b.md", "/a/?from=b.md", "/a/"},
	}
	for _, tt := range tests {
		if got := backHref(tt.urlPath); got != tt.wantHref {
			t.Errorf("backHref(%q) = %q, want %q", tt.urlPath, got, tt.wantHref)
		}
		if got := backLabel(tt.urlPath); got != tt.wantLabel {
			t.Errorf("backLabel(%q) = %q, want %q", tt.urlPath, got, tt.wantLabel)
		}
	}
}

func TestFormatSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1536, "1.5 KB"},
		{5242880, "5.0 MB"},
	}
	for _, tt := range tests {
		if got := formatSize(tt.in); got != tt.want {
			t.Errorf("formatSize(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatRelativeTime(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		ago  time.Duration
		at   time.Time
		want string
	}{
		{name: "under a minute", ago: 30 * time.Second, want: "just now"},
		{name: "one minute", ago: time.Minute, want: "1 minute ago"},
		{name: "hours", ago: 5 * time.Hour, want: "5 hours ago"},
		{name: "days", ago: 3 * 24 * time.Hour, want: "3 days ago"},
		{name: "date after 30 days", at: time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC), want: "2026-01-02"},
		{name: "future clamps", ago: -time.Hour, want: "just now"},
	}
	for _, tt := range tests {
		at := tt.at
		if at.IsZero() {
			at = now.Add(-tt.ago)
		}
		if got := formatRelativeTime(at, now); got != tt.want {
			t.Errorf("%s: formatRelativeTime = %q, want %q", tt.name, got, tt.want)
		}
	}
}
