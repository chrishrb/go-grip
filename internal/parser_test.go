package internal

import (
	"reflect"
	"strings"
	"testing"
)

func TestMdToHTML_FrontmatterRenderedAsTable(t *testing.T) {
	p := NewParser(nil)
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
	p := NewParser(nil)
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
	p := NewParser(nil)
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
	p := NewParser(nil)
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
	p := NewParser(nil)
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
	p := NewParser(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.LocalRefs([]byte(tt.input))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("LocalRefs(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParserFeatureDisablePassThrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		disabledFeature string
		input           string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "mathjax",
			disabledFeature: FeatureMathJax,
			input:           "Inline math: $x + y$\n",
			wantContains:    []string{"$x + y$"},
			wantNotContains: []string{`\(`, `\)`},
		},
		{
			name:            "mermaid",
			disabledFeature: FeatureMermaid,
			input:           "```mermaid\ngraph TD;\n  A-->B;\n```\n",
			wantContains:    []string{`class="highlight`, `graph TD;`, `A--&gt;B;`},
			wantNotContains: []string{`<pre class="mermaid">`},
		},
		{
			name:            "details",
			disabledFeature: FeatureDetails,
			input:           "<details><summary>Open me</summary>Inside details</details>\n",
			wantContains:    []string{`<details><summary>Open me</summary>Inside details</details>`},
			wantNotContains: []string{`id="details-`, `sessionStorage`},
		},
		{
			name:            "footnote",
			disabledFeature: FeatureFootnote,
			input:           "Footnote ref[^1].\n\n[^1]: Footnote text.\n",
			wantContains:    []string{`Footnote ref[^1].`, `[^1]: Footnote text.`},
			wantNotContains: []string{`class="footnote-ref"`, `class="footnotes"`},
		},
		{
			name:            "ghissue",
			disabledFeature: FeatureGHIssue,
			input:           "Issue refs: #46 and grafana/grafana#22\n",
			wantContains:    []string{`Issue refs: #46 and grafana/grafana#22`},
			wantNotContains: []string{`class="issue-link"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser([]string{tt.disabledFeature})
			html, err := parser.MdToHTML([]byte(tt.input))
			if err != nil {
				t.Fatalf("MdToHTML returned error: %v", err)
			}

			got := string(html)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Fatalf("expected output to contain %q, got %q", want, got)
				}
			}
			for _, unwanted := range tt.wantNotContains {
				if strings.Contains(got, unwanted) {
					t.Fatalf("expected output not to contain %q, got %q", unwanted, got)
				}
			}
		})
	}
}

func TestParserMathEnabledTransformsInlineMath(t *testing.T) {
	t.Parallel()

	parser := NewParser(nil)
	html, err := parser.MdToHTML([]byte("Inline math: $x + y$\n"))
	if err != nil {
		t.Fatalf("MdToHTML returned error: %v", err)
	}

	got := string(html)
	if !strings.Contains(got, `\(`) || !strings.Contains(got, `\)`) {
		t.Fatalf("expected mathjax delimiters in output, got %q", got)
	}
	if strings.Contains(got, "$x + y$") {
		t.Fatalf("expected inline math to be transformed, got %q", got)
	}
}

func TestValidateMarkdownFeaturesRejectsUnknownValues(t *testing.T) {
	t.Parallel()

	err := ValidateMarkdownFeatures([]string{"mathjax", "bogus"})
	if err == nil {
		t.Fatal("expected validation error for unknown feature")
	}
}
