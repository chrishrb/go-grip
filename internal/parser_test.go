package internal

import (
	"reflect"
	"strings"
	"testing"
)

func TestMdToHTML_FrontmatterRenderedAsTable(t *testing.T) {
	p := NewParser()
	input := []byte("---\ntitle: Hello World\nauthor: Alice\n---\n\n# Body Heading\n")
	out, err := p.MdToHTML(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `class="frontmatter-table"`) {
		t.Errorf("expected frontmatter-table in output, got:\n%s", s)
	}
	if !strings.Contains(s, "<th>title</th>") {
		t.Errorf("expected title key in table, got:\n%s", s)
	}
	if !strings.Contains(s, "<td>Hello World</td>") {
		t.Errorf("expected title value in table, got:\n%s", s)
	}
	if !strings.Contains(s, "<h1") {
		t.Errorf("expected body heading rendered, got:\n%s", s)
	}
	// Table must appear before body content
	tablePos := strings.Index(s, `class="frontmatter-table"`)
	bodyPos := strings.Index(s, "<h1")
	if tablePos > bodyPos {
		t.Errorf("frontmatter table must appear before body content")
	}
}

func TestMdToHTML_NoFrontmatter_Unchanged(t *testing.T) {
	p := NewParser()
	plain := []byte("# Just a heading\n\nSome paragraph.\n")
	out, err := p.MdToHTML(plain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if strings.Contains(s, "frontmatter-table") {
		t.Errorf("should not inject table when no frontmatter present")
	}
	if !strings.Contains(s, "<h1") {
		t.Errorf("heading should be rendered normally")
	}
}

func TestMdToHTML_FrontmatterBodyNotInTable(t *testing.T) {
	p := NewParser()
	input := []byte("---\ntitle: Test\n---\n\nBody paragraph here.\n")
	out, err := p.MdToHTML(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	// Frontmatter keys must not bleed into the body
	if strings.Contains(s, "title: Test") && !strings.Contains(s, "<th>title</th>") {
		t.Errorf("frontmatter content leaked into body as text")
	}
	if !strings.Contains(s, "Body paragraph here.") {
		t.Errorf("body paragraph should be rendered")
	}
}

func TestMdToHTML_HeadingIDsMatchGitHub(t *testing.T) {
	p := NewParser()
	input := []byte("### Implementation `ATTACK_PATTERN`\n\n[link](#implementation-attack_pattern)\n")
	out, err := p.MdToHTML(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `id="implementation-attack_pattern"`) {
		t.Errorf("expected underscore preserved in heading id, got:\n%s", s)
	}
}

func TestMdToHTML_DuplicateHeadingIDs(t *testing.T) {
	p := NewParser()
	input := []byte("# Notes\n\n# Notes\n")
	out, err := p.MdToHTML(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `id="notes"`) || !strings.Contains(s, `id="notes-1"`) {
		t.Errorf("expected deduplicated heading ids, got:\n%s", s)
	}
}

func TestLocalRefs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"relative", "![a](img/x.png)\n", []string{"img/x.png"}},
		{"root relative", "![a](/abs.png)\n", []string{"/abs.png"}},
		{"escaped with query", "![a](img/my%20pic.png?v=2#f)\n", []string{"img/my pic.png"}},
		{"external skipped", "![a](https://x/y.png)\n", []string{}},
		{"protocol relative skipped", "![a](//cdn/y.png)\n", []string{}},
		{"data skipped", "![a](data:image/png;base64,AA)\n", []string{}},
		{"deduplicated", "![a](b.png) ![c](a.png)\n\n![d](b.png)\n", []string{"b.png", "a.png"}},
		{"raw html ignored", "<img src=\"z.png\">\n", []string{}},
		{"reference style", "![a][r]\n\n[r]: p.png\n", []string{"p.png"}},
		{"frontmatter skipped", "---\ntitle: x\n---\n\n![a](f.png)\n", []string{"f.png"}},
	}
	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.LocalRefs([]byte(tt.input))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("LocalRefs(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
