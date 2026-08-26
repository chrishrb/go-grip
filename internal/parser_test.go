package internal

import (
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
