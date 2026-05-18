package internal

import (
	"strings"
	"testing"
)

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
